"""Omisha OS 5.0 control plane: company, module, deployment-plan, billing registry.
Development API only. Requires reverse proxy TLS and production hardening.
"""
import datetime as dt
import hashlib
import hmac
import json
import os
import re
import secrets
import sqlite3
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlsplit

ROOT = Path(os.getenv("OMISHA_DATA_DIR", "/data"))
ROOT.mkdir(parents=True, exist_ok=True)
DB = ROOT / "control.sqlite3"
LOCK = threading.RLock()
MODULES = ("core", "crm", "finance", "manufacturing", "carbon", "compliance", "projects", "iot", "ai")
ROLES = {"owner", "admin", "viewer"}
TOKEN = os.environ.get("OMISHA_CONTROL_ADMIN_TOKEN", "")
if not TOKEN or len(TOKEN) < 24:
    raise SystemExit("OMISHA_CONTROL_ADMIN_TOKEN must contain at least 24 characters")
def now(): return dt.datetime.now(dt.timezone.utc).isoformat()
def connect():
    db = sqlite3.connect(DB, timeout=20)
    db.row_factory = sqlite3.Row
    db.execute("PRAGMA foreign_keys=ON")
    return db
with connect() as db:
    db.executescript("""
    CREATE TABLE IF NOT EXISTS companies (
        id TEXT PRIMARY KEY, name TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'active',
        created_at TEXT NOT NULL);
    CREATE TABLE IF NOT EXISTS company_modules (
        company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
        module TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1,
        PRIMARY KEY(company_id,module));
    CREATE TABLE IF NOT EXISTS deployments (
        id TEXT PRIMARY KEY, company_id TEXT NOT NULL REFERENCES companies(id),
        target TEXT NOT NULL, status TEXT NOT NULL, manifest TEXT NOT NULL, created_at TEXT NOT NULL);
    CREATE TABLE IF NOT EXISTS invoices (
        id TEXT PRIMARY KEY, company_id TEXT NOT NULL REFERENCES companies(id),
        currency TEXT NOT NULL, amount_minor INTEGER NOT NULL,
        status TEXT NOT NULL, created_at TEXT NOT NULL);
    CREATE TABLE IF NOT EXISTS members (
        company_id TEXT NOT NULL REFERENCES companies(id),
        user_id TEXT NOT NULL, role TEXT NOT NULL,
        PRIMARY KEY(company_id,user_id));
    CREATE TABLE IF NOT EXISTS audit (
        id INTEGER PRIMARY KEY AUTOINCREMENT, action TEXT NOT NULL,
        subject TEXT NOT NULL, created_at TEXT NOT NULL);
    """)
def audit(db, action, subject):
    db.execute("INSERT INTO audit(action,subject,created_at) VALUES(?,?,?)", (action,subject,now()))
def slug(value):
    return isinstance(value,str) and bool(re.fullmatch(r"[a-z][a-z0-9-]{2,40}",value))
def company(db,cid):
    return db.execute("SELECT * FROM companies WHERE id=?",(cid,)).fetchone()
