#!/usr/bin/env python3
"""Disclosed E01 findings: immutable black-box oracles for current/historical SUTs.

Only public commands, partner controls, Docker lifecycle and read-only SQL are
used. Unexpected statuses, missing fixtures and failed observations are errors,
not evidence of absence. Run serially: outage/partner cases affect the lab.
"""
import datetime as dt
import json
import os
import pathlib
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid
import zoneinfo

ROOT = pathlib.Path(__file__).resolve().parents[2]
BASE = os.getenv('QUEUELET_URL', 'http://localhost:18200')
HOOK = os.getenv('QUEUELET_HOOK_URL', 'http://localhost:18201')
COMPOSE = ['docker', 'compose', '-f', str(ROOT / 'deploy/compose.yml')]
ID = sys.argv[1]
NAME = 'repro-' + ID.lower() + '-' + uuid.uuid4().hex[:10]
QPATH = '/v1/queues/' + NAME


def call(method, path, body=None, key='acme-key', base=BASE):
    request = urllib.request.Request(base + path,
        None if body is None else json.dumps(body).encode(),
        {'Authorization': 'Bearer ' + key, 'Content-Type': 'application/json'}, method=method)
    try:
        with urllib.request.urlopen(request, timeout=15) as response:
            raw = response.read()
            return response.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as error:
        raw = error.read()
        return error.code, json.loads(raw) if raw else None


def checked(method, path, body=None, statuses=(200, 201, 204), **kwargs):
    status, value = call(method, path, body, **kwargs)
    require(status in statuses, (method, path, status, value))
    return status, value


def ok(method, path, body=None, **kwargs):
    return checked(method, path, body, **kwargs)[1]


def require(condition, detail):
    if not condition:
        raise RuntimeError(detail)


def sql(query):
    require(query.lstrip().upper().startswith('SELECT '), 'read-only SQL required')
    return subprocess.check_output(COMPOSE + ['exec', '-T', '-e',
        'PGOPTIONS=-c default_transaction_read_only=on', 'postgres', 'psql',
        '-U', 'queuelet', '-d', 'queuelet', '-At', '-v', 'ON_ERROR_STOP=1', '-c', query], text=True, stdin=subprocess.DEVNULL).strip()


def docker(*args):
    subprocess.run(COMPOSE + list(args), check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, stdin=subprocess.DEVNULL)


def wait(test, seconds=6):
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        answer = test()
        if answer:
            return answer
        time.sleep(.12)
    raise RuntimeError('timed out waiting for required observation')


def queue(**kwargs):
    return ok('POST', '/v1/queues', dict(name=NAME, **kwargs))


def enqueue(**kwargs):
    return ok('POST', QPATH + '/jobs', dict({'kind': 'manual', 'payload': {}}, **kwargs))['id']


def worker():
    return ok('POST', '/v1/workers', {'name': NAME, 'kinds': ['manual']})['worker_id']


def lease(w=None, count=1):
    return ok('POST', QPATH + '/lease', {'worker_id': w or worker(), 'max_jobs': count})['jobs']


def job(j):
    return ok('GET', '/v1/jobs/' + str(j))


def ack(j, token, value=None):
    return checked('POST', f'/v1/jobs/{j}/ack', {'lease_token': token, 'result': value}, statuses=(200, 409))


def nack(j, token, **kwargs):
    return checked('POST', f'/v1/jobs/{j}/nack', dict(lease_token=token, error='failure', **kwargs), statuses=(200, 409))


def schedule(**kwargs):
    return ok('POST', '/v1/schedules', dict({'name': NAME, 'queue': NAME, 'kind': 'manual',
        'payload': {}, 'cron': '* * * * *', 'timezone': 'UTC', 'enabled': True}, **kwargs))


def records(j):
    return [r for r in ok('GET', '/admin/records', base=HOOK)['records']
        if isinstance(r['body'], dict) and r['body'].get('job_id') == j]


def stamp(value):
    return dt.datetime.fromisoformat(value.replace('Z', '+00:00'))


def sleep_until(instant):
    while True:
        remaining = (instant - dt.datetime.now(dt.timezone.utc)).total_seconds()
        if remaining <= 0:
            return
        time.sleep(min(remaining, 1))


