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
type config struct { token string; modules []string; registry *Registry; licensePath string; publicKeyPath string }
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
 return mux
}
func reply(w http.ResponseWriter,code int,v any){
 w.Header().Set("Content-Type","application/json");w.Header().Set("Cache-Control","no-store")
 w.WriteHeader(code);_ = json.NewEncoder(w).Encode(v)
}
func main(){
 token:=os.Getenv("OMISHA_LITE_TOKEN")
 if len(token)<24 {log.Fatal("OMISHA_LITE_TOKEN must be at least 24 characters")}
 mods:=[]string{"core"}
 if raw:=os.Getenv("OMISHA_MODULES");raw!="" {
  seen:=map[string]bool{"core":true}
  for _,v:=range strings.Split(raw,","){v=strings.TrimSpace(v);if !allowedModules[v]{log.Fatalf("unknown module: %s",v)};if !seen[v]{mods=append(mods,v);seen[v]=true}}
 }
 path:=os.Getenv("OMISHA_REGISTRY_PATH");if path==""{path="/data/companies.json"}
 registry,err:=openRegistry(path);if err!=nil{log.Fatal(err)}
 srv:=&http.Server{Addr:":8080",Handler:newHandler(config{token:token,modules:mods,registry:registry,licensePath:os.Getenv("OMISHA_LICENSE_PATH"),publicKeyPath:os.Getenv("OMISHA_LICENSE_PUBLIC_KEY_PATH")}),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second,IdleTimeout:30*time.Second,MaxHeaderBytes:1<<16}
 stop:=make(chan os.Signal,1);signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM)
 go func(){log.Println("Omisha Lite listening on :8080");if err:=srv.ListenAndServe();err!=nil&&err!=http.ErrServerClosed{log.Printf("server error: %v",err)}}()
 <-stop
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
 if err:=srv.Shutdown(ctx);err!=nil{log.Printf("shutdown: %v",err)}
}
