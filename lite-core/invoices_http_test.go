package main
import("net/http/httptest";"path/filepath";"strings";"testing")
func TestInvoiceHTTP(t *testing.T){
 dir:=t.TempDir()
 registry,err:=openRegistry(filepath.Join(dir,"companies.json"));if err!=nil{t.Fatal(err)}
 accounts,err:=openAccounting(filepath.Join(dir,"accounts.json"));if err!=nil{t.Fatal(err)}
 invoices,err:=openInvoices(filepath.Join(dir,"invoices.json"));if err!=nil{t.Fatal(err)}
 if err:=registry.upsert(Company{ID:"finance-co",Name:"Finance",Modules:[]string{"finance"}});err!=nil{t.Fatal(err)}
 token:=strings.Repeat("k",32)
 h:=newHandler(config{token:token,registry:registry,accounts:accounts,invoices:invoices})
 call:=func(method,path,body string,auth bool)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if auth{req.Header.Set("Authorization","Bearer "+token)}
  rec:=httptest.NewRecorder();h.ServeHTTP(rec,req);return rec
 }
 path:="/v1/companies/finance-co/invoices"
 body:=`{"id":"invoice-one","date":"2026-11-10","party_name":"Buyer","supply_type":"intra","items":[{"description":"Goods","hsn":"1234","taxable_paise":10000,"rate_basis_points":1800}]}`
 if r:=call("POST",path,body,false);r.Code!=401{t.Fatalf("unauthorized: %d",r.Code)}
 if r:=call("POST",path,body,true);r.Code!=201||!strings.Contains(r.Body.String(),"\"total_paise\":11800"){t.Fatalf("invoice create: %d %s",r.Code,r.Body.String())}
 if r:=call("GET",path,"",true);r.Code!=200||!strings.Contains(r.Body.String(),"invoice-one"){t.Fatalf("invoice list: %d %s",r.Code,r.Body.String())}
 if r:=call("GET","/v1/companies/unknown-co/invoices","",true);r.Code!=404{t.Fatalf("tenant gate: %d",r.Code)}
}
