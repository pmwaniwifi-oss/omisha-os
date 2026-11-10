package main

import (
 "encoding/json"
 "errors"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "sync"
)

type Contact struct {
 ID string `json:"id"`
 CompanyID string `json:"company_id"`
 Name string `json:"name"`
 Email string `json:"email,omitempty"`
 Phone string `json:"phone,omitempty"`
}
type CRMStore struct {
 mu sync.Mutex
 path string
 contacts map[string]Contact
}
func openCRM(path string)(*CRMStore,error){
 s:=&CRMStore{path:path,contacts:map[string]Contact{}}
 b,err:=os.ReadFile(path)
 if err==nil {if err=json.Unmarshal(b,&s.contacts);err!=nil{return nil,err}}
 if err!=nil&&!errors.Is(err,os.ErrNotExist){return nil,err}
 return s,nil
}
func(s *CRMStore) list(company string)[]Contact{
 s.mu.Lock();defer s.mu.Unlock()
 out:=[]Contact{}
 for _,c:=range s.contacts{if c.CompanyID==company{out=append(out,c)}}
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 return out
}
func(s *CRMStore) upsert(c Contact)error{
 if !idPattern.MatchString(c.ID)||!idPattern.MatchString(c.CompanyID)||strings.TrimSpace(c.Name)==""||len(c.Name)>160||len(c.Email)>254||len(c.Phone)>40{return errors.New("invalid contact")}
 s.mu.Lock();defer s.mu.Unlock()
 key:=c.CompanyID+"/"+c.ID
 old,exists:=s.contacts[key]
 s.contacts[key]=c
 if err:=s.persist();err!=nil{
  if exists{s.contacts[key]=old}else{delete(s.contacts,key)}
  return err
 }
 return nil
}
func(s *CRMStore) persist()error{
 dir:=filepath.Dir(s.path)
 if err:=os.MkdirAll(dir,0700);err!=nil{return err}
 f,err:=os.CreateTemp(dir,".crm-*");if err!=nil{return err}
 defer os.Remove(f.Name())
 if err=f.Chmod(0600);err!=nil{f.Close();return err}
 if err=json.NewEncoder(f).Encode(s.contacts);err!=nil{f.Close();return err}
 if err=f.Sync();err!=nil{f.Close();return err}
 if err=f.Close();err!=nil{return err}
 return os.Rename(f.Name(),s.path)
}
