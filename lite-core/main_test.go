package main
import ("net/http/httptest";"testing";"strings")
func TestHealth(t *testing.T){
 req:=httptest.NewRequest("GET","/health",nil);rec:=httptest.NewRecorder()
 newHandler(config{token:strings.Repeat("x",32),modules:[]string{"core"}}).ServeHTTP(rec,req)
 if rec.Code!=200||!strings.Contains(rec.Body.String(),"\"status\":\"ok\""){t.Fatalf("unexpected response: %d %s",rec.Code,rec.Body.String())}
}
func TestAuthorization(t *testing.T){
 handler:=newHandler(config{token:strings.Repeat("x",32),modules:[]string{"core","crm"}})
 for _,tc:=range []struct{auth string;code int}{{"",401},{"Bearer wrong",401},{"Bearer "+strings.Repeat("x",32),200}}{
  req:=httptest.NewRequest("GET","/v1/modules",nil);req.Header.Set("Authorization",tc.auth)
  rec:=httptest.NewRecorder();handler.ServeHTTP(rec,req)
  if rec.Code!=tc.code{t.Fatalf("expected %d got %d",tc.code,rec.Code)}
 }
}
