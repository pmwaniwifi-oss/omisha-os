package main
import("path/filepath";"testing")
func TestLedgerStore(t *testing.T){
 s,err:=openLedgers(filepath.Join(t.TempDir(),"ledgers.json"));if err!=nil{t.Fatal(err)}
 l:=LedgerAccount{ID:"cash",CompanyID:"company-one",Name:"Cash",Group:"asset"}
 if err:=s.create(l);err!=nil{t.Fatal(err)}
 if err:=s.create(l);err==nil{t.Fatal("duplicate accepted")}
 if len(s.list("company-two"))!=0{t.Fatal("tenant leak")}
 if !s.hasAll("company-one",[]JournalLine{{Ledger:"cash"}}){t.Fatal("missing ledger")}
 if s.hasAll("company-two",[]JournalLine{{Ledger:"cash"}}){t.Fatal("cross-tenant ledger")}
}
