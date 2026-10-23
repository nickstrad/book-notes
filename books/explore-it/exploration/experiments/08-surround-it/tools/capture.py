"""E08 capture listener (testability stand-in for E01 TR1): records every POST's headers and raw body
as JSON lines; answers 200. Usage: python3 -I capture.py <bind-ip> <port> <out.jsonl>"""
import base64, json, sys, time
from http.server import BaseHTTPRequestHandler, HTTPServer
ip, port, out = sys.argv[1], int(sys.argv[2]), sys.argv[3]
class H(BaseHTTPRequestHandler):
    def do_POST(self):
        b = self.rfile.read(int(self.headers.get("Content-Length", 0)))
        with open(out, "a") as f:
            f.write(json.dumps({"at": time.time(), "path": self.path, "headers": dict(self.headers),
                                "raw_b64": base64.b64encode(b).decode()}) + "\n")
        self.send_response(200); self.end_headers(); self.wfile.write(b'{"ok":true}')
    def log_message(self, *a): pass
HTTPServer((ip, port), H).serve_forever()
