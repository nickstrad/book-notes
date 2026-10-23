#!/usr/bin/env python3
"""E05 sequences as invariants: a seeded random walk of 100 verbs through the public API.

Usage: python3 lab/repro/walk05.py W05 [seed]
Each step picks a verb at random (enqueue, enqueue with a delay, new worker,
delete a worker, lease, ack, nack, nack without retry, heartbeat, cancel, retry,
purge, pause, resume, advance the frozen lab clock), acts, and asserts after
every step:
  W1 no job is leased more than max_attempts times since its last retry
  W2 a job has a lease token, holder and expiry exactly when it is leased
  W3 every lease_expired event matches an expired attempt
  W4 a job leased when its queue was purged never becomes pending again
  W5 a cancelled job is never leased again
  W6 attempt n is 1..k, never reused
  W7 a verb the walker's model allows is not refused
Read-only SQL is used for W2, W3 and W6. Prints `WALK HOLDS: …` or
`WALK BROKEN: …` (exit 1).
"""
import collections
import random
import sys
import probe as p

SEED = int(sys.argv[2]) if len(sys.argv) > 2 else 5
STEPS, MAX_ATTEMPTS, LEASE = 100, 2, 30
rng = random.Random(SEED)
VERBS = ['enqueue'] * 2 + ['enqueue_delay'] + ['worker'] * 2 + ['delete_worker'] + ['lease'] * 6 + \
    ['cancel', 'retry', 'purge', 'pause'] + ['resume'] * 2 + ['advance'] * 2
HOLDER = ['ack'] * 3 + ['nack'] * 2 + ['nack_dead', 'heartbeat', 'cancel', 'purge'] + ['delete_worker'] * 2 + ['advance'] * 2


def clock(action, **kwargs):
    return p.ok('POST', '/admin/clock', dict(action=action, **kwargs), key='admin-key')


def broken(rule, step, verb, detail):
    raise AssertionError(f'{rule} at step {step} ({verb}): {detail}')


