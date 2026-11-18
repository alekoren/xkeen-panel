package xkeen

import (
 "fmt"
 "log"
 "strings"
 "xkeen-panel/internal/models"
)

// ApplyServer is the legacy Xray writer, never a Mihomo profile selector.
func ApplyServer(rt Runtime,outboundsPath string,server *models.Server) error {
 if rt.Core==CoreMihomo { return fmt.Errorf("старый выбор сервера Xray отключён при активном Mihomo; используйте панель Mihomo") }
 if err:=UpdateOutbound(outboundsPath,server); err!=nil { return err }
 if !rt.Installed || rt.Dispatcher=="" { return nil }
 output,err:=TestConfig(rt.Dispatcher,rt.Core); if err==nil { return nil }
 if rbErr:=RestoreBackup(outboundsPath); rbErr!=nil {
  log.Printf("[APPLY] Откат не удался: %v",rbErr)
  return fmt.Errorf("конфигурация %s не прошла проверку, откат не удался: %v (%s)",rt.Core,rbErr,output)
 }
 return fmt.Errorf("конфигурация %s не прошла проверку, изменения отменены: %s",rt.Core,TailLines(output,4))
}
func TailLines(s string,n int) string {
 lines:=strings.Split(strings.TrimSpace(s),"\n"); kept:=make([]string,0,n)
 for i:=len(lines)-1;i>=0 && len(kept)<n;i-- { if line:=strings.TrimSpace(lines[i]);line!="" { kept=append([]string{line},kept...) } }
 if len(kept)==0 { return "вывод пуст" }; return strings.Join(kept,"; ")
}
