package main

import (
 "encoding/json"
 "errors"
 "os"
 "path/filepath"
 "regexp"
 "sort"
 "sync"
)

type Company struct {
 ID string `json:"id"`
 Name string `json:"name"`
 Modules []string `json:"modules"`
}
type Registry struct {
 mu sync.Mutex
 path string
 companies map[string]Company
}
var idPattern=regexp.MustCompile(`^[a-z][a-z0-9-]{2,40}$`)
var allowedModules=map[string]bool{"core":true,"crm":true,"finance":true,"manufacturing":true,"carbon":true,"compliance":true,"projects":true,"iot":true,"ai":true}
func openRegistry(path string)(*Registry,error){
 s:=&Registry{path:path,companies:map[string]Company{}}
 b,err:=os.ReadFile(path)
 if err==nil {if err=json.Unmarshal(b,&s.companies);err!=nil{return nil,err}}
 if err!=nil&&!errors.Is(err,os.ErrNotExist){return nil,err}
 return s,nil
}
func(s *Registry) list()[]Company{
 s.mu.Lock();defer s.mu.Unlock()
 out:=make([]Company,0,len(s.companies))
 for _,v:=range s.companies{out=append(out,v)}
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 return out
}
func(s *Registry) upsert(c Company)error{
 if !idPattern.MatchString(c.ID)||len(c.Name)<1||len(c.Name)>160{return errors.New("invalid company")}
 modules:=map[string]bool{"core":true}
 for _,m:=range c.Modules{if !allowedModules[m]{return errors.New("invalid module")};modules[m]=true}
 c.Modules=make([]string,0,len(modules))
 for m:=range modules{c.Modules=append(c.Modules,m)}
 sort.Strings(c.Modules)
 s.mu.Lock();defer s.mu.Unlock()
 old,exists:=s.companies[c.ID];s.companies[c.ID]=c
 if err:=s.persist();err!=nil{
  if exists{s.companies[c.ID]=old}else{delete(s.companies,c.ID)}
  return err
 }
 return nil
}
func(s *Registry) persist()error{
 dir:=filepath.Dir(s.path)
 if err:=os.MkdirAll(dir,0700);err!=nil{return err}
 f,err:=os.CreateTemp(dir,".omisha-*")
 if err!=nil{return err}
 defer os.Remove(f.Name())
 if err=f.Chmod(0600);err!=nil{f.Close();return err}
 if err=json.NewEncoder(f).Encode(s.companies);err!=nil{f.Close();return err}
 if err=f.Sync();err!=nil{f.Close();return err}
 if err=f.Close();err!=nil{return err}
 return os.Rename(f.Name(),s.path)
}
