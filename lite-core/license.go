package main

import (
 "crypto/ed25519"
 "encoding/base64"
 "encoding/json"
 "bytes"
 "errors"
 "os"
 "sort"
 "strings"
 "time"
)

type signedLicense struct {
 Payload struct {
  LicenseID string `json:"license_id"`
  Company string `json:"company"`
  Modules []string `json:"modules"`
  ExpiresAt string `json:"expires_at"`
 } `json:"payload"`
 Signature string `json:"signature"`
}
type Entitlement struct {
 LicenseID string `json:"license_id"`
 Company string `json:"company"`
 Modules []string `json:"modules"`
 ExpiresAt string `json:"expires_at"`
}
func decodeURL(s string)([]byte,error){return base64.RawURLEncoding.DecodeString(strings.TrimRight(s,"="))}
func verifyLicense(licenseJSON []byte, publicKey string, now time.Time)(Entitlement,error){
 var doc signedLicense
 if err:=json.Unmarshal(licenseJSON,&doc);err!=nil{return Entitlement{},err}
 key,err:=decodeURL(strings.TrimSpace(publicKey));if err!=nil{return Entitlement{},err}
 sig,err:=decodeURL(doc.Signature);if err!=nil{return Entitlement{},err}
 if len(key)!=ed25519.PublicKeySize||len(sig)!=ed25519.SignatureSize{return Entitlement{},errors.New("invalid signing key or signature")}
 // The Python licensing server signs a sorted-key compact JSON payload.
 payload:=map[string]any{}
 var outer map[string]json.RawMessage
 if err:=json.Unmarshal(licenseJSON,&outer);err!=nil{return Entitlement{},err}
 if err:=json.Unmarshal(outer["payload"],&payload);err!=nil{return Entitlement{},err}
 var buf bytes.Buffer
 enc:=json.NewEncoder(&buf);enc.SetEscapeHTML(false)
 if err:=enc.Encode(payload);err!=nil{return Entitlement{},err}
 canonical:=bytes.TrimSuffix(buf.Bytes(),[]byte("\\n"))
 // Go encoding/json sorts map keys and emits compact JSON, matching Python's
 // sorted-key canonical encoding for the ASCII fields used by license issuance.
 if !ed25519.Verify(ed25519.PublicKey(key),canonical,sig){return Entitlement{},errors.New("invalid license signature")}
 expires,err:=time.Parse(time.RFC3339,doc.Payload.ExpiresAt)
 if err!=nil{return Entitlement{},errors.New("invalid expiry")}
 if !now.Before(expires){return Entitlement{},errors.New("license expired")}
 if doc.Payload.LicenseID==""||doc.Payload.Company==""{return Entitlement{},errors.New("missing license identity")}
 modules:=map[string]bool{}
 for _,m:=range doc.Payload.Modules{if !allowedModules[m]{return Entitlement{},errors.New("unknown module")};modules[m]=true}
 if !modules["core"]{return Entitlement{},errors.New("core entitlement missing")}
 selected:=make([]string,0,len(modules));for m:=range modules{selected=append(selected,m)};sort.Strings(selected)
 return Entitlement{LicenseID:doc.Payload.LicenseID,Company:doc.Payload.Company,Modules:selected,ExpiresAt:doc.Payload.ExpiresAt},nil
}
func loadLicense(path,keyPath string)(Entitlement,error){
 data,err:=os.ReadFile(path);if err!=nil{return Entitlement{},err}
 key,err:=os.ReadFile(keyPath);if err!=nil{return Entitlement{},err}
 return verifyLicense(data,string(key),time.Now().UTC())
}
