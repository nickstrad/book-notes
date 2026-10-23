#!/usr/bin/env python3
"""E06 assigned findings. Identical public instruments for current/REF binaries.
Only API mutations, partner controls, the API clock, container lifecycle and
read-only storage observations. All unexpected outcomes are PROBE ERROR.
"""
import collections
import concurrent.futures as futures
import decimal
import json
import pathlib
import subprocess
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request
import uuid

import probe as p
from probe04 import clock, lease_job

ID = p.ID
NUMBER_BYTES = '{"big":9007199254740993,"fraction":0.12345678901234567890123456789,"nested":[9223372036854775807],"small":42}'
NUMBERS = json.loads(NUMBER_BYTES, parse_float=decimal.Decimal)


def exact(method, path, raw=None, base=p.BASE):
    req = urllib.request.Request(base + path, None if raw is None else raw.encode(),
        {'Authorization': 'Bearer acme-key', 'Content-Type': 'application/json'}, method=method)
    with urllib.request.urlopen(req, timeout=15) as response:
        p.require(response.status in (200, 201), (path, response.status))
        return json.loads(response.read(), parse_float=decimal.Decimal)


def view(j):
    return exact('GET', f'/v1/jobs/{j}')


def exact_records(j):
    return [r for r in exact('GET', '/admin/records', base=p.HOOK)['records']
            if isinstance(r['body'], dict) and r['body'].get('job_id') == j]


def qctl(*args):
    # Exec the target container's built binary; never go run from current sources.
    command = p.COMPOSE + ['exec', '-T', '-e', 'QUEUELET_URL=http://localhost:8200',
        '-e', 'QUEUELET_KEY=acme-key', 'api', 'qctl'] + list(args)
    result = subprocess.run(command, stdin=subprocess.DEVNULL, capture_output=True, text=True, timeout=20)
    p.require(result.returncode == 0, ('qctl', args, result.returncode, result.stdout, result.stderr))
    return json.loads(result.stdout, parse_float=decimal.Decimal)


def status(j):
    code, value = p.checked('GET', f'/v1/jobs/{j}', statuses=(200, 404))
    if code == 404:
        p.require(value['error']['code'] == 'job_not_found', value)
    return code


def rows(j, table):
    p.require(table in ('jobs', 'attempts', 'job_events', 'webhook_outbox'), table)
    column = 'id' if table == 'jobs' else 'job_id'
    return int(p.sql(f'SELECT count(*) FROM {table} WHERE {column}={int(j)}'))


def delete_schedule():
    p.queue()
    s = p.schedule(enabled=False)
    path = f'/v1/schedules/{s["id"]}'
    foreign, fv = p.checked('DELETE', path, key='globex-key', statuses=(204, 404))
    p.require(p.ok('GET', path)['id'] == s['id'], 'foreign delete changed schedule')
    p.require(p.checked('DELETE', path)[0] == 204, 'own delete')
    p.require(p.checked('GET', path, statuses=(404,))[1]['error']['code'] == 'schedule_not_found', 'deleted read')
    repeated, rv = p.checked('DELETE', path, statuses=(204, 404))
    missing, mv = p.checked('DELETE', '/v1/schedules/999999999', statuses=(204, 404))
    for code, value in ((foreign, fv), (repeated, rv), (missing, mv)):
        if code == 404:
            p.require(value['error']['code'] == 'schedule_not_found', value)
    p.require((foreign, repeated, missing) in ((204, 204, 204), (404, 404, 404)), 'mixed DELETE semantics')
    return missing == 204, f'foreign/repeated/missing DELETE={foreign}/{repeated}/{missing}; owner delete=204, foreign untouched, deleted GET=404'


