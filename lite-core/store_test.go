package main
import("path/filepath";"testing")
func TestRegistryPersists(t *testing.T){
 p:=filepath.Join(t.TempDir(),"registry.json")
 r,err:=openRegistry(p);if err!=nil{t.Fatal(err)}
 if err=r.upsert(Company{ID:"test-co",Name:"Test Co",Modules:[]string{"crm"}});err!=nil{t.Fatal(err)}
 r2,err:=openRegistry(p);if err!=nil{t.Fatal(err)}
 rows:=r2.list()
 if len(rows)!=1||len(rows[0].Modules)!=2||rows[0].Modules[0]!="core"{t.Fatalf("bad persistence: %+v",rows)}
 if err=r2.upsert(Company{ID:"test-co",Name:"Test",Modules:[]string{"unknown"}});err==nil{t.Fatal("invalid module accepted")}
}
