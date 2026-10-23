#!/usr/bin/env python3
"""E09 predictions: schedules a pass cannot get past, cron text the spec does not name,
and terminal transitions whose notification is written after they commit.
Only public routes, the lab clock and read-only SQL are used. Every unexpected
outcome or missing control is a PROBE ERROR, never ABSENT. Run serially: the
probes freeze the lab clock and always reset it.
"""
import concurrent.futures as futures
import json
import sys
import threading
import time
import probe as p
from probe04 import clock

HOOK = {'url': 'http://hooksink:8201/hook', 'secret': 's'}


def evidence(label, **values):
    print(json.dumps(dict(probe=p.ID, label=label, **values), sort_keys=True, default=str), file=sys.stderr, flush=True)


def queue(suffix, key='acme-key', **options):
    name = p.NAME + '-' + suffix
    p.ok('POST', '/v1/queues', dict(name=name, **options), key=key)
    return name


def schedule(name, key='acme-key', **fields):
    body = dict({'name': p.NAME, 'queue': name, 'kind': 'manual', 'payload': {}, 'cron': '0 9 * * *',
        'timezone': 'UTC'}, **fields)
    return p.call('POST', '/v1/schedules', body, key=key)


def created(status, value, what):
    p.require(status in (201, 400), (what, status, value))
    if status == 400:
        p.require(value['error']['code'] in ('invalid_cron', 'invalid_timezone'), (what, value))
    return status == 201


def remove(ids, key='acme-key'):
    for sid in ids:
        p.checked('DELETE', f'/v1/schedules/{sid}', statuses=(204, 404), key=key)


def unadvanceable():
    # A never-matching schedule (lower id, acme) ahead of another tenant's every-minute schedule.
    acme, globex, ids = queue('a'), queue('g', key='globex-key'), {'acme-key': [], 'globex-key': []}
    try:
        clock('set', at='2033-04-01T00:00:30Z')
        status, bad = schedule(acme, cron='0 0 30 2 *')
        accepted = created(status, bad, 'unadvanceable create')
        if accepted:
            ids['acme-key'].append(bad['id'])
        else:
            # The same schedule reached by PATCH from a valid one.
            status, good = schedule(acme)
            p.require(status == 201, ('valid control schedule', status, good))
            ids['acme-key'].append(good['id'])
            status, value = p.call('PATCH', f'/v1/schedules/{good["id"]}', {'cron': '0 0 30 2 *'})
            p.require(status in (200, 400), ('unadvanceable patch', status, value))
            accepted = status == 200
        status, victim = schedule(globex, key='globex-key', cron='* * * * *')
        p.require(status == 201, ('victim schedule', status, victim))
        ids['globex-key'].append(victim['id'])
        jobs = []
        for _ in range(2):
            clock('advance', seconds=60)
            time.sleep(3)
            jobs = p.ok('GET', f'/v1/queues/{globex}/jobs', key='globex-key')['jobs']
            if jobs:
                break
        evidence('victim', unadvanceable_accepted=accepted, jobs=len(jobs))
        # Without the bad schedule the victim must fire: otherwise the probe proves nothing.
        p.require(jobs or accepted, ('victim did not fire with no unadvanceable schedule stored', victim))
        return not jobs, f'unadvanceable schedule stored={accepted}; later globex every-minute schedule fired jobs={len(jobs)} over two minute ticks'
    finally:
        for key, sids in ids.items():
            remove(sids, key)
        clock('reset')


def never_matches():
    name, ids, accepted = queue('n'), [], {}
    try:
        for cron in ('0 0 30 2 *', '* * 30 2 *'):
            status, value = schedule(name, cron=cron)
            accepted['create ' + cron] = created(status, value, cron)
            if status == 201:
                ids.append(value['id'])
                p.require(value['next_run_at'], ('stored without a next fire', value))
                accepted['create ' + cron] = value['next_run_at']
        status, good = schedule(name)
        p.require(status == 201, ('valid control', status, good))
        ids.append(good['id'])
        status, value = p.call('PATCH', f'/v1/schedules/{good["id"]}', {'cron': '0 0 31 4 *'})
        p.require(status in (200, 400), ('patch', status, value))
        accepted['patch 0 0 31 4 *'] = value['next_run_at'] if status == 200 else False
        # A schedule may always be disabled.
        p.require(p.call('PATCH', f'/v1/schedules/{good["id"]}', {'enabled': False})[0] == 200, 'disable control')
        return any(accepted.values()), f'accepted={accepted}; valid control 201'
    finally:
        remove(ids)


