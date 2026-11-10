package main

import (
 "crypto/ed25519"
 "crypto/rand"
 "encoding/base64"
 "encoding/json"
 "net/http/httptest"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "time"
)

func TestSignedLicenseCompanyWrite(t *testing.T){
 dir:=t.TempDir()
 pub,priv,err:=ed25519.GenerateKey(rand.Reader);if err!=nil{t.Fatal(err)}
 payload:=map[string]any{
  "license_id":"OM-test","company":"Demo","modules":[]string{"core","crm"},
  "expires_at":time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
 }
 canonical,_:=json.Marshal(payload)
 doc,_:=json.Marshal(map[string]any{"payload":payload,"signature":base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv,canonical))})
 licenseFile:=filepath.Join(dir,"license.json");keyFile:=filepath.Join(dir,"public-key.txt")
 if err:=os.WriteFile(licenseFile,doc,0600);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(keyFile,[]byte(base64.RawURLEncoding.EncodeToString(pub)),0600);err!=nil{t.Fatal(err)}
 reg,err:=openRegistry(filepath.Join(dir,"companies.json"));if err!=nil{t.Fatal(err)}
 token:=strings.Repeat("a",32)
 h:=newHandler(config{token:token,modules:[]string{"core","crm","finance"},registry:reg,licensePath:licenseFile,publicKeyPath:keyFile})
 request:=func(path,body string)*httptest.ResponseRecorder{
  req:=httptest.NewRequest("PUT",path,strings.NewReader(body))
  req.Header.Set("Authorization","Bearer "+token)
  rec:=httptest.NewRecorder();h.ServeHTTP(rec,req);return rec
 }
 if got:=request("/v1/companies",`{"id":"demo-co","name":"Demo","modules":["crm"]}`);got.Code!=200{t.Fatalf("licensed write: %d %s",got.Code,got.Body.String())}
 if got:=request("/v1/companies",`{"id":"other-co","name":"Other","modules":["finance"]}`);got.Code!=403{t.Fatalf("unlicensed write: %d %s",got.Code,got.Body.String())}
 if len(reg.list())!=1{t.Fatalf("unexpected company count %d",len(reg.list()))}
 if err:=os.WriteFile(licenseFile,[]byte("{}"),0600);err!=nil{t.Fatal(err)}
 if got:=request("/v1/companies",`{"id":"demo-co","name":"Changed","modules":["crm"]}`);got.Code!=403{t.Fatalf("tampered license accepted: %d",got.Code)}
}
