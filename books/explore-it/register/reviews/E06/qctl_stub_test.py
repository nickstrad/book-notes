import http.server, subprocess, threading, json, os, sys

captured = []
class H(http.server.BaseHTTPRequestHandler):
    def _handle(self):
        n = int(self.headers.get('Content-Length') or 0)
        body = self.rfile.read(n)
        captured.append((self.command, self.path, body))
        out = b'{"id": 7, "state": "queued", "payload": {"big": 9007199254740993, "fraction": 0.12345678901234567890123456789}, "result": {"big": 9007199254740993}}'
        self.send_response(201)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(out)
    do_POST = _handle
    do_GET = _handle
    def log_message(self, *a): pass

srv = http.server.HTTPServer(('127.0.0.1', 18399), H)
t = threading.Thread(target=srv.serve_forever, daemon=True); t.start()

env = dict(os.environ, QUEUELET_URL='http://127.0.0.1:18399', QUEUELET_KEY='acme-key')
def run(binpath, *args):
    captured.clear()
    r = subprocess.run([binpath]+list(args), capture_output=True, text=True, env=env, timeout=15)
    return r, list(captured)

for name, binpath in [('lesson-05', '/tmp/explore-it-review06/bin/qctl-lesson05'),
                      ('lesson-06', '/tmp/explore-it-review06/bin/qctl-lesson06')]:
    for label, payload in [('valid-big', '{"big":9007199254740993}'),
                           ('trailing-garbage', '{"big":9007199254740993}xyz'),
                           ('trailing-second-value', '{"big":9007199254740993} {"big":1}'),
                           ('malformed', '{"big":}')]:
        r, caps = run(binpath, 'enqueue', 'q1', '--kind', 'manual', '--payload', payload)
        body = caps[0][2].decode() if caps else '(no request)'
        print(f'{name} enqueue {label}: exit={r.returncode} sent_body={body} stdout={r.stdout.strip()!r} stderr={r.stderr.strip()[:80]!r}')
    r, caps = run(binpath, 'job', 'get', '7')
    print(f'{name} job-get display: exit={r.returncode} stdout={r.stdout.strip()}')
srv.shutdown()
