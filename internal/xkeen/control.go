package xkeen

import (
 "bytes"
 "context"
 "fmt"
 "log"
 "os"
 "os/exec"
 "path/filepath"
 "strconv"
 "strings"
 "sync"
 "syscall"
 "time"
)
var (
 restartMu sync.Mutex
 restarting bool
 OnRestartStateChange func(bool)
 runningMu sync.Mutex
 runningCached bool
 runningCore string
 runningCheckedAt time.Time
)
const ( restartTimeout=180*time.Second; quickTimeout=60*time.Second )
func runXkeen(dispatcher string, timeout time.Duration,args ...string) (string,error) {
 if dispatcher=="" { return "",fmt.Errorf("путь к xkeen не задан") }
 out,err:=os.CreateTemp("","xkeen-out-*"); if err!=nil { return "",err }; defer os.Remove(out.Name()); defer out.Close()
 ctx,cancel:=context.WithTimeout(context.Background(),timeout); defer cancel()
 cmd:=exec.CommandContext(ctx,dispatcher,args...); cmd.Env=append(os.Environ(),"XKEEN_FOREGROUND=1"); cmd.Stdout=out; cmd.Stderr=out; cmd.SysProcAttr=&syscall.SysProcAttr{Setsid:true}
 runErr:=cmd.Run(); body,_:=os.ReadFile(out.Name()); text:=strings.TrimSpace(string(ansiRe.ReplaceAll(body,nil)))
 if ctx.Err()==context.DeadlineExceeded { return text,fmt.Errorf("команда xkeen %s не завершилась за %s",strings.Join(args," "),timeout) }; return text,runErr
}
func Restart(dispatcher string) (string,error) {
 restartMu.Lock(); if restarting { restartMu.Unlock(); return "restart already in progress",nil }; restarting=true; restartMu.Unlock()
 if OnRestartStateChange!=nil { OnRestartStateChange(true) }
 go func() {
  output,err:=runXkeen(dispatcher,restartTimeout,"-restart")
  restartMu.Lock(); restarting=false; restartMu.Unlock()
  if OnRestartStateChange!=nil { OnRestartStateChange(false) }
  if err!=nil { log.Printf("[RESTART] Ошибка: %v (%s)",err,output) } else { log.Printf("[RESTART] Завершён успешно") }
 }(); return "restart initiated",nil
}
func Start(dispatcher string) (string,error) { return runXkeen(dispatcher,restartTimeout,"-start") }
func Stop(dispatcher string) (string,error) { return runXkeen(dispatcher,quickTimeout,"-stop") }
func Status(dispatcher string) (string,error) { return runXkeen(dispatcher,quickTimeout,"-status") }
func SetAutostart(dispatcher string,on bool) (string,error) { state:="off"; if on { state="on" }; return runXkeen(dispatcher,quickTimeout,"-auto",state) }
func TestConfig(dispatcher,core string) (string,error) { flag:="-xtest"; if core==CoreMihomo { flag="-mtest" }; return runXkeen(dispatcher,quickTimeout,flag) }
func IsRestarting() bool { restartMu.Lock(); defer restartMu.Unlock(); return restarting }
func IsRunning(core string) bool {
 if core=="" { core=CoreXray }; runningMu.Lock(); defer runningMu.Unlock()
 if core==runningCore && time.Since(runningCheckedAt)<2*time.Second { return runningCached }
 runningCached=coreProcessRunning(core); runningCore=core; runningCheckedAt=time.Now(); return runningCached
}
func coreCommand(data []byte,core string) bool {
 if core!=CoreXray && core!=CoreMihomo { return false }
 args:=bytes.Split(data,[]byte{0}); if len(args)==0 || filepath.Base(string(args[0]))!=core { return false }
 for _,arg:=range args[1:] {
  switch string(arg) { case "-t","--test","-test","-version","--version","version","-v": return false }
 }; return true
}
func coreProcessRunning(core string) bool {
 if core!=CoreXray && core!=CoreMihomo { return false }
 if pid,ok:=readPIDFile("/opt/var/run/"+core+".pid"); ok && processAlive(pid,core) { return true }
 entries,err:=os.ReadDir("/proc")
 if err!=nil {
  out,_:=exec.Command("ps","-axo","comm=").Output()
  for _,line:=range strings.Split(string(out),"\n") { if filepath.Base(strings.TrimSpace(line))==core { return true } }; return false
 }
 for _,entry:=range entries {
  if !entry.IsDir() { continue }; if _,err:=strconv.Atoi(entry.Name()); err!=nil { continue }
  data,err:=os.ReadFile("/proc/"+entry.Name()+"/cmdline"); if err==nil && coreCommand(data,core) { return true }
 }; return false
}
func readPIDFile(path string) (int,bool) {
 data,err:=os.ReadFile(path); if err!=nil { return 0,false }; pid,err:=strconv.Atoi(strings.TrimSpace(string(data))); return pid,err==nil && pid>0
}
func processAlive(pid int,core string) bool { data,err:=os.ReadFile("/proc/"+strconv.Itoa(pid)+"/cmdline"); return err==nil && coreCommand(data,core) }
