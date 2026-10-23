#!/usr/bin/env python3
"""E05 public oracles: verbs in unusual orders.
Public API, the documented lab clock and the event history only; no fixture SQL.
The same probes run against current and historical binaries.
"""
import sys
import probe as p
from probe04 import clock, lease_job

ID = p.ID


def lease_mine(j, w):
    # One lease call: an empty batch is an answer here, not something to wait out.
    batch = p.ok('POST', p.QPATH + '/lease', {'worker_id': w, 'max_jobs': 1})['jobs']
    p.require(not batch or batch[0]['id'] == j, ('leased another job', batch[0]['id'] if batch else None, j))
    return batch[0] if batch else None


def held(v):
    return {k: v[k] for k in ('lease_token', 'leased_by', 'lease_expires_at') if v.get(k)}


def last_event(j):
    return p.ok('GET', f'/v1/jobs/{j}/events')['events'][-1]['event']


def run():
    if ID == 'U24':
        p.queue(max_attempts=2)
        j = p.enqueue()
        attempts = []
        for _ in range(4):
            w = p.worker()
            item = lease_mine(j, w)
            if item is None:
                break
            attempts.append(item['attempt'])
            p.checked('DELETE', f'/v1/workers/{w}')
        rows = [a['outcome'] for a in p.ok('GET', f'/v1/jobs/{j}/attempts')['attempts']]
        state = p.job(j)['state']
        return len(attempts) > 2 or attempts != sorted(set(attempts)), \
            f'max_attempts 2, lease then delete the worker ×4: {len(attempts)} deliveries, attempt {attempts}, history {rows}, now {state}'
    if ID == 'U25':
        p.queue()
        w = p.worker()
        a, c = p.enqueue(), p.enqueue()
        p.require(p.ack(a, lease_job(a, w)['lease_token'], {'ok': True})[0] == 200, 'ack')
        lease_job(c, w)
        p.ok('POST', f'/v1/jobs/{c}/cancel', {})
        left = {f'{p.job(j)["state"]} {j}': held(p.job(j)) for j in (a, c)}
        return any(left.values()), 'lease fields on terminal rows: ' + '; '.join(f'{k} {sorted(v) or "none"}' for k, v in left.items())
    if ID == 'U26':
        clock('set', at='2031-08-01T00:00:00Z')
        try:
            p.queue(lease_seconds=30, max_attempts=5)
            w = p.worker()
            a, c = p.enqueue(), p.enqueue()
            batch = p.ok('POST', p.QPATH + '/lease', {'worker_id': w, 'max_jobs': 2})['jobs']
            p.require(sorted(i['id'] for i in batch) == [a, c], ('first lease', batch))
            clock('advance', seconds=31)
            p.wait(lambda: all(p.job(j)['state'] == 'queued' for j in (a, c)), 6)
            tokens = {i['id']: i['lease_token'] for i in p.ok('POST', p.QPATH + '/lease', {'worker_id': w, 'max_jobs': 2})['jobs']}
            p.require(sorted(tokens) == [a, c], ('second lease', tokens))
            p.require(p.ack(a, tokens[a], {'ok': True})[0] == 200, 'ack after one expiry')
            p.ok('POST', f'/v1/jobs/{c}/cancel', {})
            seen = {'ack': last_event(a), 'cancel': last_event(c)}
            return seen != {'ack': 'succeeded', 'cancel': 'cancelled'}, f'after one lease expiry, ack logged {seen["ack"]}, cancel logged {seen["cancel"]}'
        finally:
            clock('reset')
    if ID == 'A2':
        p.queue(max_attempts=5)
        j = p.enqueue()
        token = lease_job(j, p.worker())['lease_token']
        deleted = p.ok('POST', p.QPATH + '/purge', {})['deleted']
        p.require(p.nack(j, token)[0] == 200, 'nack after purge')
        state = p.job(j)['state']
        return state in ('queued', 'scheduled', 'retry_wait'), f'lease → purge (deleted {deleted}) → nack leaves {state}'
    raise RuntimeError('unknown id ' + ID)


if __name__ == '__main__':
    try:
        present, detail = run(); print(('PRESENT: ' if present else 'ABSENT: ') + ID + ' — ' + detail)
    except Exception as e:
        print('PROBE ERROR: ' + ID + ' — ' + repr(e), file=sys.stderr); sys.exit(2)
