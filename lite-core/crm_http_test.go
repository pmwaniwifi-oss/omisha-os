package main

import(
 "net/http/httptest"
 "path/filepath"
 "strings"
 "testing"
)
func TestCRMHTTPModuleGate(t *testing.T){
 dir:=t.TempDir()
 registry,err:=openRegistry(filepath.Join(dir,"companies.json"));if err!=nil{t.Fatal(err)}
 crm,err:=openCRM(filepath.Join(dir,"crm.json"));if err!=nil{t.Fatal(err)}
 if err:=registry.upsert(Company{ID:"first-co",Name:"First",Modules:[]string{"crm"}});err!=nil{t.Fatal(err)}
 if err:=registry.upsert(Company{ID:"second-co",Name:"Second",Modules:[]string{"core"}});err!=nil{t.Fatal(err)}
 token:=strings.Repeat("z",32)
 handler:=newHandler(config{token:token,registry:registry,crm:crm})
 call:=func(method,path,body string,auth bool)*httptest.ResponseRecorder{
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if auth{req.Header.Set("Authorization","Bearer "+token)}
  rec:=httptest.NewRecorder();handler.ServeHTTP(rec,req);return rec
 }
 path:="/v1/companies/first-co/contacts"
 if got:=call("PUT",path,`{"id":"person-one","name":"Alice"}`,false);got.Code!=401{t.Fatalf("expected 401 got %d",got.Code)}
 if got:=call("PUT",path,`{"id":"person-one","name":"Alice"}`,true);got.Code!=200{t.Fatalf("expected 200 got %d %s",got.Code,got.Body.String())}
 if got:=call("GET",path,"",true);got.Code!=200||!strings.Contains(got.Body.String(),"Alice"){t.Fatalf("missing contact: %d %s",got.Code,got.Body.String())}
 if got:=call("GET","/v1/companies/second-co/contacts","",true);got.Code!=404{t.Fatalf("module gate failed: %d",got.Code)}
 if got:=call("GET","/v1/companies/unknown-co/contacts","",true);got.Code!=404{t.Fatalf("company gate failed: %d",got.Code)}
}