def zone_prefix():
    name, ids, accepted = queue('z'), [], {}
    try:
        clock('set', at='2033-04-01T00:00:30Z')
        for cron, zone in (('CRON_TZ=Asia/Tokyo 0 9 * * *', 'UTC'), ('TZ=Asia/Tokyo 0 9 * * *', 'UTC'),
                ('CRON_TZ=Asia/Tokyo 0 9 * * *', 'America/New_York')):
            status, value = schedule(name, cron=cron, timezone=zone)
            ok = created(status, value, cron)
            if ok:
                ids.append(value['id'])
            accepted[f'{cron} in {zone}'] = value['next_run_at'] if ok else False
        status, good = schedule(name)
        p.require(status == 201 and good['next_run_at'] == '2033-04-01T09:00:00Z', ('valid control', status, good))
        ids.append(good['id'])
        return any(accepted.values()), f'accepted={accepted}; plain 0 9 * * * UTC next=09:00Z'
    finally:
        remove(ids)
        clock('reset')


def frozen_offset():
    name, ids = queue('o'), []
    try:
        clock('set', at='2026-11-01T05:30:00Z')  # 01:30 EDT
        status, value = schedule(name, cron='0 * * * *', timezone='America/New_York')
        p.require(status == 201 and value['zone_offset'] == -14400, ('EDT control', status, value))
        ids.append(value['id'])
        clock('set', at='2026-11-01T07:10:00Z')  # 02:10 EST
        got = p.ok('GET', f'/v1/schedules/{value["id"]}')['zone_offset']
        listed = [s['zone_offset'] for s in p.ok('GET', '/v1/schedules')['schedules'] if s['id'] == value['id']]
        p.require(len(listed) == 1, listed)
        return got != -18000 or listed[0] != -18000, f'created at EDT offset -14400; at 07:10Z (EST, -18000) GET zone_offset={got}, list={listed[0]}'
    finally:
        remove(ids)
        clock('reset')


def local_zone():
    name, ids, accepted = queue('l'), [], {}
    try:
        status, value = schedule(name, timezone='Local')
        accepted['create Local'] = created(status, value, 'Local')
        if status == 201:
            ids.append(value['id'])
        status, good = schedule(name, timezone='America/New_York')
        p.require(status == 201, ('IANA control', status, good))
        ids.append(good['id'])
        for zone in ('Local', ''):
            status, value = p.call('PATCH', f'/v1/schedules/{good["id"]}', {'timezone': zone})
            p.require(status in (200, 400), ('patch', zone, status, value))
            accepted[f'patch {zone!r}'] = status == 200
            if status == 200:
                p.ok('PATCH', f'/v1/schedules/{good["id"]}', {'timezone': 'America/New_York'})
        return any(accepted.values()), f'accepted={accepted}; America/New_York control 201'
    finally:
        remove(ids)


def spin_delete(j, stop, out):
    # A second client of the same tenant deletes the job the moment it is terminal.
    while not stop.is_set():
        status, _ = p.call('DELETE', f'/v1/jobs/{j}')
        if status in (204, 404):
            out[j] = status
            return
        p.require(status == 409, ('delete', j, status))


def notified(ids):
    got = p.sql(f"SELECT COALESCE(string_agg(DISTINCT job_id::text, ','), '') FROM webhook_outbox WHERE job_id IN ({','.join(map(str, ids))})")
    return {int(x) for x in got.split(',') if x}


def committed(ids, event):
    got = p.sql(f"SELECT COALESCE(string_agg(DISTINCT job_id::text, ','), '') FROM job_events WHERE event='{event}' AND job_id IN ({','.join(map(str, ids))})")
    return {int(x) for x in got.split(',') if x}


