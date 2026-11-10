package main

import (
 "net/http/httptest"
 "path/filepath"
 "strings"
 "testing"
)
func TestAccountingHTTP(t *testing.T){
 dir:=t.TempDir()
 registry,err:=openRegistry(filepath.Join(dir,"companies.json"));if err!=nil{t.Fatal(err)}
 accounts,err:=openAccounting(filepath.Join(dir,"accounts.json"));if err!=nil{t.Fatal(err)}
 if err:=registry.upsert(Company{ID:"finance-co",Name:"Finance",Modules:[]string{"finance"}});err!=nil{t.Fatal(err)}
 if err:=registry.upsert(Company{ID:"crm-only",Name:"CRM",Modules:[]string{"crm"}});err!=nil{t.Fatal(err)}
 ledgers,err:=openLedgers(filepath.Join(dir,"ledgers.json"));if err!=nil{t.Fatal(err)}
 for _,id:=range []string{"cash","sales","output-igst"}{if err:=ledgers.create(LedgerAccount{ID:id,CompanyID:"finance-co",Name:id,Group:"asset"});err!=nil{t.Fatal(err)}}
 token:=strings.Repeat("s",32)
 h:=newHandler(config{token:token,registry:registry,accounts:accounts,ledgers:ledgers})
 call:=func(method,path,body string,auth bool)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if auth{req.Header.Set("Authorization","Bearer "+token)}
  rec:=httptest.NewRecorder();h.ServeHTTP(rec,req);return rec
 }
 path:="/v1/companies/finance-co/journals"
 body:=`{"id":"entry-one","date":"2026-11-10","lines":[{"ledger":"cash","debit_paise":11800},{"ledger":"sales","credit_paise":10000},{"ledger":"output-igst","credit_paise":1800}]}`
 if r:=call("POST",path,body,false);r.Code!=401{t.Fatalf("expected 401 got %d",r.Code)}
 if r:=call("POST",path,body,true);r.Code!=201{t.Fatalf("expected 201 got %d: %s",r.Code,r.Body.String())}
 if r:=call("POST",path,body,true);r.Code!=400{t.Fatalf("duplicate accepted: %d",r.Code)}
 if r:=call("GET","/v1/companies/finance-co/trial-balance","",true);r.Code!=200||!strings.Contains(r.Body.String(),"\"total_debit_paise\":11800"){t.Fatalf("bad trial balance: %d %s",r.Code,r.Body.String())}
 if r:=call("GET","/v1/companies/crm-only/journals","",true);r.Code!=404{t.Fatalf("finance module gate failed: %d",r.Code)}
}