def present(flag, detail):
    return flag, detail


def ready(path):
    # Restart readiness tolerates transport loss but still requires a live 200.
    # Oracle requests themselves are never converted to absence on failure.
    try:
        return call('GET', path)[0] == 200
    except (urllib.error.URLError, ConnectionError, TimeoutError):
        return False


def recover():
    docker('start', 'postgres')
    wait(lambda: ready('/v1/queues'), 20)
    docker('restart', 'api')
    wait(lambda: ready('/health'), 20)


def run():
    if ID == 'D01':
        p = subprocess.run(COMPOSE + ['exec', '-T', '-e', 'QUEUELET_URL=http://localhost:8200',
            '-e', 'QUEUELET_KEY=acme-key', 'api', 'qctl', 'job', 'get', '999999999999'], capture_output=True, text=True, stdin=subprocess.DEVNULL)
        require('404' in p.stderr, ('CLI did not reach missing-job oracle', p.returncode, p.stderr))
        return present(p.returncode == 0, f'CLI missing-job exit={p.returncode}; stderr={p.stderr.strip()}')
    if ID == 'D02':
        status, _ = checked('DELETE', QPATH, statuses=(204, 404))
        return present(status == 204, f'delete absent queue HTTP {status}')
    if ID in ('D03', 'D18'):
        queue(lease_seconds=1)
    elif ID == 'D20':
        queue(lease_seconds=1, max_attempts=1)
    elif ID in ('D04', 'D17', 'D24', 'U02', 'U03'):
        queue(webhook={'url': 'http://hooksink:8201/hook', 'secret': 'report-secret'})
    else:
        queue()
    if ID == 'D03':
        j = enqueue(); l = lease()[0]
        require(l['id'] == j and l['attempt'] == 1, l)
        wait(lambda: job(j)['state'] != 'leased', 5)
        v = job(j); attempts = ok('GET', f'/v1/jobs/{j}/attempts')['attempts']
        require(v['state'] == 'queued', v)
        return present(v['attempts'] != 1 or len(attempts) != 1,
            f'expired lease counted={v["attempts"]}; attempt rows={len(attempts)}')
    if ID == 'D20':
        # No fabricated attempts: first crash at max_attempts=1 must exhaust.
        j = enqueue(); l = lease()[0]
        require(l['id'] == j and l['attempt'] == 1, l)
        wait(lambda: job(j)['state'] != 'leased', 5)
        v = job(j)
        require(v['state'] in ('queued', 'dead'), v)
        return present(v['state'] != 'dead', f'first crash max_attempts=1 leaves {v["state"]}; completed attempts={v["attempts"]}')
    if ID in ('D04', 'D17', 'D24', 'U02'):
        if ID == 'D04':
            ok('POST', '/admin/mode', {'mode': '500'}, base=HOOK)
        try:
            j = enqueue(); l = lease()[0]
            value = {'text': 'x' * 1500} if ID == 'D17' else {'ok': True, 'n': 4}
            require(ack(j, l['lease_token'], value)[0] == 200, 'valid ack rejected')
            wait(lambda: records(j), 8)
            record = records(j)[0]
            if ID == 'D04':
                observed = wait(lambda: sql(f'SELECT attempts FROM webhook_outbox WHERE job_id={j}') not in ('', '0'))
                delivered = sql(f'SELECT delivered_at IS NOT NULL FROM webhook_outbox WHERE job_id={j}')
                require(record['status'] == 500 and observed, record)
                return present(delivered == 't', f'partner HTTP 500; marked delivered={delivered}')
            if ID == 'D17':
                sent = record['body']['result']
                # Type is U02; decode strings before checking truncation independently.
                if isinstance(sent, str):
                    try: sent = json.loads(sent)
                    except json.JSONDecodeError: pass
                clipped = sent != value
                # Also exercise dead-letter errors independently of result type.
                e = enqueue(); el = lease()[0]; message = 'failure-' + 'z' * 400
                checked('POST', f'/v1/jobs/{e}/nack', {'lease_token': el['lease_token'], 'error': message, 'retry': False})
                wait(lambda: records(e), 8)
                error = records(e)[0]['body']['last_error']
                return present(clipped or error != message, f'result complete={not clipped}; error chars={len(error)}/{len(message)}')
            if ID == 'U02':
                sent = record['body']['result']
                return present(not isinstance(sent, dict) or sent != value,
                    f'result type={type(sent).__name__}; expected object')
            return present(not record['signature_valid'], f'signature_valid={record["signature_valid"]}')
        finally:
            ok('POST', '/admin/mode', {'mode': 'ok'}, base=HOOK)
    if ID == 'D05':
        statuses = [checked('POST', QPATH + '/jobs', {'kind': 'manual', 'payload': {}, 'priority': n}, statuses=(201, 400))[0] for n in (-1, 10)]
        return present(any(s != 400 for s in statuses), f'priority -1/10 statuses={statuses}')
    if ID == 'D06':
        a, _ = checked('POST', QPATH + '/jobs:batch', {'jobs': []}, statuses=(201, 400))
        b, v = checked('POST', QPATH + '/jobs:batch', {'jobs': [{'kind': 'manual', 'payload': {}}] * 501}, statuses=(201, 400))
        count = len(ok('GET', QPATH + '/jobs?limit=500')['jobs'])
        return present(a != 400 or b != 400 or count != 0, f'batch zero={a}, 501={b}, persisted={count}')
    if ID == 'D07':
        statuses = [checked('POST', '/v1/queues', {'name': NAME + suffix}, statuses=(201, 400))[0] for suffix in ('/archive', ' space')]
        return present(any(s != 400 for s in statuses), f'slash/space queue create statuses={statuses}; contract requires rejection')
    if ID == 'D08':
        value = 9007199254740993
        j = enqueue(payload={'id': value}); stored = job(j)['payload']['id']; leased = lease()[0]['payload']['id']
        return present(stored != value or leased != value, f'input={value}; stored={stored}; leased={leased}')
    if ID == 'D09':
        j = enqueue(max_attempts=1); l = lease()[0]
        require(nack(j, l['lease_token'], delay_seconds=3600)[0] == 200, 'valid nack rejected')
        state = job(j)['state']
        require(state in ('dead', 'retry_wait'), state)
        return present(state != 'dead', f'first nack max_attempts=1 leaves {state}')
    if ID == 'D10':
        enqueue(delay_seconds=3600); v = ok('GET', QPATH + '/stats')
        exact = int(sql(f"SELECT count(*) FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE q.name='{NAME}' AND j.state='queued'"))
        return present(v['queued'] != exact, f'stats queued={v["queued"]}; read-only SQL queued={exact}')
    if ID == 'D11':
        v = ok('POST', QPATH + '/jobs:batch', {'jobs': [{'kind': 'manual', 'payload': {}}] * 12})
        ids = v['ids']; w = worker(); first = lease(w)[0]
        require(nack(first['id'], first['lease_token'], delay_seconds=0)[0] == 200, 'release failed')
        wait(lambda: job(first['id'])['state'] == 'queued')
        nextjob = lease(w)[0]
        return present(first['id'] != ids[0] or nextjob['id'] != ids[0],
            f'oldest={ids[0]}, first={first["id"]}, next after public nack/promote={nextjob["id"]}')
    if ID == 'D12':
        delayed = enqueue(delay_seconds=3600); retrying = enqueue(); l = lease()[0]
        require(l['id'] == retrying, l)
        require(nack(retrying, l['lease_token'], delay_seconds=3600)[0] == 200, 'valid nack rejected')
        queued = enqueue(); ok('POST', QPATH + '/purge', {})
        statuses = [checked('GET', f'/v1/jobs/{j}', statuses=(200, 404))[0] for j in (delayed, retrying, queued)]
        return present(any(s != 404 for s in statuses), f'purged scheduled/retry_wait/queued GET statuses={statuses}')
    if ID == 'D13':
        start = time.monotonic(); ok('POST', QPATH + '/pause', {}); s = schedule()
        # Observe two actual minute boundaries; no clock hook or SQL setup writes.
        sleep_until(stamp(s['next_run_at']) + dt.timedelta(seconds=62))
        require(not ok('GET', QPATH + '/jobs')['jobs'], 'pause allowed deliveries')
        ok('POST', QPATH + '/resume', {}); time.sleep(2.2)
        count = len(ok('GET', QPATH + '/jobs')['jobs'])
        advanced = ok('GET', f'/v1/schedules/{s["id"]}')['next_run_at']
        require(advanced and stamp(advanced) > dt.datetime.now(dt.timezone.utc), 'scheduler did not advance during observation')
        ok('DELETE', f'/v1/schedules/{s["id"]}')
        return present(count > 1, f'resume after two missed minute ticks produced={count}; elapsed={time.monotonic()-start:.1f}s; maximum allowed=1')
    if ID == 'D14':
        states = []
        for terminal in ('cancelled', 'succeeded'):
            j = enqueue()
            if terminal == 'cancelled': ok('POST', f'/v1/jobs/{j}/cancel', {})
            else:
                l = lease()[0]; require(l['id'] == j, l); require(ack(j, l['lease_token'])[0] == 200, 'ack failed')
            status, _ = checked('POST', f'/v1/jobs/{j}/retry', {}, statuses=(200, 409))
            states.append((terminal, status, job(j)['state']))
            # If v0 revived it, cancel to isolate the next fixture.
            if status == 200: ok('POST', f'/v1/jobs/{j}/cancel', {})
        return present(any(s != 409 or a != b for a, s, b in states), f'non-dead retry outcomes={states}')
    if ID == 'D15':
        j = enqueue(); s = schedule(enabled=False)
        status, _ = checked('DELETE', QPATH, statuses=(204, 409))
        js = checked('GET', f'/v1/jobs/{j}', statuses=(200, 404))[0]
        ss = checked('GET', f'/v1/schedules/{s["id"]}', statuses=(200, 404))[0]
        require((status, js, ss) in ((204, 200, 200), (409, 200, 200)), (status, js, ss))
        return present(status == 204, f'nonempty owner DELETE={status}; dependent job/schedule={js}/{ss}')
    if ID == 'D16':
        s = schedule(cron='0 0 1 1 *')
        edited = ok('PATCH', f'/v1/schedules/{s["id"]}', {'cron': '* * * * *', 'timezone': 'Asia/Tokyo'})
        disabled = schedule(name=NAME + '-disabled', enabled=False)
        enabled = ok('PATCH', f'/v1/schedules/{disabled["id"]}', {'enabled': True})
        changed = edited['next_run_at'] != s['next_run_at']
        return present(not changed or enabled['next_run_at'] is None,
            f'cron/zone recomputed={changed}; enabling sets next_run_at={enabled["next_run_at"]}')
    if ID == 'D18':
        j = enqueue(); a = lease()[0]
        wait(lambda: job(j)['state'] == 'queued', 5); b = lease()[0]
        old, _ = ack(j, a['lease_token'], {'owner': 'A'})
        current, _ = ack(j, b['lease_token'], {'owner': 'B'})
        return present(old != 409 or current != 200 or job(j)['result'] != {'owner': 'B'},
            f'expired A ack={old}; current B ack={current}; stored result={job(j)["result"]}')
    if ID == 'D19':
        outcomes = []
        # Both token and state are part of F02, with positive control.
        j = enqueue(); l = lease()[0]
        wrong, _ = nack(j, 'wrong-token', delay_seconds=3600)
        outcomes.append(('wrong-token', wrong, job(j)['state']))
        if wrong == 200: ok('POST', f'/v1/jobs/{j}/cancel', {})
        else:
            ok('POST', f'/v1/jobs/{j}/cancel', {})
            late, _ = nack(j, l['lease_token'], delay_seconds=3600)
            outcomes.append(('cancelled', late, job(j)['state']))
        queued = enqueue(); never, _ = nack(queued, 'never-leased', delay_seconds=3600)
        outcomes.append(('queued', never, job(queued)['state']))
        return present(any(status != 409 for _, status, _ in outcomes), f'nack incompatible outcomes={outcomes}')
    if ID == 'D21':
        j = enqueue(); outcomes = []
        for method, suffix, body in [('GET', '', None), ('GET', '/attempts', None), ('POST', '/cancel', {}),
            ('POST', '/retry', {}), ('POST', '/ack', {'lease_token': 'x'}),
            ('POST', '/nack', {'lease_token': 'x', 'retry': False}),
            ('POST', '/heartbeat', {'lease_token': 'x', 'extend_seconds': 1}), ('DELETE', '', None)]:
            status, _ = checked(method, f'/v1/jobs/{j}' + suffix, body, key='globex-key', statuses=(200, 204, 404, 409))
            outcomes.append(status)
        return present(any(s != 404 for s in outcomes), f'foreign namespace get/attempts/cancel/retry/ack/nack/heartbeat/delete={outcomes}')
    if ID in ('D22', 'U04'):
        try:
            docker('stop', 'postgres'); time.sleep(2.2)
            if ID == 'U04':
                auth, _ = checked('GET', '/v1/queues', statuses=(401, 503))
                health, _ = checked('GET', '/health', statuses=(200, 503))
                return present(auth != 503 or health != 503, f'during DB outage auth={auth}; health={health}')
            docker('start', 'postgres')
            wait(lambda: ready('/v1/queues'), 20)
            j = enqueue(delay_seconds=1); time.sleep(4)
            state = job(j)['state']; status, health = checked('GET', '/health', statuses=(200, 503))
            return present(state != 'queued', f'after DB recovery delayed job={state}; health HTTP {status} {health["status"]}')
        finally:
            recover()
    if ID == 'D23':
        now = dt.datetime.now(dt.timezone.utc); zone = zoneinfo.ZoneInfo('America/New_York')
        month = 1 if now.astimezone(zone).utcoffset() == dt.timedelta(hours=-4) else 7
        s = schedule(cron=f'0 2 1 {month} *', timezone='America/New_York')
        local = stamp(s['next_run_at']).astimezone(zone)
        return present(local.hour != 2, f'opposite-season cron requires 02:00 America/New_York; next local fire={local.isoformat()}')
    if ID == 'D25':
        status, v = checked('POST', QPATH + '/lease', {'worker_id': 999999999, 'max_jobs': 1}, statuses=(200, 404))
        require('error' in v, v)
        return present(status != 404, f'unknown-worker error HTTP {status}')
    if ID == 'D26':
        patch, _ = checked('PATCH', QPATH, {'lease_secs': 99}, statuses=(200, 400))
        enqueue(); filtered, _ = checked('GET', QPATH + '/jobs?state=queded', statuses=(200, 400))
        return present(patch != 400 or filtered != 400, f'typo field PATCH={patch}; unknown state GET={filtered}')
    if ID == 'D27':
        s = schedule(); control = schedule(name=NAME + '-control', kind='control')
        ok('PATCH', f'/v1/schedules/{s["id"]}', {'enabled': False})
        sleep_until(stamp(s['next_run_at']) + dt.timedelta(seconds=2.2))
        observed = ok('GET', QPATH + '/jobs')['jobs']
        count = sum(j['kind'] == 'manual' for j in observed)
        require(any(j['kind'] == 'control' for j in observed), 'enabled schedule control did not fire')
        ok('DELETE', f'/v1/schedules/{s["id"]}')
        ok('DELETE', f'/v1/schedules/{control["id"]}')
        return present(count != 0, f'disabled schedule after real minute boundary produced={count}; enabled control fired')
    if ID == 'U01':
        # Use a disabled schedule to isolate validation from background poisoning.
        s = schedule(enabled=False)
        statuses = [checked('PATCH', f'/v1/schedules/{s["id"]}', body, statuses=(200, 400))[0]
            for body in ({'cron': 'bogus'}, {'timezone': 'Mars/Base'})]
        ok('DELETE', f'/v1/schedules/{s["id"]}')
        return present(any(s != 400 for s in statuses), f'invalid cron/zone PATCH statuses={statuses}')
    if ID == 'U03':
        # Remove unrelated outbox work using a fresh API/partner start, never SQL writes.
        # Six public terminal transitions create the observed timeout-row fixture.
        ok('POST', '/admin/mode', {'mode': 'timeout'}, base=HOOK)
        ids = []
        try:
            for _ in range(6):
                j = enqueue(); l = lease()[0]
                require(l['id'] == j and ack(j, l['lease_token'], {})[0] == 200, 'fixture ack failed')
                ids.append(j)
            wait(lambda: any(records(j) for j in ids), 10)
            start = time.monotonic(); delayed = enqueue(delay_seconds=1)
            wait(lambda: job(delayed)['state'] == 'queued', 25)
            slow = time.monotonic() - start
        finally:
            ok('POST', '/admin/mode', {'mode': 'ok'}, base=HOOK)
            # Let the existing serial batch drain before its control.
            time.sleep(2.3)
            wait(lambda: sql('SELECT count(*) FROM webhook_outbox WHERE delivered_at IS NULL') == '0', 20)
        start = time.monotonic(); control = enqueue(delay_seconds=1)
        wait(lambda: job(control)['state'] == 'queued', 6)
        baseline = time.monotonic() - start
        require(baseline <= 5, f'control too slow to support approximation: {baseline:.2f}s')
        return present(slow > 5, f'six-timeout-row promotion={slow:.2f}s; ok-partner control={baseline:.2f}s; warm-lab bound=5s')
    if ID == 'U05':
        # Existing-key and intra-batch duplicates must return original ids; invalid
        # trailing work must roll back new rows without removing existing work.
        original = enqueue(idempotency_key='existing')
        body = {'jobs': [{'kind': 'manual', 'payload': {}, 'idempotency_key': 'existing'},
                         {'kind': 'manual', 'payload': {}, 'idempotency_key': 'new'},
                         {'kind': '', 'payload': {}}]}
        status, _ = checked('POST', QPATH + '/jobs:batch', body, statuses=(400,))
        persisted = ok('GET', QPATH + '/jobs')['jobs']
        broken = [v['id'] for v in persisted] != [original]
        good = ok('POST', QPATH + '/jobs:batch', {'jobs': [
            {'kind': 'manual', 'payload': {}, 'idempotency_key': 'same'},
            {'kind': 'manual', 'payload': {}, 'idempotency_key': 'same'},
            {'kind': 'manual', 'payload': {}, 'idempotency_key': 'existing'}]})['ids']
        require(good[0] == good[1] and good[2] == original, ('idempotency control failed', good))
        return present(broken, f'invalid trailing element HTTP {status}; persisted={len(persisted)} expected=1; duplicate ids preserved')
    if ID == 'U06':
        statuses = []
        for body in ({'max_attempts': 0}, {'retention_hours': -1}, {'lease_seconds': 0}):
            statuses.append(checked('PATCH', QPATH, body, statuses=(200, 400))[0])
            statuses.append(checked('POST', '/v1/queues', dict(name=NAME + uuid.uuid4().hex[:4], **body), statuses=(201, 400))[0])
        return present(any(s != 400 for s in statuses), f'invalid max/retention/lease PATCH/create statuses={statuses}')
    if ID == 'U07':
        status, _ = checked('POST', '/v1/queues', {'name': NAME + '-bad', 'backoff': {'kind': 'zigzag', 'base_seconds': 5}}, statuses=(201, 400))
        return present(status != 400, f'unknown backoff kind create HTTP {status}')
    if ID == 'U08':
        start = time.monotonic(); j = enqueue(run_at='2020-01-01T00:00:00Z')
        age = ok('GET', QPATH + '/stats')['oldest_queued_age']; elapsed = time.monotonic() - start
        require(job(j)['state'] == 'queued', 'fixture not queued')
        return present(age > elapsed + 2, f'new backdated job queued age={age:.3f}s; elapsed+rounding bound={elapsed+2:.3f}s')
    raise RuntimeError('unknown reproduction id ' + ID)


if __name__ == '__main__':
    try:
        defect, detail = run()
        print(('PRESENT: ' if defect else 'ABSENT: ') + ID + ' — ' + detail)
    except Exception as error:
        print('PROBE ERROR: ' + ID + ' — ' + repr(error), file=sys.stderr)
        sys.exit(2)
