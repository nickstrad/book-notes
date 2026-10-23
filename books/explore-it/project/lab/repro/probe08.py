#!/usr/bin/env python3
"""E08 neighbours: the database at start and frozen, a slow partner, the clock across DST.
Only public routes, Docker lifecycle, a second partner container and read-only SQL
are used. Every unexpected outcome or missing control is a PROBE ERROR, never ABSENT.
Run serially: each probe stops, freezes or retargets a neighbour of the whole lab.
"""
import concurrent.futures as futures
import json
import subprocess
import sys
import threading
import time
import urllib.error
import uuid
import probe as p
from probe04 import clock


def evidence(label, **values):
    print(json.dumps(dict(probe=p.ID, label=label, **values), sort_keys=True, default=str), file=sys.stderr, flush=True)


def state(service):
    return subprocess.check_output(p.COMPOSE + ['ps', '-a', '--format', '{{.State}} {{.Status}}', service],
        text=True, stdin=subprocess.DEVNULL).strip()


def health_within(seconds):
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        if p.ready('/health'):
            return True
        time.sleep(.5)
    return False


def start_without_database():
    # Restart only the API while PostgreSQL is stopped, then bring PostgreSQL back.
    try:
        p.docker('stop', 'postgres')
        p.docker('restart', 'api')
        time.sleep(4)
        during = state('api')
        p.docker('start', 'postgres')
        started = time.monotonic()
        recovered = health_within(30)
        elapsed = round(time.monotonic() - started, 1)
        after = state('api')
        evidence('api start without database', during=during, recovered=recovered, elapsed=elapsed, after=after)
        if not recovered:
            return True, f'api during outage={during!r}; health 200 within 30 s of database start=False; api after={after!r}'
        p.require(p.call('GET', '/v1/queues')[0] == 200, 'authenticated control failed after recovery')
        # Siblings: a process that dies on its own (not an operator stop) must come back.
        crashes = {}
        for service in ('api', 'postgres'):
            crash(service)
            crashes[service] = health_within(30)
        evidence('process crashes', recovered=crashes)
        return not all(crashes.values()), f'api during outage={during!r}; health 200 within 30 s of database start=True ({elapsed}s); after process crash recovered={crashes}'
    finally:
        # Cleanup only: an operator start is what the historical API needs.
        p.docker('start', 'postgres')
        p.wait(lambda: 'healthy' in state('postgres'), 30)
        p.docker('start', 'api')
        p.require(health_within(30), 'lab did not recover after U36')


def crash(service):
    container = subprocess.check_output(p.COMPOSE + ['ps', '-q', service], text=True).strip()
    pid = subprocess.check_output(['docker', 'inspect', '-f', '{{.State.Pid}}', container], text=True).strip()
    p.require(pid not in ('', '0'), ('no process to crash', service))
    # SIGKILL from the host is an unexpected exit; `docker kill` would count as an operator stop.
    subprocess.run(['kill', '-9', pid], check=True)
    p.wait(lambda: not p.ready('/health'), 10)


def timed(method, path, **kwargs):
    started = time.monotonic()
    try:
        status = p.call(method, path, **kwargs)[0]
    except (urllib.error.URLError, TimeoutError, ConnectionError) as error:
        status = repr(error)
    return status, round(time.monotonic() - started, 2)


def frozen_database():
    p.require(p.call('GET', '/health')[0] == 200, 'health control failed before freeze')
    freeze = 12
    p.docker('pause', 'postgres')
    # The timer ends the freeze whether or not the requests below ever return.
    thaw = threading.Timer(freeze, lambda: p.docker('unpause', 'postgres'))
    thaw.start()
    try:
        time.sleep(1)
        with futures.ThreadPoolExecutor(max_workers=2) as pool:
            health = pool.submit(timed, 'GET', '/health')
            auth = pool.submit(timed, 'GET', '/v1/queues')
            health, auth = health.result(timeout=30), auth.result(timeout=30)
    finally:
        thaw.join()
    p.require(health_within(20), 'health did not recover after unpause')
    p.require(p.call('GET', '/v1/queues')[0] == 200, 'authenticated control failed after unpause')
    evidence('frozen database', health=health, auth=auth, freeze_seconds=freeze)
    prompt = health[0] == 503 and health[1] <= 3 and auth[0] == 503 and auth[1] <= 8
    return not prompt, f'during freeze /health={health[0]} after {health[1]}s (bound 503 ≤3s); authenticated GET={auth[0]} after {auth[1]}s (bound 503 ≤8s); recovered after unpause'


