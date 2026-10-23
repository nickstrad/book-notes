#!/usr/bin/env python3
"""E03 public oracles: request shape, numeric ranges and nested PATCH.
Raw bodies go over HTTP exactly as written; no fixture SQL and no source reading.
The same probes and real-time bounds run against current and historical binaries.
"""
import datetime as dt
import json
import random
import string
import sys
import time
import urllib.error
import urllib.request
import probe as p

ID = p.ID
WRAP = 9223372037  # one second past time.Duration's range


def raw(method, path, body, key='acme-key'):
    req = urllib.request.Request(p.BASE + path, body, {'Authorization': 'Bearer ' + key,
        'Content-Type': 'application/json'}, method=method)
    try:
        response = urllib.request.urlopen(req, timeout=15)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        text = response.read().decode(errors='replace')
        try: value = json.loads(text) if text else None
        except json.JSONDecodeError: value = text
        return response.status, value


def now():
    return dt.datetime.now(dt.timezone.utc)


def lease_one(queue=None):
    queue = queue or p.QPATH
    w = p.worker()
    batch = p.wait(lambda: p.ok('POST', queue + '/lease', {'worker_id': w, 'max_jobs': 1})['jobs'], 6)
    return batch[0]


def run():
    if ID == 'U11':
        p.queue()
        p.require(raw('POST', p.QPATH + '/jobs', b'{"kind":"manual","payload":{}}')[0] == 201, 'well-formed control')
        cases = {
            'trailing text': (p.QPATH + '/jobs', b'{"kind":"manual","payload":{}} trailing-garbage'),
            'second object': ('/v1/queues', json.dumps({'name': p.NAME + '-a'}).encode() + json.dumps({'name': p.NAME + '-b'}).encode()),
            'invalid UTF-8': (p.QPATH + '/jobs', b'{"kind":"man\xffual","payload":{}}'),
        }
        seen = {}
        for label, (path, body) in cases.items():
            status, value = raw('POST', path, body)
            p.require(status in (201, 400), (label, status, value))
            seen[label] = status
        return any(s == 201 for s in seen.values()), f'statuses {seen}; well-formed control 201'
    if ID == 'U12':
        p.queue()
        p.require(raw('POST', p.QPATH + '/jobs', b'{"kind":"manual","x":1}')[0] == 400, 'exact unknown-name control')
        seen = {}
        for label, body in {'KIND/PAYLOAD': b'{"KIND":"manual","PAYLOAD":{}}',
                            'priority then PRIORITY': b'{"kind":"manual","priority":1,"PRIORITY":9}'}.items():
            status, value = raw('POST', p.QPATH + '/jobs', body)
            p.require(status in (201, 400), (label, status, value))
            seen[label] = status if status == 400 else f"201 priority={p.job(value['id'])['priority']}"
        return any(s != 400 for s in seen.values()), f'{seen}; exact unknown name 400'
    if ID == 'U13':
        p.queue(lease_seconds=60)
        wrapped = {}
        start = now()
        status, value = p.call('POST', p.QPATH + '/jobs', {'kind': 'manual', 'delay_seconds': WRAP})
        p.require(status in (201, 400), ('enqueue', status, value))
        wrapped['enqueue'] = status == 201 and p.stamp(p.job(value['id'])['run_at']) < start
        for verb, body in [('heartbeat', {'extend_seconds': WRAP}), ('nack', {'error': 'x', 'delay_seconds': WRAP})]:
            p.enqueue()
            # Act on the job actually leased: a wrapped enqueue is due first.
            leased = lease_one()
            j, token = leased['id'], leased['lease_token']
            start = now()
            status, value = p.call('POST', f'/v1/jobs/{j}/{verb}', dict(lease_token=token, **body))
            p.require(status in (200, 400), (verb, status, value))
            if status == 200:
                v = p.job(j)
                field = 'lease_expires_at' if verb == 'heartbeat' else 'run_at'
                wrapped[verb] = v[field] is None or p.stamp(v[field]) < start
            else:
                wrapped[verb] = False
        return any(wrapped.values()), f'delay/extension {WRAP}s: wrapped into the past {wrapped}'
    if ID == 'U14':
        p.queue()
        p.require(p.call('POST', '/v1/queues', {'name': p.NAME})[0] == 409, 'duplicate-name control')
        name = ''.join(random.choices(string.ascii_letters + string.digits, k=2800))
        seen = {
            'create lease 2^31': p.call('POST', '/v1/queues', {'name': p.NAME + '-o', 'lease_seconds': 2147483648})[0],
            'PATCH lease 2^31': p.call('PATCH', p.QPATH, {'lease_seconds': 2147483648})[0],
            'name 2800 random': p.call('POST', '/v1/queues', {'name': name})[0],
        }
        p.require(all(s in (201, 400, 409, 500) for s in seen.values()), seen)
        return any(s in (409, 500) for s in seen.values()), f'statuses {seen}; duplicate name 409'
    if ID == 'U15':
        big = p.NAME + '-big'
        status, _ = p.call('POST', '/v1/queues', {'name': big, 'retention_hours': 1e8})
        p.require(status in (201, 400), ('create 1e8', status))
        if status == 400:
            p.ok('POST', '/v1/queues', {'name': big, 'retention_hours': 1000000})
        held = p.ok('POST', f'/v1/queues/{big}/jobs', {'kind': 'manual'})['id']
        p.ok('POST', f'/v1/jobs/{held}/cancel')
        p.queue(retention_hours=0)
        control = p.enqueue()
        p.ok('POST', f'/v1/jobs/{control}/cancel')
        end = time.monotonic() + 6
        while time.monotonic() < end and p.call('GET', f'/v1/jobs/{control}')[0] == 200:
            time.sleep(.2)
        kept = p.call('GET', f'/v1/jobs/{control}')[0] == 200
        health = p.call('GET', '/health')[0]
        p.ok('DELETE', f'/v1/jobs/{held}')  # restore the shared sweep on historical binaries
        return kept, f'create retention 1e8 → {status}; zero-retention control {control} still present after 6 s={kept}; health {health}'
    if ID == 'U16':
        p.queue(max_attempts=10, backoff={'kind': 'exponential', 'base_seconds': 2147483647})
        j = p.enqueue()
        w = p.worker()
        for _ in range(3):
            batch = p.wait(lambda: p.ok('POST', p.QPATH + '/lease', {'worker_id': w})['jobs'], 6)
            p.require(p.nack(j, batch[0]['lease_token'], delay_seconds=0)[0] == 200, 'zero-delay nack')
        batch = p.wait(lambda: p.ok('POST', p.QPATH + '/lease', {'worker_id': w})['jobs'], 6)
        start = now()
        p.require(p.nack(j, batch[0]['lease_token'])[0] == 200, 'default nack')
        v = p.job(j)
        return p.stamp(v['run_at']) < start, f'attempt 4 at base 2147483647 s: state {v["state"]}, run_at {v["run_at"]}'
    if ID == 'U17':
        negative = p.call('POST', '/v1/queues', {'name': p.NAME + '-neg', 'backoff': {'kind': 'exponential', 'base_seconds': 4, 'max_seconds': -7}})[0]
        p.require(negative in (201, 400), ('negative cap', negative))
        p.queue(backoff={'kind': 'exponential', 'base_seconds': 4, 'max_seconds': 5})
        j = p.enqueue()
        w = p.worker()
        batch = p.wait(lambda: p.ok('POST', p.QPATH + '/lease', {'worker_id': w})['jobs'], 6)
        p.nack(j, batch[0]['lease_token'], delay_seconds=0)
        batch = p.wait(lambda: p.ok('POST', p.QPATH + '/lease', {'worker_id': w})['jobs'], 6)
        start = now()
        p.nack(j, batch[0]['lease_token'])
        wait = (p.stamp(p.job(j)['run_at']) - start).total_seconds()
        return negative == 201 or wait > 6.5, f'max_seconds -7 → {negative}; attempt 2 at base 4, cap 5 waits {wait:.1f} s'
    if ID == 'U18':
        hook = 'http://hooksink:8201/hook'
        p.queue(webhook={'url': hook, 'secret': 'report-secret'})
        after_secret = p.ok('PATCH', p.QPATH, {'webhook': {'secret': 'rotated'}})
        after_url = p.ok('PATCH', p.QPATH, {'webhook': {'url': hook + '2'}})
        cleared = after_secret['webhook_url'] == '' or after_url['webhook_secret'] == ''
        return cleared, (f'secret-only PATCH → url {after_secret["webhook_url"]!r}; '
                         f'url-only PATCH → secret {after_url["webhook_secret"]!r}')
    raise RuntimeError('unknown id ' + ID)


if __name__ == '__main__':
    try:
        present, detail = run(); print(('PRESENT: ' if present else 'ABSENT: ') + ID + ' — ' + detail)
    except Exception as e:
        print('PROBE ERROR: ' + ID + ' — ' + repr(e), file=sys.stderr); sys.exit(2)
