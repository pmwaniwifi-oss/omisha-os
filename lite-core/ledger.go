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
type LedgerAccount struct {
 ID string `json:"id"`
 CompanyID string `json:"company_id"`
 Name string `json:"name"`
 Group string `json:"group"`
}
var ledgerGroups=map[string]bool{"asset":true,"liability":true,"equity":true,"income":true,"expense":true}
type LedgerStore struct {
 mu sync.Mutex
 path string
 ledgers map[string]LedgerAccount
}
func openLedgers(path string)(*LedgerStore,error){
 s:=&LedgerStore{path:path,ledgers:map[string]LedgerAccount{}}
 b,err:=os.ReadFile(path)
 if err==nil{if err=json.Unmarshal(b,&s.ledgers);err!=nil{return nil,err}}
 if err!=nil&&!errors.Is(err,os.ErrNotExist){return nil,err}
 return s,nil
}
func(s *LedgerStore) create(v LedgerAccount)error{
 if !idPattern.MatchString(v.ID)||!idPattern.MatchString(v.CompanyID)||strings.TrimSpace(v.Name)==""||len(v.Name)>160||!ledgerGroups[v.Group]{return errors.New("invalid ledger")}
 s.mu.Lock();defer s.mu.Unlock()
 key:=v.CompanyID+"/"+v.ID
 if _,ok:=s.ledgers[key];ok{return errors.New("ledger already exists")}
 s.ledgers[key]=v
 if err:=s.persist();err!=nil{delete(s.ledgers,key);return err}
 return nil
}
func(s *LedgerStore) list(company string)[]LedgerAccount{
 s.mu.Lock();defer s.mu.Unlock()
 out:=[]LedgerAccount{}
 for _,v:=range s.ledgers{if v.CompanyID==company{out=append(out,v)}}
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 return out
}
func(s *LedgerStore) hasAll(company string,lines []JournalLine)bool{
 s.mu.Lock();defer s.mu.Unlock()
 for _,l:=range lines{if _,ok:=s.ledgers[company+"/"+l.Ledger];!ok{return false}}
 return true
}
func(s *LedgerStore) persist()error{
 dir:=filepath.Dir(s.path)
 if err:=os.MkdirAll(dir,0700);err!=nil{return err}
 f,err:=os.CreateTemp(dir,".ledgers-*");if err!=nil{return err}
 defer os.Remove(f.Name())
 if err=f.Chmod(0600);err!=nil{f.Close();return err}
 if err=json.NewEncoder(f).Encode(s.ledgers);err!=nil{f.Close();return err}
 if err=f.Sync();err!=nil{f.Close();return err}
 if err=f.Close();err!=nil{return err}
 return os.Rename(f.Name(),s.path)
}