def jobs_of(kind):
    return [j for j in p.ok('GET', p.QPATH + '/jobs?limit=500')['jobs'] if j['kind'] == kind]


def step(at):
    clock('set', at=at)
    time.sleep(2.5)


def dst():
    zone = 'America/New_York'
    made = []
    try:
        clock('set', at='2027-03-13T12:00:00Z')
        p.queue()
        for kind, cron in (('gap-0230', '30 2 * * *'), ('gap-0200', '0 2 * * *'), ('nine', '0 9 * * *')):
            made.append(p.schedule(name=p.NAME + kind, kind=kind, cron=cron, timezone=zone))
        for at in ('2027-03-14T07:00:30Z', '2027-03-14T13:00:30Z', '2027-03-15T06:00:30Z', '2027-03-15T06:30:30Z'):
            step(at)
        # Remove the spring schedules so the autumn jump cannot add a catch-up run to them.
        while made:
            p.ok('DELETE', f'/v1/schedules/{made.pop()["id"]}')
        clock('set', at='2027-11-07T04:00:00Z')
        for kind, cron in (('fb-0130', '30 1 * * *'), ('hourly', '0 * * * *')):
            made.append(p.schedule(name=p.NAME + kind, kind=kind, cron=cron, timezone=zone))
        for at in ('2027-11-07T05:00:30Z', '2027-11-07T05:30:30Z', '2027-11-07T06:00:30Z', '2027-11-07T06:30:30Z', '2027-11-07T07:00:30Z'):
            step(at)
        runs = {kind: sorted(j['run_at'] for j in jobs_of(kind)) for kind in ('gap-0230', 'gap-0200', 'nine', 'fb-0130', 'hourly')}
    finally:
        clock('reset')
        for s in made:
            p.ok('DELETE', f'/v1/schedules/{s["id"]}')
    evidence('dst runs', runs=runs)
    # Controls: the offset change outside the transition hour and wildcard hours hold at every tag.
    p.require([r[:16] for r in runs['nine']] == ['2027-03-13T14:00', '2027-03-14T13:00'], ('09:00 control', runs['nine']))
    p.require([r[:16] for r in runs['hourly']] == ['2027-11-07T05:00', '2027-11-07T06:00', '2027-11-07T07:00'], ('hourly control', runs['hourly']))
    gap = {k: [r[:10] for r in runs[k]] for k in ('gap-0230', 'gap-0200')}
    once_per_day = all(days == ['2027-03-14', '2027-03-15'] for days in gap.values())
    once_in_overlap = len(runs['fb-0130']) == 1
    return not (once_per_day and once_in_overlap), f'spring gap days={gap}; fall-back 01:30 runs={runs["fb-0130"]}; 09:00 and hourly controls hold'


