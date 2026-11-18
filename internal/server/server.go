package server

import (
 "encoding/json"
 "fmt"
 "io/fs"
 "log"
 "net/http"
 "path/filepath"
 "time"
 "xkeen-panel/internal/api"
 "xkeen-panel/internal/auth"
 "xkeen-panel/internal/coreupdate"
 "xkeen-panel/internal/geoip"
 "xkeen-panel/internal/keenetic"
 "xkeen-panel/internal/models"
 "xkeen-panel/internal/monitor"
 "xkeen-panel/internal/sse"
 "xkeen-panel/internal/updater"
 "xkeen-panel/internal/xkeen"
 "github.com/go-chi/chi/v5"
 "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
 config *models.Config
 userManager *auth.UserManager
 subscription *xkeen.SubscriptionManager
 watchdog *monitor.Watchdog
 detector *xkeen.Detector
 pool *xkeen.PoolStore
 sites *xkeen.SiteManager
 geoip *geoip.Matcher
 eventBus *sse.EventBus
 updater *updater.Updater
 coreUpdater *coreupdate.Updater
 settings *api.PanelSettingsHandler
 frontendFS fs.FS
}

func New(cfg *models.Config, um *auth.UserManager, sub *xkeen.SubscriptionManager, wd *monitor.Watchdog, det *xkeen.Detector, pool *xkeen.PoolStore, sites *xkeen.SiteManager, matcher *geoip.Matcher, bus *sse.EventBus, upd *updater.Updater, core *coreupdate.Updater, settings *api.PanelSettingsHandler, frontendFS fs.FS) *Server {
 return &Server{config:cfg,userManager:um,subscription:sub,watchdog:wd,detector:det,pool:pool,sites:sites,geoip:matcher,eventBus:bus,updater:upd,coreUpdater:core,settings:settings,frontendFS:frontendFS}
}
func (s *Server) Handler() http.Handler {
 r := chi.NewRouter()
 r.Use(middleware.Logger); r.Use(middleware.Recoverer)
 rateLimiter := api.NewRateLimiter(5,time.Minute)
 a := api.NewAuthHandler(s.userManager,rateLimiter,s.config)
 wa := api.NewWebAuthnHandler(s.userManager,rateLimiter,s.config)
 u := api.NewUpdateHandler(s.updater)
 cu := api.NewCoreUpdateHandler(s.coreUpdater)
 d := api.NewDevicesHandler(s.detector,keenetic.DefaultAddr)
 setup := api.NewSetupHandler(s.detector,s.subscription,s.watchdog,d)
 sites := api.NewSitesHandler(s.sites)
 h := api.NewHandlers(s.config,s.subscription,s.watchdog,s.detector,s.pool,s.geoip)
 core := &coreSwitch{detector:s.detector}
 sources := &sourceManager{dir:filepath.Join(s.config.DataDir,"sources")}
 r.Route("/api",func(r chi.Router) {
  r.Route("/auth",func(r chi.Router) {
   r.Get("/status",a.HandleAuthStatus); r.Post("/setup",a.HandleSetup); r.Post("/setup/confirm",a.HandleSetupConfirm)
   r.With(api.RateLimitMiddleware(rateLimiter,s.config.TrustProxyHeaders)).Post("/login",a.HandleLogin)
   r.With(api.RateLimitMiddleware(rateLimiter,s.config.TrustProxyHeaders)).Post("/login/passkey/begin",wa.HandleLoginBegin)
   r.With(api.RateLimitMiddleware(rateLimiter,s.config.TrustProxyHeaders)).Post("/login/passkey/finish",wa.HandleLoginFinish)
  })
  r.Group(func(r chi.Router) {
   r.Use(api.AuthMiddleware(s.userManager))
   r.Get("/events",sse.HandleEvents(s.eventBus,s.watchdog))
   r.Get("/servers/check",sse.HandleStreamLatency(s.subscription,s.config.ProbeConcurrency,time.Duration(s.config.ProbeTimeoutMs)*time.Millisecond))
  })
  r.Group(func(r chi.Router) {
   r.Use(api.AuthMiddleware(s.userManager)); r.Use(middleware.Timeout(30*time.Second))
   r.Get("/sources",sources.list); r.Post("/sources",sources.create)
   r.Get("/status",h.HandleStatus)
   r.Get("/subscription",h.HandleGetSubscription); r.Post("/subscription",h.HandleUpdateSubscription); r.Post("/subscription/refresh",h.HandleRefreshSubscription)
   r.Get("/servers",h.HandleGetServers); r.Post("/servers/select",h.HandleSelectServer); r.Post("/servers/check",h.HandleCheckServers); r.Post("/servers/country",h.HandleSetCountry)
   r.Post("/account/passkey/register/begin",wa.HandleRegisterBegin); r.Post("/account/passkey/register/finish",wa.HandleRegisterFinish); r.Get("/account/passkey",wa.HandlePasskeyList); r.Delete("/account/passkey",wa.HandlePasskeyDelete)
   r.Get("/xkeen/core",core.get); r.Post("/xkeen/core",core.post)
   r.Post("/xkeen/restart",h.HandleRestart); r.Post("/xkeen/start",h.HandleStart); r.Post("/xkeen/stop",h.HandleStop); r.Post("/xkeen/selftest",h.HandleSelfTest)
   r.Get("/xkeen/settings",h.HandleGetSettings); r.Put("/xkeen/settings",h.HandleUpdateSettings); r.Get("/xkeen/lists/{name}",h.HandleGetList); r.Put("/xkeen/lists/{name}",h.HandleUpdateList)
   r.Post("/mihomo/sync",h.HandleMihomoSync)
   r.Get("/setup/state",setup.HandleState); r.Post("/setup/policy",setup.HandlePolicy); r.Post("/setup/finish",setup.HandleFinish)
   r.Get("/panel/settings",s.settings.HandleGet); r.Put("/panel/settings",s.settings.HandlePut)
   r.Get("/devices",d.HandleList); r.Post("/devices/vpn",d.HandleSetVPN); r.Post("/devices/new",d.HandleNewDevicesVPN)
   r.Get("/pool",h.HandlePoolStatus); r.Post("/pool/enable",h.HandlePoolEnable); r.Post("/pool/disable",h.HandlePoolDisable); r.Post("/pool/sync",h.HandlePoolSync); r.Post("/pool/pin",h.HandlePoolPin); r.Post("/pool/auto",h.HandlePoolAuto); r.Post("/pool/members",h.HandlePoolMember)
   r.Get("/sites",sites.HandleList); r.Post("/sites",sites.HandleCreate); r.Post("/sites/{id}/check",sites.HandleCheck); r.Delete("/sites/{id}",sites.HandleDelete); r.Post("/sites/{id}/lock",sites.HandleLock); r.Put("/sites/{id}/source",sites.HandleSource); r.Get("/sites/sources",sites.HandleSources); r.Delete("/sites/{id}/lock",sites.HandleUnlock)
   r.Get("/update",u.HandleStatus); r.Post("/update/check",u.HandleCheck); r.Post("/update/apply",u.HandleApply); r.Post("/update/rollback",u.HandleRollback); r.Put("/update/settings",u.HandleSettings)
   r.Get("/core/update",cu.HandleStatus); r.Post("/core/update/check",cu.HandleCheck); r.Post("/core/update/apply",cu.HandleApply); r.Post("/core/update/rollback",cu.HandleRollback)
   r.Get("/logs",h.HandleLogs)
   r.Post("/watchdog/toggle",func(w http.ResponseWriter,r *http.Request) {
    var req struct { Active bool `json:"active"` }
    if json.NewDecoder(r.Body).Decode(&req)!=nil { coreJSON(w,400,map[string]string{"error":"неверный формат"}); return }
    s.watchdog.SetActive(req.Active); coreJSON(w,200,map[string]bool{"active":req.Active})
   })
  })
 })
 if s.frontendFS!=nil {
  fileServer:=http.FileServer(http.FS(s.frontendFS))
  r.Get("/*",func(w http.ResponseWriter,r *http.Request) {
   path:=r.URL.Path; if path=="/" { path="/index.html" }
   f,err:=s.frontendFS.Open(path[1:]); if err==nil { f.Close(); fileServer.ServeHTTP(w,r); return }
   r.URL.Path="/"; fileServer.ServeHTTP(w,r)
  })
 }
 return r
}
func (s *Server) Start() error { addr:=fmt.Sprintf(":%d",s.config.Port); log.Printf("Сервер запущен на %s",addr); return http.ListenAndServe(addr,s.Handler()) }
