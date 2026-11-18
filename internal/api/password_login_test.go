package api

import (
 "bytes"
 "net/http/httptest"
 "testing"
 "time"
 "xkeen-panel/internal/auth"
 "xkeen-panel/internal/models"
)
func TestPasswordOnlyLogin(t *testing.T) {
 um:=auth.NewUserManager(t.TempDir())
 if err:=um.CreatePendingUser("owner","test-password","OLD-TOTP-SECRET"); err!=nil { t.Fatal(err) }
 if err:=um.ConfirmSetup(); err!=nil { t.Fatal(err) }
 h:=NewAuthHandler(um,NewRateLimiter(5,time.Minute),&models.Config{})
 for _,test:=range []struct{body string; status int}{
  {`{"username":"owner","password":"test-password"}`,200},
  {`{"username":"owner","password":"wrong"}`,401},
 } {
  w:=httptest.NewRecorder(); r:=httptest.NewRequest("POST","/api/auth/login",bytes.NewBufferString(test.body)); h.HandleLogin(w,r)
  if w.Code!=test.status { t.Fatalf("login status %d, want %d",w.Code,test.status) }
 }
}
func TestPasswordSetupWithoutTOTP(t *testing.T) {
 um:=auth.NewUserManager(t.TempDir()); h:=NewAuthHandler(um,NewRateLimiter(5,time.Minute),&models.Config{})
 w:=httptest.NewRecorder(); h.HandleSetup(w,httptest.NewRequest("POST","/api/auth/setup",bytes.NewBufferString(`{"username":"owner","password":"test-password"}`)))
 if w.Code!=200 || um.SetupRequired() || !um.CheckPassword("owner","test-password") || um.GetUser().TOTPSecret!="" { t.Fatal("password setup failed") }
 if um.CreatePasswordUser("other","replacement")==nil { t.Fatal("existing user overwritten") }
}
