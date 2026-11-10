package main

import (
 "encoding/json"
 "errors"
 "math"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "sync"
 "time"
)
type JournalLine struct {
 Ledger string `json:"ledger"`
 DebitPaise int64 `json:"debit_paise"`
 CreditPaise int64 `json:"credit_paise"`
}
type Journal struct {
 ID string `json:"id"`
 CompanyID string `json:"company_id"`
 Date string `json:"date"`
 Narration string `json:"narration"`
 Lines []JournalLine `json:"lines"`
}
type AccountingStore struct {
 mu sync.Mutex
 path string
 journals map[string]Journal
}
func openAccounting(path string)(*AccountingStore,error){
 s:=&AccountingStore{path:path,journals:map[string]Journal{}}
 b,err:=os.ReadFile(path)
 if err==nil{if err=json.Unmarshal(b,&s.journals);err!=nil{return nil,err}}
 if err!=nil&&!errors.Is(err,os.ErrNotExist){return nil,err}
 return s,nil
}
func validateJournal(j Journal)error{
 if !idPattern.MatchString(j.ID)||!idPattern.MatchString(j.CompanyID)||len(j.Date)!=10||len(j.Narration)>500||len(j.Lines)<2||len(j.Lines)>100{return errors.New("invalid journal")}
 if parsed,err:=time.Parse("2006-01-02",j.Date);err!=nil||parsed.Format("2006-01-02")!=j.Date{return errors.New("invalid journal date")}
 var debit,credit int64
 for _,l:=range j.Lines{
  if !idPattern.MatchString(l.Ledger)||l.DebitPaise<0||l.CreditPaise<0||(l.DebitPaise>0&&l.CreditPaise>0)||(l.DebitPaise==0&&l.CreditPaise==0){return errors.New("invalid journal line")}
  if l.DebitPaise>math.MaxInt64-debit||l.CreditPaise>math.MaxInt64-credit{return errors.New("amount overflow")}
  debit+=l.DebitPaise;credit+=l.CreditPaise
 }
 if debit==0||debit!=credit{return errors.New("journal is not balanced")}
 return nil
}
func(s *AccountingStore) create(j Journal)error{
 if err:=validateJournal(j);err!=nil{return err}
 s.mu.Lock();defer s.mu.Unlock()
 key:=j.CompanyID+"/"+j.ID
 if _,exists:=s.journals[key];exists{return errors.New("journal already exists; reversal required")}
 s.journals[key]=j
 if err:=s.persist();err!=nil{delete(s.journals,key);return err}
 return nil
}
func(s *AccountingStore) list(company string)[]Journal{
 s.mu.Lock();defer s.mu.Unlock()
 out:=[]Journal{}
 for _,j:=range s.journals{if j.CompanyID==company{out=append(out,j)}}
 sort.Slice(out,func(i,k int)bool{return strings.Compare(out[i].ID,out[k].ID)<0})
 return out
}
func(s *AccountingStore) persist()error{
 dir:=filepath.Dir(s.path)
 if err:=os.MkdirAll(dir,0700);err!=nil{return err}
 f,err:=os.CreateTemp(dir,".accounts-*");if err!=nil{return err}
 defer os.Remove(f.Name())
 if err=f.Chmod(0600);err!=nil{f.Close();return err}
 if err=json.NewEncoder(f).Encode(s.journals);err!=nil{f.Close();return err}
 if err=f.Sync();err!=nil{f.Close();return err}
 if err=f.Close();err!=nil{return err}
 return os.Rename(f.Name(),s.path)
}
