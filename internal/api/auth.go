package api

import (
 "encoding/json"
 "net/http"
 "strings"
 "xkeen-panel/internal/auth"
 "xkeen-panel/internal/models"
)

type AuthHandler struct { userManager *auth.UserManager; rateLimiter *RateLimiter; cfg *models.Config }
func NewAuthHandler(um *auth.UserManager,rl *RateLimiter,cfg *models.Config) *AuthHandler { return &AuthHandler{userManager:um,rateLimiter:rl,cfg:cfg} }
func (h *AuthHandler) HandleAuthStatus(w http.ResponseWriter,r *http.Request) {
 writeJSON(w,200,map[string]interface{}{"setup_required":h.userManager.SetupRequired(),"passkey_enabled":false,"totp_required":false})
}
func (h *AuthHandler) HandleSetup(w http.ResponseWriter,r *http.Request) {
 if !h.userManager.SetupRequired() { writeJSON(w,400,map[string]string{"error":"пользователь уже настроен"}); return }
 var req models.SetupRequest
 if json.NewDecoder(http.MaxBytesReader(w,r.Body,4096)).Decode(&req)!=nil { writeJSON(w,400,map[string]string{"error":"неверный формат запроса"}); return }
 req.Username=strings.TrimSpace(req.Username)
 if req.Username=="" || len(req.Username)>100 || len(req.Password)<8 || len(req.Password)>72 { writeJSON(w,400,map[string]string{"error":"нужны логин до 100 байт и пароль от 8 до 72 байт"}); return }
 if h.userManager.CreatePasswordUser(req.Username,req.Password)!=nil { writeJSON(w,400,map[string]string{"error":"не удалось создать пользователя; проверьте состояние настройки"}); return }
 h.passwordToken(w)
}
func (h *AuthHandler) HandleSetupConfirm(w http.ResponseWriter,r *http.Request) {
 writeJSON(w,http.StatusGone,map[string]string{"error":"подтверждение TOTP больше не используется; обновите страницу"})
}
func (h *AuthHandler) passwordToken(w http.ResponseWriter) {
 user:=h.userManager.GetUser(); if user==nil { writeJSON(w,401,map[string]string{"error":"пользователь не настроен"}); return }
 token,err:=auth.GenerateToken(user.Username,user.JWTSecret)
 if err!=nil { writeJSON(w,500,map[string]string{"error":"ошибка генерации токена"}); return }
 writeJSON(w,200,map[string]string{"token":token})
}
func (h *AuthHandler) HandleLogin(w http.ResponseWriter,r *http.Request) {
 var req models.LoginRequest
 if json.NewDecoder(http.MaxBytesReader(w,r.Body,4096)).Decode(&req)!=nil { writeJSON(w,400,map[string]string{"error":"неверный формат запроса"}); return }
 if !h.userManager.CheckPassword(req.Username,req.Password) { writeJSON(w,401,map[string]string{"error":"неверный логин или пароль"}); return }
 h.rateLimiter.Reset(strings.TrimSpace(clientIP(r,h.cfg.TrustProxyHeaders)))
 h.passwordToken(w)
}
func clientIP(r *http.Request,trustProxy bool) string {
 if trustProxy { if forwarded:=r.Header.Get("X-Forwarded-For"); forwarded!="" { parts:=strings.Split(forwarded,","); return strings.TrimSpace(parts[len(parts)-1]) } }
 return r.RemoteAddr
}
func writeJSON(w http.ResponseWriter,status int,data interface{}) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(status); json.NewEncoder(w).Encode(data) }