def schedule_payload():
    clock('set', at='2032-06-01T00:00:00Z')
    schedules = []
    try:
        p.queue()
        base = dict(name=p.NAME, queue=p.NAME, kind='manual', cron='* * * * *', timezone='UTC')
        bad = []
        for value in (None, [1, 2], 'x', 7, True):
            code, v = p.checked('POST', '/v1/schedules', dict(base, payload=value), statuses=(201, 400))
            bad.append(code)
            if code == 201:
                schedules.append(v['id'])
                p.require(v['payload'] == value, ('invalid stored form', v))
                p.require(p.checked('POST', f'/v1/schedules/{v["id"]}/run', {}, statuses=(400,))[1]['error']['code'] == 'invalid_job', 'invalid run')
            else:
                p.require(v['error']['code'] == 'invalid_schedule', v)
        omitted = p.ok('POST', '/v1/schedules', dict(base, name=p.NAME + '-omitted'))
        schedules.append(omitted['id'])
        good = p.schedule(name=p.NAME + '-object', payload={'big': NUMBERS['big']})
        schedules.append(good['id'])
        manual = p.ok('POST', f'/v1/schedules/{good["id"]}/run', {})['id']
        p.require(p.job(manual)['payload'] == {'big': NUMBERS['big']}, 'manual object mismatch')
        oc, ov = p.checked('POST', f'/v1/schedules/{omitted["id"]}/run', {}, statuses=(201, 400))
        if oc == 201:
            p.require(p.job(ov['id'])['payload'] == {}, 'omitted manual payload')
        else:
            p.require(ov['error']['code'] == 'invalid_job', ov)
        # PATCH retains its old field set and does not mutate payload.
        p.checked('PATCH', f'/v1/schedules/{good["id"]}', {'payload': {}}, statuses=(400,))
        p.require(p.ok('GET', f'/v1/schedules/{good["id"]}')['payload'] == {'big': NUMBERS['big']}, 'PATCH mutated payload')
        before = {j['id'] for j in p.ok('GET', p.QPATH + '/jobs')['jobs']}
        clock('advance', seconds=61)
        expected = len(schedules)
        def fired():
            return [j for j in p.ok('GET', p.QPATH + '/jobs')['jobs'] if j['id'] not in before]
        p.wait(lambda: len(fired()) == expected, 8)
        payloads = [j['payload'] for j in fired()]
        broken = all(c == 201 for c in bad)
        if broken:
            p.require(omitted['payload'] is None and oc == 400, ('historical omission', omitted, oc))
            p.require(sum(not isinstance(v, dict) for v in payloads) == 6, payloads)
        else:
            p.require(all(c == 400 for c in bad) and omitted['payload'] == {} and oc == 201, (bad, omitted, oc))
            p.require(len(payloads) == 2 and {} in payloads and {'big': NUMBERS['big']} in payloads, payloads)
        stored = p.ok('GET', '/v1/schedules')['schedules']
        p.require(sum(s['queue_id'] == good['queue_id'] for s in stored) == expected, 'rejected schedule persisted')
        return broken, f'null/array/string/number/bool create={bad}; omitted payload={omitted["payload"]}, run={oc}; cron fired {expected} templates; object run/cron and PATCH controls hold'
    finally:
        for sid in schedules:
            p.checked('DELETE', f'/v1/schedules/{sid}')
        clock('reset')


def worker_precision():
    p.queue(lease_seconds=30, max_attempts=5, backoff={'kind': 'linear', 'base_seconds': 0},
            webhook={'url': 'http://hooksink:8201/hook', 'secret': 'report-secret'})
    j = exact('POST', p.QPATH + '/jobs', '{"kind":"echo","payload":' + NUMBER_BYTES + '}')['id']
    p.require(view(j)['payload'] == NUMBERS, 'raw API write control failed')
    name = 'queuelet-probe06-' + uuid.uuid4().hex[:10]
    with tempfile.TemporaryFile(mode='w+') as log:
        command = p.COMPOSE + ['run', '--rm', '-T', '--no-deps', '--name', name,
            '-e', 'WORKER_QUEUE=' + p.NAME, '-e', 'WORKER_NAME=' + name, 'worker-a']
        process = subprocess.Popen(command, stdin=subprocess.DEVNULL, stdout=log, stderr=log)
        try:
            p.wait(lambda: p.job(j)['state'] == 'succeeded', 12)
            result = view(j)['result']
            p.wait(lambda: exact_records(j), 8)
            record = exact_records(j)[0]
            # hooksink's convenience body uses float64; raw_body is the received wire evidence.
            wire = json.loads(record['raw_body'], parse_float=decimal.Decimal)
            p.require(wire['job_id'] == j, 'wrong raw webhook')
            p.require(record['status'] == 200 and record['signature_valid'], record)
            p.require(wire['result'] == result, 'raw webhook differs from stored result')
            p.require(view(j)['payload'] == NUMBERS, 'worker changed payload')
            # Decoder's numeric consumers must still work after UseNumber.
            start = time.monotonic()
            sleep = p.enqueue(kind='sleep', payload={'seconds': .3})
            p.wait(lambda: p.job(sleep)['state'] == 'succeeded', 8)
            p.require(time.monotonic() - start >= .29, 'sleep numeric consumer skipped delay')
            fail = p.enqueue(kind='fail3', payload={'n': 1})
            p.wait(lambda: p.job(fail)['state'] == 'succeeded', 10)
            p.require(p.job(fail)['attempts'] == 2, 'fail numeric consumer ignored n=1')
            p.require(result.get('small') == 42 and result.get('nested') is not None, result)
            return result != NUMBERS, f'raw payload exact; worker echo big={result["big"]}, fraction={result["fraction"]}; stored result=valid signed webhook; sleep/fail numeric controls hold'
        finally:
            subprocess.run(['docker', 'stop', '-t', '2', name], capture_output=True, timeout=10)
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill(); process.wait()
            log.seek(0)
            if sys.exc_info()[0]:
                print('worker diagnostics:\n' + log.read(), file=sys.stderr)


