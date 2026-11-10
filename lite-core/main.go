package main

import (
 "crypto/subtle"
 "encoding/json"
 "log"
 "net/http"
 "os"
 "os/signal"
 "strings"
 "syscall"
 "context"
 "time"
)

type config struct { token string; modules []string }
func newHandler(c config) http.Handler {
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){reply(w,200,map[string]any{"status":"ok","service":"omisha-lite-core","version":"5.1-dev"})})
 mux.HandleFunc("GET /ready",func(w http.ResponseWriter,r *http.Request){reply(w,200,map[string]any{"ready":true})})
 mux.HandleFunc("GET /v1/modules",func(w http.ResponseWriter,r *http.Request){
  auth:=strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer ")
  if len(auth)!=len(c.token)||subtle.ConstantTimeCompare([]byte(auth),[]byte(c.token))!=1 {reply(w,401,map[string]string{"error":"unauthorized"});return}
  reply(w,200,map[string]any{"modules":c.modules,"mode":"lite","note":"module catalog only; business modules not implemented"})
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
  allowed:=map[string]bool{"core":true,"crm":true,"finance":true,"manufacturing":true,"carbon":true,"compliance":true,"projects":true,"iot":true,"ai":true}
  seen:=map[string]bool{"core":true}
  for _,v:=range strings.Split(raw,","){v=strings.TrimSpace(v);if !allowed[v]{log.Fatalf("unknown module: %s",v)};if !seen[v]{mods=append(mods,v);seen[v]=true}}
 }
 srv:=&http.Server{Addr:":8080",Handler:newHandler(config{token:token,modules:mods}),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second,IdleTimeout:30*time.Second,MaxHeaderBytes:1<<16}
 stop:=make(chan os.Signal,1);signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM)
 go func(){log.Println("Omisha Lite listening on :8080");if err:=srv.ListenAndServe();err!=nil&&err!=http.ErrServerClosed{log.Fatal(err)}}()
 <-stop
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
 if err:=srv.Shutdown(ctx);err!=nil{log.Printf("shutdown: %v",err)}
}
