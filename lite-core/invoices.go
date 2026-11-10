package main

import (
 "encoding/json"
 "errors"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "sync"
 "time"
)
type InvoiceItem struct {
 Description string `json:"description"`
 HSN string `json:"hsn"`
 TaxablePaise int64 `json:"taxable_paise"`
 RateBasisPoints int64 `json:"rate_basis_points"`
}
type TaxInvoice struct {
 ID string `json:"id"`
 CompanyID string `json:"company_id"`
 Date string `json:"date"`
 PartyName string `json:"party_name"`
 PartyGSTIN string `json:"party_gstin,omitempty"`
 SupplyType string `json:"supply_type"`
 Items []InvoiceItem `json:"items"`
 TaxablePaise int64 `json:"taxable_paise"`
 CGSTPaise int64 `json:"cgst_paise"`
 SGSTPaise int64 `json:"sgst_paise"`
 IGSTPaise int64 `json:"igst_paise"`
 TotalPaise int64 `json:"total_paise"`
}
type InvoiceStore struct {
 mu sync.Mutex
 path string
 invoices map[string]TaxInvoice
}
func openInvoices(path string)(*InvoiceStore,error){
 s:=&InvoiceStore{path:path,invoices:map[string]TaxInvoice{}}
 b,err:=os.ReadFile(path)
 if err==nil{if err=json.Unmarshal(b,&s.invoices);err!=nil{return nil,err}}
 if err!=nil&&!errors.Is(err,os.ErrNotExist){return nil,err}
 return s,nil
}
func calculateInvoice(v TaxInvoice)(TaxInvoice,error){
 if !idPattern.MatchString(v.ID)||!idPattern.MatchString(v.CompanyID)||strings.TrimSpace(v.PartyName)==""||len(v.PartyName)>160||len(v.PartyGSTIN)>15||len(v.Items)<1||len(v.Items)>100{return v,errors.New("invalid invoice")}
 d,err:=time.Parse("2006-01-02",v.Date);if err!=nil||d.Format("2006-01-02")!=v.Date{return v,errors.New("invalid invoice date")}
 if v.SupplyType!="intra"&&v.SupplyType!="inter"{return v,errors.New("supply_type must be intra or inter")}
 v.TaxablePaise=0;v.CGSTPaise=0;v.SGSTPaise=0;v.IGSTPaise=0;v.TotalPaise=0
 for _,item:=range v.Items{
  if strings.TrimSpace(item.Description)==""||len(item.Description)>300||len(item.HSN)>12{return v,errors.New("invalid invoice item")}
  gst,err:=gstEstimate(item.TaxablePaise,item.RateBasisPoints,v.SupplyType=="inter")
  if err!=nil{return v,err}
  // Bound the entire invoice to 1 trillion rupees to prevent int64 overflow.
  if v.TotalPaise>100000000000000-gst.TotalPaise{return v,errors.New("invoice amount too large")}
  v.TaxablePaise+=gst.TaxablePaise
  v.CGSTPaise+=gst.CGSTPaise
  v.SGSTPaise+=gst.SGSTPaise
  v.IGSTPaise+=gst.IGSTPaise
  v.TotalPaise+=gst.TotalPaise
 }
 return v,nil
}
func(s *InvoiceStore) create(v TaxInvoice)(TaxInvoice,error){
 v,err:=calculateInvoice(v);if err!=nil{return v,err}
 s.mu.Lock();defer s.mu.Unlock()
 key:=v.CompanyID+"/"+v.ID
 if _,ok:=s.invoices[key];ok{return v,errors.New("invoice already exists")}
 s.invoices[key]=v
 if err:=s.persist();err!=nil{delete(s.invoices,key);return v,err}
 return v,nil
}
func(s *InvoiceStore) list(company string)[]TaxInvoice{
 s.mu.Lock();defer s.mu.Unlock()
 out:=[]TaxInvoice{}
 for _,v:=range s.invoices{if v.CompanyID==company{out=append(out,v)}}
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 return out
}
func(s *InvoiceStore) persist()error{
 dir:=filepath.Dir(s.path)
 if err:=os.MkdirAll(dir,0700);err!=nil{return err}
 f,err:=os.CreateTemp(dir,".invoices-*");if err!=nil{return err}
 defer os.Remove(f.Name())
 if err=f.Chmod(0600);err!=nil{f.Close();return err}
 if err=json.NewEncoder(f).Encode(s.invoices);err!=nil{f.Close();return err}
 if err=f.Sync();err!=nil{f.Close();return err}
 if err=f.Close();err!=nil{return err}
 return os.Rename(f.Name(),s.path)
}
