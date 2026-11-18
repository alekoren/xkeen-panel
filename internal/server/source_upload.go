package server

import (
 "bytes"
 "encoding/json"
 "io"
 "net/http"
)

// upload normalizes WireGuard files before the existing staged-source validator.
func (m *sourceManager) upload(w http.ResponseWriter,r *http.Request) {
 var req struct { Name string `json:"name"`; Kind string `json:"kind"`; Payload string `json:"payload"` }
 dec:=json.NewDecoder(http.MaxBytesReader(w,r.Body,2<<20)); dec.DisallowUnknownFields()
 if dec.Decode(&req)!=nil { coreJSON(w,400,map[string]string{"error":"Неверный запрос или превышен размер"}); return }
 var extra any; if dec.Decode(&extra)!=io.EOF { coreJSON(w,400,map[string]string{"error":"Лишние данные"}); return }
 if req.Kind=="wireguard-conf" {
  if len(req.Payload)>1<<20 { coreJSON(w,400,map[string]string{"error":"Файл больше 1 МиБ"}); return }
  text,err:=normalizeWG(req.Name,req.Payload); if err!=nil { coreJSON(w,400,map[string]string{"error":err.Error()}); return }; req.Payload=text; req.Kind="mihomo-yaml"
 }
 body,err:=json.Marshal(req); if err!=nil { coreJSON(w,400,map[string]string{"error":"Ошибка импорта"}); return }
 clone:=r.Clone(r.Context()); clone.Body=io.NopCloser(bytes.NewReader(body)); clone.ContentLength=int64(len(body)); m.create(w,clone)
}
