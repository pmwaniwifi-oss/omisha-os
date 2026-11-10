package main
import(
 "crypto/ed25519"
 "crypto/rand"
 "encoding/base64"
 "encoding/json"
 "testing"
 "time"
)
func TestLicenseVerification(t *testing.T){
 pub,priv,err:=ed25519.GenerateKey(rand.Reader);if err!=nil{t.Fatal(err)}
 payload:=map[string]any{"license_id":"OM-123","company":"Test Company","modules":[]string{"core","crm"},"expires_at":time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
 raw,err:=json.Marshal(payload);if err!=nil{t.Fatal(err)}
 signed:=map[string]any{"payload":payload,"signature":base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv,raw))}
 data,_:=json.Marshal(signed)
 key:=base64.RawURLEncoding.EncodeToString(pub)
 got,err:=verifyLicense(data,key,time.Now());if err!=nil{t.Fatal(err)}
 if got.Company!="Test Company"||len(got.Modules)!=2{t.Fatalf("unexpected entitlement: %+v",got)}
 payload["company"]="Tampered"
 tampered,_:=json.Marshal(signed)
 if _,err=verifyLicense(tampered,key,time.Now());err==nil{t.Fatal("tampering accepted")}
}
func TestExpiredLicense(t *testing.T){
 pub,priv,_:=ed25519.GenerateKey(rand.Reader)
 payload:=map[string]any{"license_id":"OM-123","company":"Test","modules":[]string{"core"},"expires_at":"2020-01-01T00:00:00Z"}
 raw,_:=json.Marshal(payload)
 data,_:=json.Marshal(map[string]any{"payload":payload,"signature":base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv,raw))})
 if _,err:=verifyLicense(data,base64.RawURLEncoding.EncodeToString(pub),time.Now());err==nil{t.Fatal("expired license accepted")}
}
