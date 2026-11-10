package main
import("path/filepath";"testing")
func TestTrialBalance(t *testing.T){
 s,err:=openAccounting(filepath.Join(t.TempDir(),"accounts.json"));if err!=nil{t.Fatal(err)}
 j:=Journal{ID:"sale-one",CompanyID:"demo-co",Date:"2026-11-10",Lines:[]JournalLine{
  {Ledger:"cash",DebitPaise:11800},{Ledger:"sales",CreditPaise:10000},{Ledger:"output-igst",CreditPaise:1800},
 }}
 if err:=s.create(j);err!=nil{t.Fatal(err)}
 report,err:=s.trialBalance("demo-co");if err!=nil{t.Fatal(err)}
 if report.TotalDebitPaise!=11800||report.TotalCreditPaise!=11800||len(report.Ledgers)!=3{t.Fatalf("unexpected report: %+v",report)}
 other,err:=s.trialBalance("other-co");if err!=nil{t.Fatal(err)}
 if len(other.Ledgers)!=0{t.Fatal("tenant leakage")}
}
