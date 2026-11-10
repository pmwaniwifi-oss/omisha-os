package main
import("path/filepath";"testing")
func TestAccountingBalancedAndTenantScoped(t *testing.T){
 s,err:=openAccounting(filepath.Join(t.TempDir(),"journals.json"));if err!=nil{t.Fatal(err)}
 j:=Journal{ID:"entry-one",CompanyID:"company-one",Date:"2026-11-10",Lines:[]JournalLine{{Ledger:"cash",DebitPaise:11800},{Ledger:"sales",CreditPaise:10000},{Ledger:"output-igst",CreditPaise:1800}}}
 if err:=s.create(j);err!=nil{t.Fatal(err)}
 if len(s.list("company-two"))!=0{t.Fatal("tenant leak")}
 if err:=s.create(j);err==nil{t.Fatal("duplicate accepted")}
 j.ID="entry-two";j.Lines[2].CreditPaise=1700
 if err:=s.create(j);err==nil{t.Fatal("unbalanced journal accepted")}
}
func TestGSTEstimate(t *testing.T){
 intra,err:=gstEstimate(10000,1800,false);if err!=nil{t.Fatal(err)}
 if intra.CGSTPaise!=900||intra.SGSTPaise!=900||intra.TotalPaise!=11800{t.Fatalf("bad intra GST: %+v",intra)}
 inter,err:=gstEstimate(10000,1800,true);if err!=nil{t.Fatal(err)}
 if inter.IGSTPaise!=1800||inter.TotalPaise!=11800{t.Fatalf("bad inter GST: %+v",inter)}
}
