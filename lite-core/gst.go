package main

import ("errors";"math/big")
type GSTBreakup struct {
 TaxablePaise int64 `json:"taxable_paise"`
 RateBasisPoints int64 `json:"rate_basis_points"`
 CGSTPaise int64 `json:"cgst_paise"`
 SGSTPaise int64 `json:"sgst_paise"`
 IGSTPaise int64 `json:"igst_paise"`
 TotalPaise int64 `json:"total_paise"`
}
func gstEstimate(taxablePaise,rateBasisPoints int64,interstate bool)(GSTBreakup,error){
 if taxablePaise<0||taxablePaise>100000000000000||rateBasisPoints<0||rateBasisPoints>10000{return GSTBreakup{},errors.New("invalid taxable amount or GST rate")}
 // Basis points of percent: 1800 means 18%; half-up rounding to paise.
 round:=func(numerator,denominator int64)int64{
  n:=new(big.Int).Mul(big.NewInt(taxablePaise),big.NewInt(numerator))
  n.Add(n,big.NewInt(denominator/2))
  n.Div(n,big.NewInt(denominator))
  return n.Int64()
 }
 out:=GSTBreakup{TaxablePaise:taxablePaise,RateBasisPoints:rateBasisPoints}
 if interstate{out.IGSTPaise=round(rateBasisPoints,10000)}else{
  out.CGSTPaise=round(rateBasisPoints,20000)
  out.SGSTPaise=round(rateBasisPoints,20000)
 }
 out.TotalPaise=taxablePaise+out.CGSTPaise+out.SGSTPaise+out.IGSTPaise
 return out,nil
}