def cli_precision():
    p.queue()
    # Two reported surfaces, each directly checked; siblings exercise --result/--json.
    j = qctl('enqueue', p.NAME, '--kind', 'manual', '--payload', NUMBER_BYTES)['id']
    written = view(j)['payload']
    direct = exact('POST', p.QPATH + '/jobs', '{"kind":"manual","payload":' + NUMBER_BYTES + '}')['id']
    p.require(view(direct)['payload'] == NUMBERS, 'direct API precision control')
    token = lease_job(j, p.worker())['lease_token']
    qctl('job', 'ack', str(j), '--lease-token', token, '--result', NUMBER_BYTES)
    result = view(j)['result']
    # Give GET an exact payload AND result from the API, independent of CLI input.
    token2 = lease_job(direct, p.worker())['lease_token']
    p.require(p.ack(direct, token2, {'big': NUMBERS['big']})[0] == 200, 'direct ack control')
    printed = qctl('job', 'get', str(direct))
    batch = qctl('batch', p.NAME, '--json', '{"jobs":[{"kind":"manual","payload":' + NUMBER_BYTES + '}]}')['ids'][0]
    sid = qctl('schedule', 'create', '--json', '{"name":"' + p.NAME + '","queue":"' + p.NAME + '","kind":"manual","payload":' + NUMBER_BYTES + ',"cron":"* * * * *","enabled":false}')['id']
    scheduled = exact('GET', f'/v1/schedules/{sid}')['payload']
    p.checked('DELETE', f'/v1/schedules/{sid}')
    observations = {'enqueue-input': written == NUMBERS,
        'job-get-payload': printed['payload'] == NUMBERS,
        'job-get-result': printed['result'] == {'big': NUMBERS['big']},
        'ack-result': result == NUMBERS, 'batch-json': view(batch)['payload'] == NUMBERS,
        'schedule-json': scheduled == NUMBERS}
    p.require(all(observations.values()) or not any(observations.values()), observations)
    return not all(observations.values()), 'exact numeric value by path=' + str(observations) + '; direct API input/result controls exact'


