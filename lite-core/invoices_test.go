package main
import("path/filepath";"testing")
func TestInvoiceCalculationAndIsolation(t *testing.T){
 s,err:=openInvoices(filepath.Join(t.TempDir(),"invoices.json"));if err!=nil{t.Fatal(err)}
 inv:=TaxInvoice{ID:"invoice-one",CompanyID:"company-one",Date:"2026-11-10",PartyName:"Buyer",SupplyType:"intra",Items:[]InvoiceItem{{Description:"Goods",HSN:"1234",TaxablePaise:10000,RateBasisPoints:1800}}}
 saved,err:=s.create(inv);if err!=nil{t.Fatal(err)}
 if saved.CGSTPaise!=900||saved.SGSTPaise!=900||saved.TotalPaise!=11800{t.Fatalf("wrong tax: %+v",saved)}
 if _,err:=s.create(inv);err==nil{t.Fatal("duplicate accepted")}
 if len(s.list("company-two"))!=0{t.Fatal("tenant data leak")}
 reloaded,err:=openInvoices(s.path);if err!=nil{t.Fatal(err)}
 if len(reloaded.list("company-one"))!=1{t.Fatal("invoice not persisted")}
 inv.ID="invoice-two";inv.SupplyType="inter"
 saved,err=s.create(inv);if err!=nil||saved.IGSTPaise!=1800{t.Fatalf("IGST failed: %+v %v",saved,err)}
}
