package main

import (
 "context"
 "crypto/subtle"
 "encoding/json"
 "io"
 "log"
 "net/http"
 "os"
 "os/signal"
 "strings"
 "syscall"
 "time"
)
type config struct { token string; modules []string; registry *Registry; crm *CRMStore; accounts *AccountingStore; licensePath string; publicKeyPath string }
func newHandler(c config) http.Handler {
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){reply(w,200,map[string]any{"status":"ok","service":"omisha-lite-core","version":"5.1-dev"})})
 mux.HandleFunc("GET /ready",func(w http.ResponseWriter,r *http.Request){reply(w,200,map[string]any{"ready":true})})
 authenticated:=func(w http.ResponseWriter,r *http.Request)bool{
  header:=r.Header.Get("Authorization")
  expected:="Bearer "+c.token
  if len(header)!=len(expected)||subtle.ConstantTimeCompare([]byte(header),[]byte(expected))!=1{
   reply(w,401,map[string]string{"error":"unauthorized"});return false
  }
  if c.licensePath!="" {
   if c.publicKeyPath==""{reply(w,503,map[string]string{"error":"license public key missing"});return false}
   if _,err:=loadLicense(c.licensePath,c.publicKeyPath);err!=nil{reply(w,403,map[string]string{"error":"license invalid or expired"});return false}
  }
  return true
 }
 mux.HandleFunc("GET /v1/modules",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  modules:=c.modules
  if c.licensePath!=""{
   ent,err:=loadLicense(c.licensePath,c.publicKeyPath)
   if err!=nil{reply(w,403,map[string]string{"error":"license invalid or expired"});return}
   allowed:=map[string]bool{};for _,m:=range ent.Modules{allowed[m]=true}
   modules=[]string{};for _,m:=range c.modules{if allowed[m]{modules=append(modules,m)}}
  }
  reply(w,200,map[string]any{"modules":modules,"mode":"lite","note":"module catalog only; business modules not implemented"})
 })
 mux.HandleFunc("GET /v1/companies",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  if c.registry==nil{reply(w,503,map[string]string{"error":"registry unavailable"});return}
  reply(w,200,map[string]any{"companies":c.registry.list()})
 })
 mux.HandleFunc("PUT /v1/companies",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  if c.registry==nil{reply(w,503,map[string]string{"error":"registry unavailable"});return}
  r.Body=http.MaxBytesReader(w,r.Body,16384)
  var company Company
  dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
  if err:=dec.Decode(&company);err!=nil{reply(w,400,map[string]string{"error":"invalid JSON"});return}
  var extra any
  if err:=dec.Decode(&extra);err!=io.EOF{reply(w,400,map[string]string{"error":"trailing JSON"});return}
  if c.licensePath!=""{
   ent,err:=loadLicense(c.licensePath,c.publicKeyPath)
   if err!=nil{reply(w,403,map[string]string{"error":"license invalid or expired"});return}
   allowed:=map[string]bool{};for _,m:=range ent.Modules{allowed[m]=true}
   for _,m:=range company.Modules{if !allowed[m]{reply(w,403,map[string]string{"error":"module not licensed"});return}}
  }
  if err:=c.registry.upsert(company);err!=nil{reply(w,400,map[string]string{"error":err.Error()});return}
  reply(w,200,map[string]any{"company":company.ID,"status":"saved"})
 })
 mux.HandleFunc("GET /v1/companies/{company}/contacts",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  company:=r.PathValue("company")
  if !crmAccess(w,c,company){return}
  reply(w,200,map[string]any{"contacts":c.crm.list(company)})
 })
 mux.HandleFunc("PUT /v1/companies/{company}/contacts",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  company:=r.PathValue("company")
  if !crmAccess(w,c,company){return}
  r.Body=http.MaxBytesReader(w,r.Body,16384)
  var contact Contact
  dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
  if err:=dec.Decode(&contact);err!=nil{reply(w,400,map[string]string{"error":"invalid JSON"});return}
  var extra any
  if err:=dec.Decode(&extra);err!=io.EOF{reply(w,400,map[string]string{"error":"trailing JSON"});return}
  contact.CompanyID=company
  if err:=c.crm.upsert(contact);err!=nil{reply(w,400,map[string]string{"error":err.Error()});return}
  reply(w,200,map[string]string{"status":"saved","id":contact.ID})
 })
 mux.HandleFunc("POST /v1/companies/{company}/journals",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  company:=r.PathValue("company")
  if !financeAccess(w,c,company){return}
  r.Body=http.MaxBytesReader(w,r.Body,65536)
  var j Journal
  dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
  if err:=dec.Decode(&j);err!=nil{reply(w,400,map[string]string{"error":"invalid JSON"});return}
  var extra any
  if err:=dec.Decode(&extra);err!=io.EOF{reply(w,400,map[string]string{"error":"trailing JSON"});return}
  j.CompanyID=company
  if err:=c.accounts.create(j);err!=nil{reply(w,400,map[string]string{"error":err.Error()});return}
  reply(w,201,map[string]string{"status":"posted","id":j.ID})
 })
 mux.HandleFunc("GET /v1/companies/{company}/journals",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  company:=r.PathValue("company")
  if !financeAccess(w,c,company){return}
  reply(w,200,map[string]any{"journals":c.accounts.list(company)})
 })
 mux.HandleFunc("GET /v1/companies/{company}/trial-balance",func(w http.ResponseWriter,r *http.Request){
  if !authenticated(w,r){return}
  company:=r.PathValue("company")
  if !financeAccess(w,c,company){return}
  result,err:=c.accounts.trialBalance(company)
  if err!=nil{reply(w,500,map[string]string{"error":"trial balance calculation failed"});return}
  reply(w,200,result)
 })
 return mux
}
func financeAccess(w http.ResponseWriter,c config,company string)bool{
 if c.registry==nil||c.accounts==nil{reply(w,503,map[string]string{"error":"accounting store unavailable"});return false}
 if !idPattern.MatchString(company){reply(w,400,map[string]string{"error":"invalid company id"});return false}
 enabled:=false
 for _,v:=range c.registry.list(){if v.ID==company{for _,m:=range v.Modules{if m=="finance"{enabled=true}}}}
 if !enabled{reply(w,404,map[string]string{"error":"company or finance module not found"});return false}
 if c.licensePath!=""{
  ent,err:=loadLicense(c.licensePath,c.publicKeyPath)
  if err!=nil{reply(w,403,map[string]string{"error":"license invalid or expired"});return false}
  licensed:=false;for _,m:=range ent.Modules{if m=="finance"{licensed=true}}
  if !licensed{reply(w,403,map[string]string{"error":"finance not licensed"});return false}
 }
 return true
}
func crmAccess(w http.ResponseWriter,c config,company string)bool{
 if c.registry==nil||c.crm==nil{reply(w,503,map[string]string{"error":"store unavailable"});return false}
 if !idPattern.MatchString(company){reply(w,400,map[string]string{"error":"invalid company id"});return false}
 enabled:=false
 for _,v:=range c.registry.list(){if v.ID==company{for _,m:=range v.Modules{if m=="crm"{enabled=true}}}}
 if !enabled{reply(w,404,map[string]string{"error":"company or CRM module not found"});return false}
 if c.licensePath!=""{
  ent,err:=loadLicense(c.licensePath,c.publicKeyPath)
  if err!=nil{reply(w,403,map[string]string{"error":"license invalid or expired"});return false}
  licensed:=false;for _,m:=range ent.Modules{if m=="crm"{licensed=true}}
  if !licensed{reply(w,403,map[string]string{"error":"CRM not licensed"});return false}
 }
 return true
}
func reply(w http.ResponseWriter,code int,v any){
 w.Header().Set("Content-Type","application/json");w.Header().Set("Cache-Control","no-store")
 w.WriteHeader(code);_ = json.NewEncoder(w).Encode(v)
}
func main(){
 token:=os.Getenv("OMISHA_LITE_TOKEN")
 if os.Getenv("OMISHA_DEV_MODE")!="true" && (os.Getenv("OMISHA_LICENSE_PATH")=="" || os.Getenv("OMISHA_LICENSE_PUBLIC_KEY_PATH")==""){log.Fatal("production requires license and pinned public key; set OMISHA_DEV_MODE=true only for local development")}
 if len(token)<24 {log.Fatal("OMISHA_LITE_TOKEN must be at least 24 characters")}
 mods:=[]string{"core"}
 if raw:=os.Getenv("OMISHA_MODULES");raw!="" {
  seen:=map[string]bool{"core":true}
  for _,v:=range strings.Split(raw,","){v=strings.TrimSpace(v);if !allowedModules[v]{log.Fatalf("unknown module: %s",v)};if !seen[v]{mods=append(mods,v);seen[v]=true}}
 }
 path:=os.Getenv("OMISHA_REGISTRY_PATH");if path==""{path="/data/companies.json"}
 registry,err:=openRegistry(path);if err!=nil{log.Fatal(err)}
 crmPath:=os.Getenv("OMISHA_CRM_PATH");if crmPath==""{crmPath="/data/crm.json"}
 crm,err:=openCRM(crmPath);if err!=nil{log.Fatal(err)}
 accountingPath:=os.Getenv("OMISHA_ACCOUNTING_PATH");if accountingPath==""{accountingPath="/data/accounting.json"}
 accounts,err:=openAccounting(accountingPath);if err!=nil{log.Fatal(err)}
 srv:=&http.Server{Addr:":8080",Handler:newHandler(config{token:token,modules:mods,registry:registry,crm:crm,accounts:accounts,licensePath:os.Getenv("OMISHA_LICENSE_PATH"),publicKeyPath:os.Getenv("OMISHA_LICENSE_PUBLIC_KEY_PATH")}),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second,IdleTimeout:30*time.Second,MaxHeaderBytes:1<<16}
 stop:=make(chan os.Signal,1);signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM)
 go func(){log.Println("Omisha Lite listening on :8080");if err:=srv.ListenAndServe();err!=nil&&err!=http.ErrServerClosed{log.Printf("server error: %v",err)}}()
 <-stop
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
 if err:=srv.Shutdown(ctx);err!=nil{log.Printf("shutdown: %v",err)}
}
