package main
import("path/filepath";"testing")
func TestCRMIsolationAndPersistence(t *testing.T){
 path:=filepath.Join(t.TempDir(),"crm.json")
 s,err:=openCRM(path);if err!=nil{t.Fatal(err)}
 for _,co:=range []string{"company-one","company-two"}{
  if err=s.upsert(Contact{ID:"contact-one",CompanyID:co,Name:co});err!=nil{t.Fatal(err)}
 }
 s2,err:=openCRM(path);if err!=nil{t.Fatal(err)}
 if len(s2.list("company-one"))!=1||s2.list("company-one")[0].Name!="company-one"{t.Fatal("tenant data mixed")}
 if len(s2.list("company-two"))!=1||s2.list("company-two")[0].Name!="company-two"{t.Fatal("tenant data mixed")}
 if len(s2.list("company-three"))!=0{t.Fatal("unexpected contact")}
}
