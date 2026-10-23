#!/usr/bin/env python3
"""E04 rules as invariants: a seeded random walk over 200 jobs through the public API.

Usage: python3 lab/repro/invariants04.py I04 [seed]
Each round leases a batch, then acks, nacks (retry or not), lets a lease lapse
or cancels at random, then advances the frozen lab clock so retries come due.
After every round it asserts the E04 rules (exploration/models/rules.md):
  I1 a job is never leased by two workers at once
  I2 attempts never exceed max_attempts; attempt n is 1..k, never reused
  I3 within a priority, leases are FIFO
  I4 stats equal a direct count
  I5 a retry comes after the failure and within the backoff cap
  I6 an error is a 4xx/5xx with the error envelope (stale ack, unknown route)
  I7 a cancelled job is never leased again
Prints `INVARIANTS HOLD: …` or `INVARIANT BROKEN: …` (exit 1).
"""
import collections
import datetime as dt
import random
import sys
import probe as p

SEED = int(sys.argv[2]) if len(sys.argv) > 2 else 4
JOBS, MAX_ATTEMPTS, CAP = 200, 3, 3600
rng = random.Random(SEED)
T0 = '2032-01-01T00:00:00Z'


def clock(action, **kwargs):
    p.ok('POST', '/admin/clock', dict(action=action, **kwargs), key='admin-key')


def broken(rule, detail):
    raise AssertionError(f'{rule} {detail}')


def walk():
    clock('set', at=T0)
    now = p.stamp(T0)
    p.queue(max_attempts=MAX_ATTEMPTS, lease_seconds=30, backoff={'kind': 'exponential', 'base_seconds': 2000})
    ids = []
    for start in range(0, JOBS, 100):
        batch = [{'kind': 'manual', 'payload': {'i': i}, 'priority': rng.randint(0, 9)} for i in range(start, start + 100)]
        ids += p.ok('POST', p.QPATH + '/jobs:batch', {'jobs': batch})['ids']
    priority = {}
    workers = [p.worker() for _ in range(3)]
    held, cancelled, steps = {}, set(), collections.Counter()
    for rnd in range(60):
        for w in workers:
            batch = p.ok('POST', p.QPATH + '/lease', {'worker_id': w, 'max_jobs': rng.randint(1, 20)})['jobs']
            keys = []
            for item in batch:
                j = item['id']
                if j in held:
                    broken('I1', f'job {j} leased by worker {w} while held by {held[j][0]}')
                if j in cancelled:
                    broken('I7', f'cancelled job {j} leased again')
                if item['attempt'] > MAX_ATTEMPTS:
                    broken('I2', f'job {j} leased for attempt {item["attempt"]}')
                if j not in priority:
                    priority[j] = p.job(j)['priority']
                keys.append((-priority[j], j))
                held[j] = (w, item['lease_token'])
            if keys != sorted(keys):
                broken('I3', f'lease order {keys}')
        for j, (w, token) in list(held.items()):
            action = rng.choice(['ack', 'ack', 'nack', 'nack', 'dead', 'lapse', 'cancel'])
            steps[action] += 1
            if action == 'lapse':
                continue
            del held[j]
            if action == 'ack':
                p.require(p.ack(j, token, {'ok': j})[0] == 200, ('ack', j))
                status, body = p.call('POST', f'/v1/jobs/{j}/ack', {'lease_token': token, 'result': None})
                if status != 409 or not isinstance(body, dict) or 'error' not in body:
                    broken('I6', f'second ack of job {j} → {status} {body}')
            elif action in ('nack', 'dead'):
                p.require(p.nack(j, token, retry=action == 'nack')[0] == 200, ('nack', j))
                v = p.job(j)
                if v['state'] == 'retry_wait':
                    wait = (p.stamp(v['run_at']) - now).total_seconds()
                    if not 0 <= wait <= CAP:
                        broken('I5', f'job {j} retry in {wait} s')
            else:
                p.ok('POST', f'/v1/jobs/{j}/cancel', {})
                cancelled.add(j)
        if rng.random() < .3:
            for j in rng.sample(ids, 3):
                if p.call('POST', f'/v1/jobs/{j}/cancel', {})[0] == 200:
                    cancelled.add(j)
                    held.pop(j, None)
        status, body = p.call('GET', '/v1/no-such-route')
        if status != 404 or not isinstance(body, dict) or 'error' not in body:
            broken('I6', f'unknown route → {status} {body}')
        clock('advance', seconds=CAP + 1)
        now += dt.timedelta(seconds=CAP + 1)
        # Lapsed leases have expired: they are no longer held once the sweep runs.
        p.wait(lambda: p.sql(f"SELECT count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{p.NAME}') AND (state='leased' OR (state IN ('retry_wait','scheduled') AND run_at <= '{now.isoformat()}'))") == '0', 8)
        held.clear()
        stats = p.ok('GET', p.QPATH + '/stats')
        direct = dict(line.split('|') for line in p.sql(f"SELECT state,count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{p.NAME}') GROUP BY state").splitlines())
        for state in ('queued', 'leased', 'scheduled', 'retry_wait', 'succeeded', 'dead', 'cancelled'):
            if int(stats[state]) != int(direct.get(state, 0)):
                broken('I4', f'round {rnd} stats {state}={stats[state]} direct={direct.get(state, 0)}')
        if int(direct.get('queued', 0)) == 0 and int(direct.get('retry_wait', 0)) == 0:
            break
    for j in ids:
        v = p.job(j)
        ns = [a['n'] for a in p.ok('GET', f'/v1/jobs/{j}/attempts')['attempts']]
        if v['attempts'] > MAX_ATTEMPTS or len(ns) > MAX_ATTEMPTS or ns != list(range(1, len(ns) + 1)):
            broken('I2', f'job {j} attempts={v["attempts"]} history n={ns}')
        if v['state'] not in ('succeeded', 'dead', 'cancelled'):
            broken('walk', f'job {j} left in {v["state"]} after {rnd + 1} rounds')
    return rnd + 1, steps


if __name__ == '__main__':
    try:
        rounds, steps = walk()
        print(f'INVARIANTS HOLD: I1–I7 over {JOBS} jobs, {rounds} rounds, seed {SEED}; actions {dict(steps)}')
    except AssertionError as e:
        print(f'INVARIANT BROKEN: {e} (seed {SEED})'); sys.exit(1)
    finally:
        clock('reset')
