#!/usr/bin/env python3
"""E04 public oracles: error envelopes, attempt history, delays, lease end, queued age, backoff cap.
Public API and the documented lab clock only; no fixture SQL and no source reading.
The same probes run against current and historical binaries.
"""
import datetime as dt
import json
import sys
import urllib.error
import urllib.request
import probe as p

ID = p.ID


def raw(method, path, key='acme-key'):
    req = urllib.request.Request(p.BASE + path, None, {'Authorization': 'Bearer ' + key}, method=method)
    try:
        response = urllib.request.urlopen(req, timeout=15)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        text = response.read().decode(errors='replace')
        return response.status, response.headers.get('Content-Type', ''), text


def envelope(text):
    try:
        value = json.loads(text)
    except json.JSONDecodeError:
        return False
    return isinstance(value, dict) and isinstance(value.get('error'), dict) and bool(value['error'].get('code'))


def clock(action, **kwargs):
    p.ok('POST', '/admin/clock', dict(action=action, **kwargs), key='admin-key')


def now():
    return dt.datetime.now(dt.timezone.utc)


def lease_job(j, w, queue=None):
    # Act on the job this probe enqueued, never on "the next job".
    batch = p.wait(lambda: p.ok('POST', (queue or p.QPATH) + '/lease', {'worker_id': w, 'max_jobs': 1})['jobs'], 6)
    p.require(batch[0]['id'] == j, ('leased another job', batch[0]['id'], j))
    return batch[0]


def run():
    if ID == 'U19':
        seen = {}
        for method, path in (('PUT', '/v1/queues'), ('GET', '/v1/nope'), ('PATCH', '/v1/jobs/1')):
            status, kind, text = raw(method, path)
            p.require(status in (404, 405), (method, path, status))
            seen[f'{method} {path}'] = (status, kind.split(';')[0], envelope(text))
        bad = [k for k, (_, kind, ok) in seen.items() if kind != 'application/json' or not ok]
        return bool(bad), '; '.join(f'{k} → {s} {t} envelope={e}' for k, (s, t, e) in seen.items())
    if ID == 'U20':
        p.queue(max_attempts=1)
        j = p.enqueue()
        w = p.worker()
        p.require(p.nack(j, lease_job(j, w)['lease_token'])[0] == 200, 'first nack')
        p.require(p.job(j)['state'] == 'dead', 'dead after one attempt')
        p.ok('POST', f'/v1/jobs/{j}/retry', {})
        p.require(p.ack(j, lease_job(j, w)['lease_token'], {'ok': True})[0] == 200, 'ack after retry')
        rows = p.ok('GET', f'/v1/jobs/{j}/attempts')['attempts']
        ns = [r['n'] for r in rows]
        outcomes = [r['outcome'] for r in rows]
        return ns != [1, 2] or outcomes != ['failed', 'succeeded'], f'attempts after dead → retry → ack: n={ns} outcomes={outcomes}'
    if ID == 'U21':
        p.queue(max_attempts=5)
        j = p.enqueue()
        w = p.worker()
        before = now()
        nack, _ = p.call('POST', f'/v1/jobs/{j}/nack', {'lease_token': lease_job(j, w)['lease_token'], 'error': 'x', 'delay_seconds': -5})
        p.require(nack in (200, 400), ('nack status', nack))
        early = ''
        if nack == 200:
            early = f', retry run_at {(p.stamp(p.job(j)["run_at"]) - before).total_seconds():+.1f} s from the failure'
        # Enqueue last: an accepted negative delay is due at once and would be leased first.
        enq, _ = p.call('POST', p.QPATH + '/jobs', {'kind': 'manual', 'payload': {}, 'delay_seconds': -5})
        p.require(enq in (201, 400), ('enqueue status', enq))
        return enq == 201 or nack == 200, f'delay_seconds -5: enqueue → {enq}, nack → {nack}{early}'
    if ID == 'U22':
        clock('set', at='2031-06-01T00:00:00Z')
        try:
            w = p.worker()
            statuses = []
            for i in range(3):
                # A fresh queue per round: the sweep may requeue the previous round's job.
                path = f'/v1/queues/{p.NAME}-{i}'
                p.ok('POST', '/v1/queues', {'name': f'{p.NAME}-{i}', 'lease_seconds': 30})
                j = p.ok('POST', path + '/jobs', {'kind': 'manual', 'payload': {}})['id']
                token = lease_job(j, w, path)['lease_token']
                clock('advance', seconds=31)
                statuses.append(p.call('POST', f'/v1/jobs/{j}/heartbeat', {'lease_token': token, 'extend_seconds': 30})[0])
            p.require(all(s in (200, 409) for s in statuses), ('heartbeat status', statuses))
            return 200 in statuses, f'heartbeat 1 s after lease_expires_at (before or after the sweep): {statuses}'
        finally:
            clock('reset')
    if ID == 'U23':
        clock('set', at='2031-07-01T00:00:00Z')
        try:
            p.queue(lease_seconds=30)
            j = p.enqueue()
            lease_job(j, p.worker())
            clock('advance', seconds=600)
            p.wait(lambda: p.job(j)['state'] == 'queued', 6)
            expired = p.ok('GET', p.QPATH + '/stats')['oldest_queued_age']
            other = '/v1/queues/' + p.NAME + '-w'
            p.ok('POST', '/v1/queues', {'name': p.NAME + '-w', 'lease_seconds': 3600})
            k = p.ok('POST', other + '/jobs', {'kind': 'manual', 'payload': {}})['id']
            w = p.worker()
            lease_job(k, w, other)
            clock('advance', seconds=100)
            p.checked('DELETE', f'/v1/workers/{w}')
            p.require(p.job(k)['state'] == 'queued', 'released by worker deletion')
            deleted = p.ok('GET', other + '/stats')['oldest_queued_age']
            return expired > 30 or deleted > 30, f'oldest_queued_age just after release: expiry {expired:.0f} s, worker deletion {deleted:.0f} s'
        finally:
            clock('reset')
    if ID == 'A1':
        p.queue(max_attempts=5, backoff={'kind': 'exponential', 'base_seconds': 3000})
        j = p.enqueue()
        w = p.worker()
        p.require(p.nack(j, lease_job(j, w)['lease_token'], delay_seconds=0)[0] == 200, 'zero-delay nack')
        token = lease_job(j, w)['lease_token']
        start = now()
        p.require(p.nack(j, token)[0] == 200, 'default nack')
        wait = (p.stamp(p.job(j)['run_at']) - start).total_seconds()
        return wait > 3660, f'attempt 2 at base 3000 s, no max_seconds: waits {wait:.0f} s'
    raise RuntimeError('unknown id ' + ID)


if __name__ == '__main__':
    try:
        present, detail = run(); print(('PRESENT: ' if present else 'ABSENT: ') + ID + ' — ' + detail)
    except Exception as e:
        print('PROBE ERROR: ' + ID + ' — ' + repr(e), file=sys.stderr); sys.exit(2)