def invalid_values():
    p.queue(lease_seconds=120)
    outcomes = []
    for verb, field, bad_values, clean in (
        ('ack', 'result', ['a\0b', {'nested': ['a\0b']}, {'a\0b': 1}], {'literal': r'\u0000', 'text': 'clean'}),
        ('nack', 'error', ['a\0b'], 'clean error')):
        for bad in bad_values:
            j = p.enqueue()
            token = lease_job(j, p.worker())['lease_token']
            before = p.job(j)
            body = dict(lease_token=token, **{field: bad})
            if verb == 'nack': body['retry'] = False
            code, value = p.checked('POST', f'/v1/jobs/{j}/{verb}', body, statuses=(400, 409))
            expected = 'invalid_result' if verb == 'ack' else 'invalid_error'
            if code == 400:
                p.require(value['error']['code'] == expected and 'U+0000' in value['error']['message'], value)
                p.require('pq:' not in value['error']['message'], value)
            else:
                p.require(value['error']['code'] == 'state_conflict' and 'pq:' in value['error']['message'], value)
            after = p.job(j)
            for key in ('state', 'lease_token', 'leased_by', 'lease_expires_at', 'attempts'):
                p.require(after[key] == before[key], ('invalid value mutated lease', key))
            p.require(not p.ok('GET', f'/v1/jobs/{j}/attempts')['attempts'], 'invalid attempt inserted')
            body[field] = clean
            p.require(p.checked('POST', f'/v1/jobs/{j}/{verb}', body)[0] == 200, 'clean retry failed')
            final = p.job(j)
            p.require(final['state'] == ('succeeded' if verb == 'ack' else 'dead'), final)
            p.require(final['result' if verb == 'ack' else 'last_error'] == clean, final)
            p.require(len(p.ok('GET', f'/v1/jobs/{j}/attempts')['attempts']) == 1, 'clean attempt missing')
            outcomes.append(code)
    p.require(len(set(outcomes)) == 1, outcomes)
    return outcomes[0] == 409, f'ack string/nested/key and nack NUL statuses={outcomes}; valid lease unchanged, clean retry stored exact value and one attempt each'


def owner_races():
    # Public concurrency only: no fixture SQL or synchronization hook.
    counts = {kind: collections.Counter() for kind in ('enqueue', 'schedule', 'batch')}
    bugs = {kind: 0 for kind in counts}
    examples = {}
    for kind in counts:
        # Positive and missing-owner controls independent of race sampling.
        qname = p.NAME + '-' + kind + '-control'
        p.ok('POST', '/v1/queues', {'name': qname})
        if kind == 'schedule':
            path, body = '/v1/schedules', dict(name=qname, queue=qname, kind='manual', payload={}, cron='* * * * *', enabled=False)
        else:
            path = '/v1/queues/' + qname + ('/jobs' if kind == 'enqueue' else '/jobs:batch')
            body = {'kind': 'manual', 'payload': {}} if kind == 'enqueue' else {'jobs': [{'kind': 'manual', 'payload': {}}]}
        control = p.checked('POST', path, body, statuses=(201,))[1]
        if kind == 'schedule':
            p.require(p.ok('GET', f'/v1/schedules/{control["id"]}')['queue_id'] > 0, control)
            p.checked('DELETE', f'/v1/schedules/{control["id"]}')
        else:
            jid = control['id'] if kind == 'enqueue' else control['ids'][0]
            p.require(p.job(jid)['payload'] == {}, control)
            p.ok('POST', f'/v1/jobs/{jid}/cancel', {}); p.checked('DELETE', f'/v1/jobs/{jid}')
        p.checked('DELETE', '/v1/queues/' + qname)
        code, value = p.checked('POST', path, body, statuses=(404,))
        p.require(value['error']['code'] == 'queue_not_found', value)
        rounds = 50 if kind != 'batch' else 8
        with futures.ThreadPoolExecutor(max_workers=17) as pool:
            for round_no in range(rounds):
                name = p.NAME + '-' + kind + '-' + str(round_no)
                p.ok('POST', '/v1/queues', {'name': name})
                path = '/v1/schedules' if kind == 'schedule' else '/v1/queues/' + name + ('/jobs' if kind == 'enqueue' else '/jobs:batch')
                body = dict(name=name, queue=name, kind='manual', payload={}, cron='* * * * *', enabled=False) if kind == 'schedule' else ({'kind': 'manual', 'payload': {}} if kind == 'enqueue' else {'jobs': [{'kind': 'manual', 'payload': {}}]})
                gate = threading.Barrier(17)
                def insert(i):
                    gate.wait(timeout=10)
                    time.sleep((i % 10) * .0007)
                    return p.call('POST', path, body)
                def remove():
                    gate.wait(timeout=10)
                    return p.checked('DELETE', '/v1/queues/' + name, statuses=(204, 409))
                work = [pool.submit(insert, i) for i in range(16)]
                deletion = pool.submit(remove)
                results = [f.result(timeout=25) for f in work]
                dc, dv = deletion.result(timeout=25)
                if dc == 409: p.require(dv['error']['code'] == 'queue_not_empty', dv)
                accepted = []
                for code, value in results:
                    counts[kind][code] += 1
                    if code == 201:
                        accepted.append(value)
                    elif code == 404:
                        p.require(value['error']['code'] == 'queue_not_found', value)
                        p.require('constraint' not in value['error']['message'] and 'pq:' not in value['error']['message'], value)
                    elif (kind in ('enqueue', 'batch') and code == 400 or kind == 'schedule' and code == 500):
                        err = value['error']
                        p.require(err['code'] == ('database_error' if kind == 'schedule' else 'invalid_job') and
                            ('schedules_queue_owner' if kind == 'schedule' else 'jobs_queue_owner') in err['message'], value)
                        bugs[kind] += 1; examples[kind] = value
                    else:
                        raise RuntimeError(('unexpected race outcome', kind, code, value))
                p.require(not (dc == 204 and accepted), ('delete orphaned accepted write', dc, accepted))
                if dc == 409:
                    if kind == 'schedule':
                        listing = [s for s in p.ok('GET', '/v1/schedules')['schedules'] if s['name'] == name]
                        p.require({s['id'] for s in listing} == {s['id'] for s in accepted}, 'schedule race cardinality')
                        for s in listing: p.checked('DELETE', f'/v1/schedules/{s["id"]}')
                    else:
                        listing = p.ok('GET', '/v1/queues/' + name + '/jobs')['jobs']
                        expected = {v['id'] if kind == 'enqueue' else v['ids'][0] for v in accepted}
                        p.require({j['id'] for j in listing} == expected, 'job race cardinality')
                        for j in listing:
                            p.ok('POST', f'/v1/jobs/{j["id"]}/cancel', {}); p.checked('DELETE', f'/v1/jobs/{j["id"]}')
                    p.checked('DELETE', '/v1/queues/' + name)
                # Historical evidence needs both actually reported races; current
                # always exhausts the bounded sampling budget.
                if kind != 'batch' and bugs[kind] >= 2:
                    break
        p.require(counts[kind][404] > 0, ('no concurrent missing-owner observations', counts))
    orphaned = p.sql('SELECT (SELECT count(*) FROM jobs j WHERE NOT EXISTS(SELECT 1 FROM queues q WHERE q.id=j.queue_id)) + (SELECT count(*) FROM schedules s WHERE NOT EXISTS(SELECT 1 FROM queues q WHERE q.id=s.queue_id))')
    p.require(orphaned == '0', ('owner orphans', orphaned))
    p.require(bool(bugs['enqueue']) == bool(bugs['schedule']), ('only one reported race reproduced', counts, examples))
    return bool(bugs['enqueue']), f'concurrent statuses={ {k:dict(v) for k,v in counts.items()} }; FK errors={bugs}; accepted writes equal stored rows, no orphans; positive/missing-owner controls hold; historical examples={examples}'


