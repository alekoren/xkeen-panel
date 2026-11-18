package server

import (
 "encoding/base64"
 "strings"
 "testing"
 "gopkg.in/yaml.v3"
)
func TestNormalizeWG(t *testing.T) {
 key:=base64.StdEncoding.EncodeToString(make([]byte,32))
 conf:="[Interface]\nPrivateKey = "+key+"\nAddress = 10.2.0.2/32, fd00::2/128\nDNS = 10.2.0.1\n[Peer]\nPublicKey = "+key+"\nAllowedIPs = 0.0.0.0/0, ::/0\nEndpoint = [2001:db8::1]:51820\nPersistentKeepalive = 25\n"
 result,err:=normalizeWG("test",conf); if err!=nil { t.Fatal(err) }
 var root struct { Proxies []map[string]any `yaml:"proxies"` }; if yaml.Unmarshal([]byte(result),&root)!=nil || len(root.Proxies)!=1 { t.Fatal("invalid YAML") }
 node:=root.Proxies[0]; if node["server"]!="2001:db8::1" || node["ip"]!="10.2.0.2" || node["type"]!="wireguard" { t.Fatal("incorrect mapping") }
 for _,bad:=range []string{conf+"[Peer]\n",strings.Replace(conf,"DNS = 10.2.0.1","PostUp = echo unsafe",1),strings.Replace(conf,"DNS = 10.2.0.1","Jc = 5",1),strings.Replace(conf,":51820",":0",1),strings.Replace(conf,"PrivateKey = "+key,"PrivateKey = bad",1)} { if _,err:=normalizeWG("test",bad); err==nil { t.Fatal("invalid config accepted") } }
}
