"""Omisha OS licensing prototype: Ed25519 signed, company/module entitlements."""
import base64, datetime as dt, hashlib, hmac, json, os, re, secrets, sqlite3, threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlparse
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey
from cryptography.hazmat.primitives import serialization

DATA=Path("/data"); DATA.mkdir(exist_ok=True)
DB=DATA/"licenses.sqlite3"; KEY=DATA/"signing.key"
MODULES={"core","carbon","manufacturing","finance","ai","iot","crm","compliance","projects"}
LOCK=threading.RLock()
def b64(b): return base64.urlsafe_b64encode(b).rstrip(b"=").decode()
def raw(s): return base64.urlsafe_b64decode(s+"="*(-len(s)%4))
def canonical(d): return json.dumps(d,sort_keys=True,separators=(",",":"),ensure_ascii=False).encode()
if not KEY.exists():
    with open(KEY,"xb") as f:
        os.chmod(KEY,0o600)
        f.write(Ed25519PrivateKey.generate().private_bytes(serialization.Encoding.Raw,serialization.PrivateFormat.Raw,serialization.NoEncryption()))
private=Ed25519PrivateKey.from_private_bytes(KEY.read_bytes())
public=b64(private.public_key().public_bytes(serialization.Encoding.Raw,serialization.PublicFormat.Raw))
with sqlite3.connect(DB) as db:
    db.execute("CREATE TABLE IF NOT EXISTS licenses (id TEXT PRIMARY KEY, company TEXT NOT NULL, payload TEXT NOT NULL, revoked INTEGER NOT NULL DEFAULT 0)")
    db.execute("CREATE TABLE IF NOT EXISTS activations (license_id TEXT NOT NULL, installation_id TEXT NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY(license_id,installation_id))")
def utc(): return dt.datetime.now(dt.timezone.utc).isoformat()
def token(payload):
    return {"payload":payload,"signature":b64(private.sign(canonical(payload)))}
def valid(doc):
    from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PublicKey
    Ed25519PublicKey.from_public_bytes(raw(public)).verify(raw(doc["signature"]),canonical(doc["payload"]))
    return doc["payload"]
class Handler(BaseHTTPRequestHandler):
    def reply(self,code,obj):
        body=json.dumps(obj,ensure_ascii=False).encode()
        self.send_response(code);self.send_header("Content-Type","application/json; charset=utf-8")
        self.send_header("Cache-Control","no-store");self.send_header("Content-Length",str(len(body)));self.end_headers();self.wfile.write(body)
    def admin(self):
        value=os.getenv("OMISHA_LICENSE_ADMIN_TOKEN","")
        return bool(value) and hmac.compare_digest(self.headers.get("X-Admin-Token",""),value)
    def body(self):
        n=int(self.headers.get("Content-Length","0"))
        if n<0 or n>16384: raise ValueError("body too large")
        return json.loads(self.rfile.read(n))
    def do_GET(self):
        if self.path=="/health": return self.reply(200,{"status":"ok"})
        if self.path=="/v1/public-key": return self.reply(200,{"algorithm":"Ed25519","public_key":public})
        if self.path=="/v1/licenses":
            if not self.admin():return self.reply(401,{"error":"unauthorized"})
            with sqlite3.connect(DB) as db:
                rows=db.execute("SELECT id,company,revoked FROM licenses ORDER BY rowid DESC").fetchall()
            return self.reply(200,{"licenses":[{"id":r[0],"company":r[1],"revoked":bool(r[2])} for r in rows]})
        return self.reply(404,{"error":"not found"})
    def do_POST(self):
        try:
            path=urlparse(self.path).path
            data=self.body()
            if path=="/v1/licenses":
                if not self.admin():return self.reply(401,{"error":"unauthorized"})
                company=data["company"]; modules=data["modules"]; expires=data["expires_at"]
                if not isinstance(company,str) or not 1<=len(company)<=120:raise ValueError("invalid company")
                if not isinstance(modules,list) or not modules or not set(modules)<=MODULES or len(modules)!=len(set(modules)):raise ValueError("invalid modules")
                dt.datetime.fromisoformat(expires.replace("Z","+00:00"))
                max_installations=int(data.get("max_installations",1))
                if not 1<=max_installations<=1000:raise ValueError("invalid max_installations")
                lid="OM-"+secrets.token_hex(10)
                payload={"license_id":lid,"company":company,"modules":sorted(set(modules)|{"core"}),"expires_at":expires,"issued_at":utc(),"max_installations":max_installations,"type":data.get("type","annual"),"schema":1}
                with LOCK,sqlite3.connect(DB) as db:
                    db.execute("INSERT INTO licenses(id,company,payload) VALUES(?,?,?)",(lid,company,json.dumps(payload)))
                return self.reply(201,{"license":token(payload),"public_key":public})
            if path=="/v1/activate":
                doc=data["license"]; installation=data["installation_id"]
                if not isinstance(installation,str) or not re.fullmatch(r"[a-zA-Z0-9_-]{8,100}",installation):raise ValueError("invalid installation_id")
                payload=valid(doc)
                expiry=dt.datetime.fromisoformat(payload["expires_at"].replace("Z","+00:00"))
                if expiry.tzinfo is None:raise ValueError("expiry must include timezone")
                if dt.datetime.now(dt.timezone.utc)>expiry:return self.reply(403,{"error":"license expired"})
                with LOCK,sqlite3.connect(DB) as db:
                    row=db.execute("SELECT revoked FROM licenses WHERE id=?",(payload["license_id"],)).fetchone()
                    if row is None or row[0]:return self.reply(403,{"error":"license unknown or revoked"})
                    count=db.execute("SELECT COUNT(*) FROM activations WHERE license_id=?",(payload["license_id"],)).fetchone()[0]
                    existing=db.execute("SELECT 1 FROM activations WHERE license_id=? AND installation_id=?",(payload["license_id"],installation)).fetchone()
                    if not existing and count>=payload["max_installations"]:return self.reply(403,{"error":"activation limit reached"})
                    db.execute("INSERT OR IGNORE INTO activations VALUES(?,?,?)",(payload["license_id"],installation,utc()))
                return self.reply(200,{"active":True,"entitlements":payload["modules"],"expires_at":payload["expires_at"]})
            if path.startswith("/v1/revoke/"):
                if not self.admin():return self.reply(401,{"error":"unauthorized"})
                lid=path.rsplit("/",1)[-1]
                with sqlite3.connect(DB) as db:
                    changed=db.execute("UPDATE licenses SET revoked=1 WHERE id=?",(lid,)).rowcount
                return self.reply(200 if changed else 404,{"revoked":bool(changed)})
            return self.reply(404,{"error":"not found"})
        except (KeyError,ValueError,TypeError,json.JSONDecodeError) as e:return self.reply(400,{"error":str(e)})
        except Exception:return self.reply(400,{"error":"invalid license or signature"})
ThreadingHTTPServer(("0.0.0.0",8090),Handler).serve_forever()