def slow_partner():
    config = json.loads(subprocess.check_output(p.COMPOSE + ['config', '--format', 'json'], text=True))
    image = config['services']['hooksink']['image']
    network = config['networks']['default']['name']
    alias = 'slow-' + uuid.uuid4().hex[:8]
    container = 'queuelet-u37-' + alias
    subprocess.run(['docker', 'run', '-d', '--rm', '--name', container, '--network', network, '--network-alias', alias,
        '-p', '127.0.0.1::8201', image, 'hooksink'], check=True, capture_output=True, stdin=subprocess.DEVNULL)
    try:
        port = subprocess.check_output(['docker', 'port', container, '8201/tcp'], text=True).split(':')[-1].strip()
    except subprocess.CalledProcessError:
        subprocess.run(['docker', 'rm', '-f', container], capture_output=True, timeout=30)
        raise
    slow = 'http://127.0.0.1:' + port
    try:
        p.wait(lambda: p.ready_url(slow + '/admin/mode'), 10)
        p.ok('POST', '/admin/mode', {'mode': 'timeout'}, base=slow)
        # Tenant globex: 22 completed jobs whose destination never answers within the deadline.
        g = p.NAME + '-g'
        p.ok('POST', '/v1/queues', {'name': g, 'webhook': {'url': f'http://{alias}:8201/x', 'secret': 's'}}, key='globex-key')
        for _ in range(22):
            p.ok('POST', f'/v1/queues/{g}/jobs', {'kind': 'manual', 'payload': {}}, key='globex-key')
        w = p.ok('POST', '/v1/workers', {'name': g, 'kinds': ['manual']}, key='globex-key')['worker_id']
        leased = p.ok('POST', f'/v1/queues/{g}/lease', {'worker_id': w, 'max_jobs': 22}, key='globex-key')['jobs']
        p.require(len(leased) == 22, ('globex lease', len(leased)))
        for j in leased:
            p.ok('POST', f'/v1/jobs/{j["id"]}/ack', {'lease_token': j['lease_token'], 'result': {}}, key='globex-key')
        # Tenant acme: one completed job whose destination answers at once.
        p.queue(webhook={'url': 'http://hooksink:8201/hook', 'secret': 'report-secret'})
        j = p.enqueue(); token = p.lease()[0]['lease_token']
        p.ok('POST', f'/v1/jobs/{j}/ack', {'lease_token': token, 'result': {}})
        started = time.monotonic()
        delivered = None
        while time.monotonic() - started < 20:
            if any(r['status'] == 200 for r in p.records(j)):
                delivered = round(time.monotonic() - started, 1)
                break
            time.sleep(.5)
        slow_seen = len(p.ok('GET', '/admin/records', base=slow)['records'])
        p.require(slow_seen > 0, 'slow destination was never attempted')
        row = p.sql(f'SELECT attempts, delivered_at IS NOT NULL FROM webhook_outbox WHERE job_id={j}')
        evidence('slow partner', delivered_after=delivered, slow_requests=slow_seen, acme_row=row)
        return delivered is None, f'acme notification delivered after={"never" if delivered is None else str(delivered) + "s"} (bound 20 s) behind 22 globex rows to a timing-out destination; slow destination requests={slow_seen}; acme outbox attempts|delivered={row}'
    finally:
        # Cleanup only: let the slow partner accept, so this probe leaves no undelivered rows
        # for later probes that wait for the whole outbox to drain.
        try:
            p.ok('POST', '/admin/mode', {'mode': 'ok'}, base=slow)
            p.wait(lambda: p.sql("SELECT count(*) FROM webhook_outbox WHERE delivered_at IS NULL AND queue IN ('"
                + p.NAME + "', '" + p.NAME + "-g')") == '0', 60)
        finally:
            subprocess.run(['docker', 'rm', '-f', container], capture_output=True, timeout=30)


def ready_url(url):
    try:
        return p.call('GET', '', base=url)[0] == 200
    except (urllib.error.URLError, ConnectionError, TimeoutError):
        return False


p.ready_url = ready_url

if __name__ == '__main__':
    try:
        present, detail = {'U36': start_without_database, 'U37': slow_partner, 'U38': frozen_database, 'A5': dst}[p.ID]()
        print(('PRESENT: ' if present else 'ABSENT: ') + p.ID + ' — ' + detail)
    except Exception as error:
        evidence('PROBE ERROR', error=repr(error))
        print('PROBE ERROR: ' + p.ID + ' — ' + repr(error), file=sys.stderr)
        sys.exit(2)