def after_commit(trials=150):
    # Ack and cancel, each racing a delete; every committed transition must answer 2xx and be notified.
    name = queue('t', webhook=HOOK, lease_seconds=300)
    w = p.ok('POST', '/v1/workers', {'name': p.NAME, 'kinds': ['manual']})['worker_id']
    found = {'ack': [], 'cancel': []}
    for verb in ('ack', 'cancel'):
        statuses, ids = {}, []
        for n in range(trials):
            j = p.ok('POST', f'/v1/queues/{name}/jobs', {'kind': 'manual', 'payload': {'n': n}})['id']
            body = None
            if verb == 'ack':
                leased = p.ok('POST', f'/v1/queues/{name}/lease', {'worker_id': w, 'max_jobs': 1})['jobs']
                p.require(len(leased) == 1 and leased[0]['id'] == j, leased)
                body = {'lease_token': leased[0]['lease_token'], 'result': {'n': n}}
            stop, out = threading.Event(), {}
            spinner = threading.Thread(target=spin_delete, args=(j, stop, out))
            spinner.start()
            status, value = p.call('POST', f'/v1/jobs/{j}/{verb}', body)
            spinner.join(timeout=10)
            stop.set()
            spinner.join()
            p.require(status in (200, 409), (verb, j, status, value))
            statuses[status] = statuses.get(status, 0) + 1
            ids.append((j, status))
            if len(found[verb]) >= 3:
                break
            if n % 25 == 24:
                found[verb] = check(verb, ids)
        found[verb] = check(verb, ids)
        evidence(verb, trials=len(ids), statuses=statuses, defects=found[verb][:10])
    lost = {v: len(found[v]) for v in found}
    return any(lost.values()), f'committed transitions answered non-2xx or never notified: {lost} (ack {trials}, cancel {trials} trials max; delete spinner per job)'


def check(verb, ids):
    event = 'succeeded' if verb == 'ack' else 'cancelled'
    jobs = [j for j, _ in ids]
    done, sent = committed(jobs, event), notified(jobs)
    return [(j, s, j in sent) for j, s in ids if j in done and (s != 200 or j not in sent)]


def pass_notifications(trials=20, size=8):
    # Expiry and worker deletion end several leases in one pass while every job is being deleted.
    name = queue('p', webhook=HOOK, lease_seconds=30, max_attempts=1)
    problems = {'expire': [], 'worker delete': []}
    try:
        clock('set', at='2033-05-01T00:00:00Z')
        for path in ('expire', 'worker delete'):
            for n in range(trials):
                w = p.ok('POST', '/v1/workers', {'name': p.NAME, 'kinds': ['manual']})['worker_id']
                ids = p.ok('POST', f'/v1/queues/{name}/jobs:batch', {'jobs': [{'kind': 'manual', 'payload': {}}] * size})['ids']
                leased = p.ok('POST', f'/v1/queues/{name}/lease', {'worker_id': w, 'max_jobs': size})['jobs']
                p.require(sorted(x['id'] for x in leased) == sorted(ids), ('lease', leased))
                stop, out = threading.Event(), {}
                spinners = [threading.Thread(target=spin_delete, args=(j, stop, out)) for j in ids]
                for s in spinners:
                    s.start()
                status = None
                if path == 'expire':
                    clock('advance', seconds=31)
                else:
                    status, value = p.call('DELETE', f'/v1/workers/{w}')
                    p.require(status in (200, 204, 500), ('worker delete', status, value))
                deadline = time.monotonic() + 10
                for s in spinners:
                    s.join(timeout=max(.1, deadline - time.monotonic()))
                stop.set()
                for s in spinners:
                    s.join()
                p.require(len(out) == size, (path, 'jobs not terminal within 10 s', out))
                # Delete succeeds only on a finished job: each one went dead in this pass.
                p.require(set(out.values()) == {204}, (path, 'not every job was deleted', out))
                missing = sorted(set(ids) - notified(ids))
                if missing or status == 500:
                    problems[path].append({'trial': n, 'missing': missing, 'worker_delete': status})
                if problems[path]:
                    break
            evidence(path, trials=n + 1, problems=problems[path])
        return any(problems.values()), f'dead jobs without notification or 500 for a committed delete: ' + \
            ', '.join(f'{k} {len(v)} trial(s) {v[:1]}' for k, v in problems.items()) + f' ({trials} trials × {size} jobs max)'
    finally:
        # Deliver this probe's rows while the frozen clock still makes them due: once the
        # clock resets they would wait years, and U03's control drains the whole outbox.
        p.wait(lambda: p.sql(f"SELECT count(*) FROM webhook_outbox WHERE queue='{name}' AND delivered_at IS NULL") == '0', 60)
        clock('reset')


if __name__ == '__main__':
    try:
        present, detail = {'U39': unadvanceable, 'U40': never_matches, 'U41': zone_prefix, 'U42': frozen_offset,
            'U43': after_commit, 'U44': pass_notifications, 'U45': local_zone}[p.ID]()
        print(('PRESENT: ' if present else 'ABSENT: ') + p.ID + ' — ' + detail)
    except Exception as error:
        evidence('PROBE ERROR', error=repr(error))
        print('PROBE ERROR: ' + p.ID + ' — ' + repr(error), file=sys.stderr)
        sys.exit(2)
