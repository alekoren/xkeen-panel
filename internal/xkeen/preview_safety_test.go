package xkeen

import (
 "os"
 "path/filepath"
 "testing"
)
func TestCoreCommandExactExecutable(t *testing.T) {
 for _,test:=range []struct{cmd,core string; want bool}{
  {"/opt/sbin/mihomo\x00-d\x00/opt/etc/mihomo\x00",CoreMihomo,true},
  {"/opt/sbin/xray\x00run\x00-confdir\x00/opt/etc/xray/configs\x00",CoreXray,true},
  {"/bin/sh\x00/opt/sbin/xkeen\x00-mihomo\x00",CoreMihomo,false},
  {"/opt/sbin/xkeen-panel\x00-config\x00mihomo.yaml\x00",CoreMihomo,false},
  {"/opt/sbin/mihomo\x00-t\x00",CoreMihomo,false},
  {"/opt/sbin/xray\x00run\x00-test\x00",CoreXray,false},
 } { if got:=coreCommand([]byte(test.cmd),test.core);got!=test.want { t.Errorf("core=%s got=%v want=%v",test.core,got,test.want) } }
}
func TestLegacyApplyCannotWriteDuringMihomo(t *testing.T) {
 path:=filepath.Join(t.TempDir(),"outbounds.json")
 if err:=os.WriteFile(path,[]byte("unchanged"),0600);err!=nil { t.Fatal(err) }
 if ApplyServer(Runtime{Core:CoreMihomo},path,nil)==nil { t.Fatal("legacy write was allowed") }
 data,err:=os.ReadFile(path);if err!=nil || string(data)!="unchanged" { t.Fatal("legacy config changed") }
}
