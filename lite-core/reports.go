package main

import (
 "errors"
 "math"
 "sort"
)
type LedgerBalance struct {
 Ledger string `json:"ledger"`
 DebitPaise int64 `json:"debit_paise"`
 CreditPaise int64 `json:"credit_paise"`
}
type TrialBalance struct {
 CompanyID string `json:"company_id"`
 Ledgers []LedgerBalance `json:"ledgers"`
 TotalDebitPaise int64 `json:"total_debit_paise"`
 TotalCreditPaise int64 `json:"total_credit_paise"`
}
func(s *AccountingStore) trialBalance(company string)(TrialBalance,error){
 result:=TrialBalance{CompanyID:company,Ledgers:[]LedgerBalance{}}
 net:=map[string]int64{}
 for _,j:=range s.list(company){
  for _,line:=range j.Lines{
   delta:=line.DebitPaise-line.CreditPaise
   prior:=net[line.Ledger]
   if delta>0&&prior>math.MaxInt64-delta{return TrialBalance{},errors.New("balance overflow")}
   if delta<0&&prior<math.MinInt64-delta{return TrialBalance{},errors.New("balance overflow")}
   net[line.Ledger]=prior+delta
  }
 }
 names:=make([]string,0,len(net))
 for name:=range net{names=append(names,name)}
 sort.Strings(names)
 for _,name:=range names{
  value:=net[name]
  row:=LedgerBalance{Ledger:name}
  if value>0{row.DebitPaise=value}else if value<0{row.CreditPaise= -value}
  if row.DebitPaise>math.MaxInt64-result.TotalDebitPaise||row.CreditPaise>math.MaxInt64-result.TotalCreditPaise{return TrialBalance{},errors.New("total overflow")}
  result.TotalDebitPaise+=row.DebitPaise
  result.TotalCreditPaise+=row.CreditPaise
  result.Ledgers=append(result.Ledgers,row)
 }
 if result.TotalDebitPaise!=result.TotalCreditPaise{return TrialBalance{},errors.New("unbalanced trial balance")}
 return result,nil
}
