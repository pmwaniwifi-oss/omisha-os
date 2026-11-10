"""Customer-premises installer: verifies signed entitlements, then renders Compose."""
import argparse, base64, datetime as dt, json, os, pathlib, re, secrets, subprocess, sys
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PublicKey
from cryptography.exceptions import InvalidSignature
MODULES={"core","carbon","manufacturing","finance","ai","iot","crm","compliance","projects"}
def decode(s):return base64.urlsafe_b64decode(s+"="*(-len(s)%4))
def canonical(d):return json.dumps(d,sort_keys=True,separators=(",",":"),ensure_ascii=False).encode()
def verify(document,pubkey):
    Ed25519PublicKey.from_public_bytes(decode(pubkey)).verify(decode(document["signature"]),canonical(document["payload"]))
    p=document["payload"]
    expiry=dt.datetime.fromisoformat(p["expires_at"].replace("Z","+00:00"))
    if expiry.tzinfo is None or expiry<dt.datetime.now(dt.timezone.utc):raise ValueError("license expired or timezone missing")
    if not set(p["modules"])<=MODULES:raise ValueError("unknown module entitlement")
    return p
def main():
    parser=argparse.ArgumentParser()
    parser.add_argument("--license",required=True,help="signed license JSON file")
    parser.add_argument("--public-key",required=True,help="pinned Ed25519 public key file")
    parser.add_argument("--modules",required=True,help="comma-separated module selection")
    parser.add_argument("--customer-id",required=True)
    parser.add_argument("--output",default="./customer-deployment")
    parser.add_argument("--deploy",action="store_true")
    args=parser.parse_args()
    if not re.fullmatch(r"[a-z][a-z0-9-]{2,30}",args.customer_id):parser.error("invalid customer ID")
    try:
        payload=verify(json.loads(pathlib.Path(args.license).read_text()),pathlib.Path(args.public_key).read_text().strip())
    except (ValueError,KeyError,InvalidSignature) as e:parser.error("license verification failed: "+str(e))
    requested=set(filter(None,(x.strip() for x in args.modules.split(","))))|{"core"}
    if not requested<=set(payload["modules"]):parser.error("unlicensed module requested")
    folder=pathlib.Path(args.output).resolve()
    folder.mkdir(parents=True,exist_ok=True,mode=0o700)
    services={}
    for module in sorted(requested):
        service={"build":{"context":"../module-runtime"},"restart":"unless-stopped","environment":{"OMISHA_MODULE":module,"OMISHA_CUSTOMER":args.customer_id},"networks":["tenant"],"read_only":True,"security_opt":["no-new-privileges:true"],"cap_drop":["ALL"],"mem_limit":"128m"}
        if module=="core":service["ports"]=["127.0.0.1:18080:8000"]
        services[module]=service
    # Separate DB and volume for on-prem by default; shared DB mode requires an explicit trusted DB operator.
    password=secrets.token_urlsafe(30)
    services["db"]={"image":"postgres:16-alpine","restart":"unless-stopped","environment":{"POSTGRES_USER":"omisha","POSTGRES_PASSWORD":password,"POSTGRES_DB":"omisha"},"volumes":["pgdata:/var/lib/postgresql/data"],"networks":["tenant"]}
    compose={"name":"omisha-"+args.customer_id,"services":services,"networks":{"tenant":{"internal":True}},"volumes":{"pgdata":{}}}
    # core needs host port, internal network still permits localhost port publishing.
    import yaml
    target=folder/"compose.yml"
    target.write_text(yaml.safe_dump(compose,sort_keys=False))
    target.chmod(0o600)
    (folder/"license.json").write_text(json.dumps({"license_id":payload["license_id"],"modules":sorted(requested),"expires_at":payload["expires_at"]},indent=2))
    print("Verified license:",payload["license_id"],"\nCompose:",target)
    if args.deploy:
        subprocess.run(["docker","compose","-f",str(target),"up","-d","--build"],check=True)
        print("Local dashboard health: http://127.0.0.1:18080/health")
if __name__=="__main__":main()
