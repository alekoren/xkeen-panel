package server

import (
 "encoding/base64"
 "fmt"
 "net"
 "net/netip"
 "strconv"
 "strings"
 "gopkg.in/yaml.v3"
)

// normalizeWG accepts one ordinary WireGuard peer, never shell hooks or AWG.
func normalizeWG(name, text string) (string, error) {
 fail := func() (string,error) { return "",fmt.Errorf("Неподдерживаемый или некорректный WireGuard .conf; нужен один Peer, без команд и параметров AmneziaWG") }
 fields:=map[string]string{}; section:=""; seen:=map[string]bool{}
 for _,line:=range strings.Split(strings.TrimPrefix(text,"\ufeff"),"\n") {
  line=strings.TrimSpace(strings.SplitN(line,"#",2)[0]); if line=="" { continue }
  if strings.HasPrefix(line,"[") {
   if (line!="[Interface]" && line!="[Peer]") || seen[line] { return fail() }; seen[line]=true; section=line; continue
  }
  key,value,ok:=strings.Cut(line,"="); key=strings.TrimSpace(key); value=strings.TrimSpace(value)
  if !ok || value=="" || section=="" { return fail() }
  allowed:="|PrivateKey|Address|DNS|MTU|"; if section=="[Peer]" { allowed="|PublicKey|PresharedKey|AllowedIPs|Endpoint|PersistentKeepalive|" }
  field:=section+key
  if !strings.Contains(allowed,"|"+key+"|") || fields[field]!="" { return fail() }; fields[field]=value
 }
 i:=func(key string) string { return fields["[Interface]"+key] }; p:=func(key string) string { return fields["[Peer]"+key] }
 validKey:=func(s string) bool { b,e:=base64.StdEncoding.DecodeString(s); return e==nil && len(b)==32 }
 if !seen["[Interface]"] || !seen["[Peer]"] || !validKey(i("PrivateKey")) || !validKey(p("PublicKey")) { return fail() }
 host,portText,err:=net.SplitHostPort(p("Endpoint")); port,e:=strconv.Atoi(portText)
 if err!=nil || e!=nil || host=="" || port<1 || port>65535 || strings.ContainsAny(host," /\t\r\n") { return fail() }
 proxy:=map[string]any{"name":name,"type":"wireguard","server":host,"port":port,"private-key":i("PrivateKey"),"public-key":p("PublicKey"),"udp":true}
 for _,address:=range strings.Split(i("Address"),",") {
  prefix,err:=netip.ParsePrefix(strings.TrimSpace(address)); if err!=nil { return fail() }
  key:="ip"; if prefix.Addr().Is6() { key="ipv6" }; if _,exists:=proxy[key]; exists { return fail() }; proxy[key]=prefix.Addr().String()
 }
 var ranges []string
 for _,cidr:=range strings.Split(p("AllowedIPs"),",") { prefix,err:=netip.ParsePrefix(strings.TrimSpace(cidr)); if err!=nil { return fail() }; ranges=append(ranges,prefix.Masked().String()) }; proxy["allowed-ips"]=ranges
 if p("PresharedKey")!="" { if !validKey(p("PresharedKey")) { return fail() }; proxy["pre-shared-key"]=p("PresharedKey") }
 if i("MTU")!="" { mtu,err:=strconv.Atoi(i("MTU")); if err!=nil || mtu<576 || mtu>65535 { return fail() }; proxy["mtu"]=mtu }
 if p("PersistentKeepalive")!="" { n,err:=strconv.Atoi(p("PersistentKeepalive")); if err!=nil || n<0 || n>65535 { return fail() }; proxy["persistent-keepalive"]=n }
 if i("DNS")!="" {
  var dns []string; for _,value:=range strings.Split(i("DNS"),",") { addr,err:=netip.ParseAddr(strings.TrimSpace(value)); if err!=nil { return fail() }; dns=append(dns,addr.String()) }; proxy["dns"]=dns; proxy["remote-dns-resolve"]=true
 }
 data,err:=yaml.Marshal(map[string]any{"proxies":[]any{proxy}}); return string(data),err
}