class Walk:
    def __init__(self):
        self.now = p.stamp(clock('set', at='2032-02-01T00:00:00Z')['now'])
        p.queue(max_attempts=MAX_ATTEMPTS, lease_seconds=LEASE, backoff={'kind': 'exponential', 'base_seconds': 10})
        self.qid = p.sql(f"SELECT id FROM queues WHERE name='{p.NAME}'")
        self.jobs, self.workers = [], [p.worker()]
        self.held = {}  # job -> (worker, token, expiry)
        self.grants = collections.Counter()
        self.cancelled, self.purged = set(), set()
        self.paused = False
        self.verbs = collections.Counter()

    def allowed(self, step, verb, status, what):
        if status != 200:
            broken('W7', step, verb, f'{what} → {status}')

    def act(self, step, verb):
        if verb in ('enqueue', 'enqueue_delay'):
            body = {'kind': 'manual', 'payload': {'step': step}}
            if verb == 'enqueue_delay':
                body['delay_seconds'] = rng.randint(20, 60)
            self.jobs.append(p.ok('POST', p.QPATH + '/jobs', body)['id'])
        elif verb == 'worker':
            self.workers.append(p.worker())
        elif verb == 'delete_worker' and self.workers:
            w = self.workers.pop(rng.randrange(len(self.workers)))
            p.checked('DELETE', f'/v1/workers/{w}')
            for j in [j for j, h in self.held.items() if h[0] == w]:
                del self.held[j]
        elif verb == 'lease' and self.workers:
            w = rng.choice(self.workers)
            for item in p.ok('POST', p.QPATH + '/lease', {'worker_id': w, 'max_jobs': rng.randint(1, 3)})['jobs']:
                j = item['id']
                if j in self.cancelled:
                    broken('W5', step, verb, f'cancelled job {j} leased again')
                self.grants[j] += 1
                if self.grants[j] > MAX_ATTEMPTS:
                    broken('W1', step, verb, f'job {j} leased {self.grants[j]} times, attempt {item["attempt"]}')
                self.held[j] = (w, item['lease_token'], p.stamp(item['lease_expires_at']))
        elif verb in ('ack', 'nack', 'nack_dead', 'heartbeat') and self.held:
            j = rng.choice(sorted(self.held))
            w, token, expiry = self.held[j]
            if verb == 'ack':
                self.allowed(step, verb, p.ack(j, token, {'step': step})[0], f'ack {j}')
            elif verb == 'heartbeat':
                status, _ = p.call('POST', f'/v1/jobs/{j}/heartbeat', {'lease_token': token, 'extend_seconds': LEASE})
                self.allowed(step, verb, status, f'heartbeat {j}')
                self.held[j] = (w, token, self.now + p.dt.timedelta(seconds=LEASE))
                return
            else:
                self.allowed(step, verb, p.nack(j, token, retry=verb == 'nack')[0], f'{verb} {j}')
            del self.held[j]
        elif verb == 'cancel' and self.jobs:
            j = rng.choice(self.jobs)
            if p.call('POST', f'/v1/jobs/{j}/cancel', {})[0] == 200:
                self.cancelled.add(j)
                self.held.pop(j, None)
        elif verb == 'retry' and self.jobs:
            j = rng.choice(self.jobs)
            if p.call('POST', f'/v1/jobs/{j}/retry', {})[0] == 200:
                self.grants[j] = 0
                self.purged.discard(j)
        elif verb == 'purge':
            self.purged |= set(self.held)
            p.ok('POST', p.QPATH + '/purge', {})
        elif verb in ('pause', 'resume') and self.paused != (verb == 'pause'):
            p.ok('POST', p.QPATH + '/' + verb, {})
            self.paused = verb == 'pause'
        elif verb == 'advance':
            seconds = rng.choice([rng.randint(5, 25), LEASE + 1])
            self.now = p.stamp(clock('advance', seconds=seconds)['now'])
            self.held = {j: h for j, h in self.held.items() if h[2] > self.now}
            # Let the loop sweep what the step made due before judging it.
            p.wait(lambda: p.sql(f"SELECT count(*) FROM jobs WHERE queue_id={self.qid} AND ((state='leased' AND lease_expires_at <= '{self.now.isoformat()}') OR (state IN ('scheduled','retry_wait') AND run_at <= '{self.now.isoformat()}'))") == '0', 8)
        else:
            return False
        return True

    def check(self, step, verb):
        rows = p.sql(f"SELECT id,state,lease_token IS NOT NULL,leased_by IS NOT NULL,lease_expires_at IS NOT NULL FROM jobs WHERE queue_id={self.qid}").splitlines()
        states = {}
        for row in rows:
            j, state, *lease = row.split('|')
            states[int(j)] = state
            if (state == 'leased') != (lease == ['t', 't', 't']) or len(set(lease)) != 1:
                broken('W2', step, verb, f'job {j} {state} with token/holder/expiry {lease}')
        for j in self.purged:
            if states.get(j) in ('queued', 'scheduled', 'retry_wait'):
                broken('W4', step, verb, f'job {j}, leased at a purge, is {states[j]} again')
        mismatch = p.sql(f"SELECT j.id FROM jobs j WHERE j.queue_id={self.qid} AND (SELECT count(*) FROM job_events e WHERE e.job_id=j.id AND e.event='lease_expired') <> (SELECT count(*) FROM attempts a WHERE a.job_id=j.id AND a.outcome='expired')")
        if mismatch:
            broken('W3', step, verb, f'lease_expired events without an expiry on jobs {mismatch.split()}')
        gaps = p.sql(f"SELECT job_id FROM attempts WHERE job_id IN (SELECT id FROM jobs WHERE queue_id={self.qid}) GROUP BY job_id HAVING min(n) <> 1 OR max(n) <> count(*) OR count(DISTINCT n) <> count(*)")
        if gaps:
            broken('W6', step, verb, f'attempt numbers reused or skipped on jobs {gaps.split()}')

    def run(self):
        step = 0
        while step < STEPS:
            # Holder verbs need a lease in hand; favour them while one is held.
            verb = rng.choice(HOLDER if self.held and rng.random() < .5 else VERBS)
            if not self.act(step + 1, verb):
                continue
            step += 1
            self.verbs[verb] += 1
            self.check(step, verb)
        return step


if __name__ == '__main__':
    try:
        walk = Walk()
        steps = walk.run()
        print(f'WALK HOLDS: {steps}/{STEPS} steps, W1–W7 after each, seed {SEED}; verbs {dict(sorted(walk.verbs.items()))}')
    except AssertionError as e:
        print(f'WALK BROKEN: {e} (seed {SEED})'); sys.exit(1)
    finally:
        clock('reset')