class API(BaseHTTPRequestHandler):
    def send(self,code,data):
        body=json.dumps(data,ensure_ascii=False).encode()
        self.send_response(code)
        self.send_header("Content-Type","application/json; charset=utf-8")
        self.send_header("Cache-Control","no-store")
        self.send_header("Content-Length",str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def authorized(self):
        supplied=self.headers.get("Authorization","")
        return hmac.compare_digest(supplied,"Bearer "+TOKEN)
    def read_json(self):
        length=int(self.headers.get("Content-Length","0"))
        if length<=0 or length>32768: raise ValueError("invalid body size")
        obj=json.loads(self.rfile.read(length))
        if not isinstance(obj,dict): raise ValueError("object expected")
        return obj
    def handle_request(self):
        path=urlsplit(self.path).path
        if path=="/health" and self.command=="GET":
            return self.send(200,{"status":"ok","service":"omisha-control-plane","version":"5.0-dev"})
        if not self.authorized():return self.send(401,{"error":"unauthorized"})
        if path=="/v1/catalog" and self.command=="GET":
            return self.send(200,{"modules":list(MODULES),"deployment_targets":["onprem","cloud"],"note":"module services are placeholders until independently implemented"})
        if path=="/v1/companies" and self.command=="GET":
            with connect() as db:
                rows=db.execute("SELECT * FROM companies ORDER BY created_at DESC").fetchall()
                return self.send(200,{"companies":[dict(r) for r in rows]})
        if path=="/v1/companies" and self.command=="POST":
            data=self.read_json()
            cid=data.get("id");name=data.get("name")
            if not slug(cid) or not isinstance(name,str) or not 1<=len(name)<=160:
                return self.send(400,{"error":"invalid company id or name"})
            with LOCK,connect() as db:
                if company(db,cid):return self.send(409,{"error":"company already exists"})
                db.execute("INSERT INTO companies(id,name,created_at) VALUES(?,?,?)",(cid,name,now()))
                db.execute("INSERT INTO company_modules(company_id,module) VALUES(?,?)",(cid,"core"))
                audit(db,"company.created",cid)
            return self.send(201,{"id":cid,"name":name})
        match=re.fullmatch(r"/v1/companies/([a-z][a-z0-9-]{2,40})/modules",path)
        if match:
            cid=match.group(1)
            with LOCK,connect() as db:
                if not company(db,cid):return self.send(404,{"error":"company not found"})
                if self.command=="GET":
                    rows=db.execute("SELECT module,enabled FROM company_modules WHERE company_id=? ORDER BY module",(cid,)).fetchall()
                    return self.send(200,{"company_id":cid,"modules":[dict(r) for r in rows]})
                if self.command=="PUT":
                    data=self.read_json();selected=data.get("modules")
                    if not isinstance(selected,list) or any(not isinstance(m,str) for m in selected) or not set(selected)<=set(MODULES):
                        return self.send(400,{"error":"invalid modules"})
                    selected=set(selected)|{"core"}
                    db.execute("DELETE FROM company_modules WHERE company_id=?",(cid,))
                    db.executemany("INSERT INTO company_modules(company_id,module) VALUES(?,?)",[(cid,m) for m in sorted(selected)])
                    audit(db,"modules.updated",cid)
                    return self.send(200,{"company_id":cid,"modules":sorted(selected)})
        if path=="/v1/deployments" and self.command=="POST":
            data=self.read_json();cid=data.get("company_id");target=data.get("target")
            if target not in ("onprem","cloud"):return self.send(400,{"error":"invalid target"})
            with LOCK,connect() as db:
                if not company(db,cid):return self.send(404,{"error":"company not found"})
                modules=[r["module"] for r in db.execute("SELECT module FROM company_modules WHERE company_id=? AND enabled=1 ORDER BY module",(cid,))]
                did="dep-"+secrets.token_hex(8)
                manifest={"schema":1,"company_id":cid,"target":target,"modules":modules,"status":"planned"}
                db.execute("INSERT INTO deployments VALUES(?,?,?,?,?,?)",(did,cid,target,"planned",json.dumps(manifest),now()))
                audit(db,"deployment.planned",did)
            return self.send(201,{"id":did,"manifest":manifest,"note":"plan only; execution requires a trusted authenticated deployment agent"})
        if path=="/v1/deployments" and self.command=="GET":
            with connect() as db:
                rows=db.execute("SELECT id,company_id,target,status,created_at FROM deployments ORDER BY created_at DESC").fetchall()
                return self.send(200,{"deployments":[dict(r) for r in rows]})
        if path=="/v1/invoices" and self.command=="POST":
            data=self.read_json();cid=data.get("company_id");amount=data.get("amount_minor")
            if type(amount) is not int or amount<0 or amount>10**12:return self.send(400,{"error":"invalid amount_minor"})
            with LOCK,connect() as db:
                if not company(db,cid):return self.send(404,{"error":"company not found"})
                iid="inv-"+secrets.token_hex(8)
                db.execute("INSERT INTO invoices VALUES(?,?,?,?,?,?)",(iid,cid,"INR",amount,"draft",now()))
                audit(db,"invoice.draft",iid)
            return self.send(201,{"id":iid,"status":"draft","amount_minor":amount,"currency":"INR"})
        if path=="/v1/invoices" and self.command=="GET":
            with connect() as db:
                rows=db.execute("SELECT * FROM invoices ORDER BY created_at DESC").fetchall()
                return self.send(200,{"invoices":[dict(r) for r in rows]})
        if path=="/v1/audit" and self.command=="GET":
            with connect() as db:
                rows=db.execute("SELECT * FROM audit ORDER BY id DESC LIMIT 100").fetchall()
                return self.send(200,{"events":[dict(r) for r in rows]})
        return self.send(404,{"error":"not found"})
    def do_GET(self):
        try:self.handle_request()
        except (ValueError,TypeError,json.JSONDecodeError) as e:self.send(400,{"error":str(e)})
        except Exception:self.send(500,{"error":"internal error"})
    def do_POST(self):self.do_GET()
    def do_PUT(self):self.do_GET()
if __name__=="__main__":
    ThreadingHTTPServer(("0.0.0.0",8080),API).serve_forever()
