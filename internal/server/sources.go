package server

import (
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"

 "gopkg.in/yaml.v3"
)

type sourceEntry struct {
 ID string `json:"id"`
 Name string `json:"name"`
 Kind string `json:"kind"`
 Created string `json:"created"`
 Status string `json:"status"`
 Payload string `json:"payload,omitempty"`
}

type sourceManager struct { dir string; mu sync.Mutex }

func (m *sourceManager) list(w http.ResponseWriter, r *http.Request) {
 m.mu.Lock(); defer m.mu.Unlock()
 entries, err := os.ReadDir(m.dir)
 if err != nil && !os.IsNotExist(err) { coreJSON(w, 500, map[string]string{"error":"Не удалось прочитать источники"}); return }
 out := []sourceEntry{}
 for _, file := range entries {
  if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") { continue }
  f, err := os.Open(filepath.Join(m.dir, file.Name())); if err != nil { coreJSON(w, 500, map[string]string{"error":"Ошибка чтения источника"}); return }
  var entry sourceEntry
  err = json.NewDecoder(io.LimitReader(f, 2<<20)).Decode(&entry); f.Close()
  if err != nil { coreJSON(w, 500, map[string]string{"error":"Повреждён каталог источников"}); return }
  entry.Payload = ""
  out = append(out, entry)
 }
 coreJSON(w, 200, out)
}

func validateSource(kind, payload string) error {
 if kind == "subscription" {
  u, err := url.Parse(payload)
  if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" { return fmt.Errorf("Требуется HTTPS-ссылка без логина и фрагмента") }
  return nil
 }
 if kind != "mihomo-yaml" { return fmt.Errorf("Поддерживаются HTTPS-ссылки и YAML Mihomo; импорт WireGuard .conf ещё не готов") }
 var doc yaml.Node
 decoder := yaml.NewDecoder(strings.NewReader(payload))
 if decoder.Decode(&doc) != nil || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode { return fmt.Errorf("Нужен YAML-объект Mihomo с непустым списком proxies") }
 var extra yaml.Node
 if decoder.Decode(&extra) != io.EOF { return fmt.Errorf("Допускается один YAML-документ") }
 // A typed decode also rejects duplicate mapping keys. No aliases are expanded here.
 var mapping map[string]json.RawMessage
 _ = mapping
 var root map[string]any
 if yaml.Unmarshal([]byte(payload), &root) != nil { return fmt.Errorf("Некорректный YAML или повторяющиеся ключи") }
 proxies, ok := root["proxies"].([]any)
 if !ok || len(proxies) == 0 || len(proxies) > 2000 { return fmt.Errorf("Требуется от 1 до 2000 узлов в proxies; это не проверка совместимости ядром") }
 for _, raw := range proxies {
  proxy, ok := raw.(map[string]any); if !ok { return fmt.Errorf("Некорректный узел proxies") }
  name, nOK := proxy["name"].(string); typ, tOK := proxy["type"].(string)
  if !nOK || !tOK || strings.TrimSpace(name) == "" || strings.TrimSpace(typ) == "" { return fmt.Errorf("У каждого узла нужны name и type") }
 }
 return nil
}

func (m *sourceManager) create(w http.ResponseWriter, r *http.Request) {
 var req struct { Name string `json:"name"`; Kind string `json:"kind"`; Payload string `json:"payload"` }
 dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)); dec.DisallowUnknownFields()
 if dec.Decode(&req) != nil { coreJSON(w, 400, map[string]string{"error":"Неверный запрос или превышен размер"}); return }
 var extra any
 if dec.Decode(&extra) != io.EOF { coreJSON(w, 400, map[string]string{"error":"Лишние данные в запросе"}); return }
 req.Name = strings.TrimSpace(req.Name)
 if req.Name == "" || len(req.Name) > 160 || len(req.Payload) > 1<<20 { coreJSON(w, 400, map[string]string{"error":"Нужно имя до 160 байт и файл до 1 МиБ"}); return }
 if err := validateSource(req.Kind, req.Payload); err != nil { coreJSON(w, 400, map[string]string{"error":err.Error()}); return }
 m.mu.Lock(); defer m.mu.Unlock()
 if err := os.MkdirAll(m.dir, 0700); err != nil { coreJSON(w, 500, map[string]string{"error":"Не удалось создать каталог"}); return }
 var id [16]byte
 if _, err := rand.Read(id[:]); err != nil { coreJSON(w, 500, map[string]string{"error":"Не удалось создать ID"}); return }
 entry := sourceEntry{ID:hex.EncodeToString(id[:]), Name:req.Name, Kind:req.Kind, Created:time.Now().UTC().Format(time.RFC3339), Status:"stored-not-applied", Payload:req.Payload}
 data, err := json.Marshal(entry)
 if err != nil { coreJSON(w, 500, map[string]string{"error":"Ошибка сохранения"}); return }
 f, err := os.CreateTemp(m.dir, ".source-*")
 if err != nil { coreJSON(w, 500, map[string]string{"error":"Ошибка сохранения"}); return }
 tmp := f.Name(); defer os.Remove(tmp)
 if _, err = f.Write(data); err == nil { err = f.Sync() }
 closeErr := f.Close(); if err == nil { err = closeErr }
 if err == nil { err = os.Rename(tmp, filepath.Join(m.dir, entry.ID+".json")) }
 if err != nil { coreJSON(w, 500, map[string]string{"error":"Ошибка сохранения"}); return }
 entry.Payload = ""
 coreJSON(w, 201, entry)
}