def attempt_lifetime():
    clock('set', at='2032-07-01T00:00:00Z')
    try:
        p.queue(retention_hours=1, lease_seconds=120,
                webhook={'url': 'http://hooksink:8201/hook', 'secret': 'report-secret'})
        p.ok('POST', '/admin/mode', {'mode': '500'}, base=p.HOOK)
        left = {}
        for mode in ('purge', 'delete', 'retention'):
            j = p.enqueue()
            token = lease_job(j, p.worker())['lease_token']
            if mode == 'purge':
                p.require(p.nack(j, token, delay_seconds=600)[0] == 200, 'retry fixture')
            else:
                p.require(p.ack(j, token, {'kept': mode})[0] == 200, 'ack fixture')
            p.require(rows(j, 'attempts') == 1 and rows(j, 'jobs') == 1, 'real attempt fixture')
            if mode == 'purge':
                p.require(p.ok('POST', p.QPATH + '/purge', {})['deleted'] == 1, 'purge count')
            elif mode == 'delete':
                p.checked('DELETE', f'/v1/jobs/{j}')
            else:
                clock('advance', seconds=3601)
                p.wait(lambda: status(j) == 404, 8)
            p.require(status(j) == 404 and rows(j, 'jobs') == 0, ('job removal', mode))
            p.require(rows(j, 'job_events') > 0 and p.ok('GET', f'/v1/jobs/{j}/events')['events'], 'durable events lost')
            if mode != 'purge':
                p.require(rows(j, 'webhook_outbox') == 1, 'pending snapshot deleted')
                p.require(p.sql(f'SELECT delivered_at IS NULL FROM webhook_outbox WHERE job_id={j}') == 't', 'snapshot not pending')
            left[mode] = rows(j, 'attempts')
        p.require(set(left.values()) in ({0}, {1}), left)
        # Unrelated live job retains its attempt until its own removal.
        live = p.enqueue(); token = lease_job(live, p.worker())['lease_token']
        p.require(p.nack(live, token, delay_seconds=600)[0] == 200, 'live control nack')
        p.require(rows(live, 'attempts') == 1 and rows(live, 'jobs') == 1, 'live attempts removed')
        p.ok('POST', '/admin/mode', {'mode': 'ok'}, base=p.HOOK)
        # Prove independent snapshots actually deliver after their jobs disappear.
        clock('advance', seconds=6)
        p.wait(lambda: p.sql(f"SELECT count(*) FROM webhook_outbox WHERE body->>'result' LIKE '%kept%' AND delivered_at IS NULL") == '0', 10)
        return bool(left['delete']), f'attempt rows after purge/delete/retention={left}; removed jobs absent in API+storage; live attempt retained; durable events and pending deliveries survive'
    finally:
        p.ok('POST', '/admin/mode', {'mode': 'ok'}, base=p.HOOK)
        clock('reset')


