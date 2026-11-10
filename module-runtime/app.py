import json,os
from http.server import BaseHTTPRequestHandler,HTTPServer
class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path!="/health":
            self.send_error(404);return
        body=json.dumps({"status":"ok","module":os.getenv("OMISHA_MODULE","core"),"customer":os.getenv("OMISHA_CUSTOMER","unknown"),"scope":"prototype"}).encode()
        self.send_response(200);self.send_header("Content-Type","application/json");self.send_header("Content-Length",str(len(body)));self.end_headers();self.wfile.write(body)
HTTPServer(("0.0.0.0",8000),Handler).serve_forever()
