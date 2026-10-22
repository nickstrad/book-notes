import http.server, json, sys
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        b=self.rfile.read(int(self.headers.get('Content-Length',0)))
        with open(sys.argv[2],'a') as f: f.write(json.dumps({"sig":self.headers.get('X-Queuelet-Signature'),"body":b.decode()})+"\n")
        self.send_response(200); self.end_headers(); self.wfile.write(b'{}')
    def log_message(self,*a): pass
http.server.HTTPServer((sys.argv[1],18229),H).serve_forever()