def retention_policy():
    clock('set', at='2032-08-01T00:00:00Z')
    try:
        p.queue(retention_hours=1, lease_seconds=120)
        w = p.worker()
        ids = {}
        for terminal in ('dead', 'succeeded', 'cancelled'):
            j = p.enqueue(); ids[terminal] = j
            if terminal == 'cancelled':
                p.ok('POST', f'/v1/jobs/{j}/cancel', {})
            else:
                token = lease_job(j, w)['lease_token']
                if terminal == 'dead': p.require(p.nack(j, token, retry=False)[0] == 200, 'dead fixture')
                else: p.require(p.ack(j, token, {'ok': True})[0] == 200, 'success fixture')
            p.require(p.job(j)['state'] == terminal and rows(j, 'jobs') == 1, ('terminal fixture', terminal))
        # Below cutoff: all three remain, proving the policy/age fixture.
        clock('advance', seconds=3599)
        time.sleep(1.3)
        p.require(all(status(j) == 200 and rows(j, 'jobs') == 1 for j in ids.values()), 'premature retention')
        clock('advance', seconds=3)
        p.wait(lambda: all(status(ids[t]) == 404 for t in ('succeeded', 'cancelled')), 8)
        p.require(all(rows(ids[t], 'jobs') == 0 for t in ('succeeded', 'cancelled')), 'API/storage retention mismatch')
        dead_status = status(ids['dead']); dead_rows = rows(ids['dead'], 'jobs')
        p.require((dead_status, dead_rows) in ((200, 1), (404, 0)), 'dead API/storage mismatch')
        if dead_status == 200:
            p.require(p.job(ids['dead'])['state'] == 'dead', 'dead changed state')
        # Another sweep/age boundary rules out a one-cycle observation.
        clock('advance', seconds=3600); time.sleep(1.3)
        p.require(status(ids['dead']) == dead_status and rows(ids['dead'], 'jobs') == dead_rows, 'dead retention unstable')
        if dead_status == 200:
            p.ok('POST', f'/v1/jobs/{ids["dead"]}/retry', {})
            p.require(p.job(ids['dead'])['state'] == 'queued', 'retained dead not retryable')
        return dead_status == 404, f'API clock ages 3599/3602/7202s: all remain below cutoff; succeeded/cancelled API=404 storage=0; dead API={dead_status} storage={dead_rows}; retained dead retry control holds when present'
    finally:
        clock('reset')


RUN = {'U27': delete_schedule, 'U28': schedule_payload, 'U29': worker_precision,
       'U30': cli_precision, 'U31': invalid_values, 'U32': owner_races,
       'U33': attempt_lifetime, 'A4': retention_policy}
if __name__ == '__main__':
    try:
        present, detail = RUN[ID]()
        print(('PRESENT: ' if present else 'ABSENT: ') + ID + ' — ' + detail)
    except Exception as error:
        print('PROBE ERROR: ' + ID + ' — ' + repr(error), file=sys.stderr)
        sys.exit(2)
