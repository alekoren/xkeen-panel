package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"xkeen-panel/internal/xkeen"
)

type coreSwitch struct {
	detector *xkeen.Detector
	mu sync.Mutex
	target string
	lastError string
}

func coreJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (c *coreSwitch) get(w http.ResponseWriter, _ *http.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()
	rt := c.detector.Runtime()
	coreJSON(w, http.StatusOK, map[string]any{"core": rt.Core, "mode": rt.Mode, "available": rt.Installed && rt.Generation == 2, "running": xkeen.IsRunning(rt.Core), "busy": c.target != "" || xkeen.IsRestarting(), "target": c.target, "error": c.lastError})
}

func (c *coreSwitch) post(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target string `json:"target"`
		Expected string `json:"expected"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	dec.DisallowUnknownFields()
	if dec.Decode(&req) != nil || (req.Target != xkeen.CoreXray && req.Target != xkeen.CoreMihomo) {
		coreJSON(w, http.StatusBadRequest, map[string]string{"error": "Недопустимое ядро"})
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	rt := c.detector.Runtime()
	if c.target != "" || xkeen.IsRestarting() || !rt.Installed || rt.Generation != 2 || rt.Core != req.Expected || rt.Core == req.Target || !xkeen.IsRunning(rt.Core) {
		coreJSON(w, http.StatusConflict, map[string]string{"error": "Состояние XKeen изменилось или смена ядра недоступна. Обновите статус."})
		return
	}
	c.target, c.lastError = req.Target, ""
	go func(target string) {
		err := c.detector.SwitchCore(target)
		c.mu.Lock()
		defer c.mu.Unlock()
		if err != nil {
			c.lastError = err.Error()
		}
		c.target = ""
	}(req.Target)
	coreJSON(w, http.StatusAccepted, map[string]bool{"accepted": true})
}
