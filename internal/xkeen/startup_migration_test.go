package xkeen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartupMigrationRuntimeGate(t *testing.T) {
	for _, tc := range []struct {
		name string
		rt Runtime
		wantChange bool
	}{
		{"mihomo", Runtime{Installed: true, Core: CoreMihomo}, false},
		{"unknown", Runtime{Installed: true, Core: "unknown"}, false},
		{"missing", Runtime{Core: CoreXray}, false},
		{"xray", Runtime{Installed: true, Core: CoreXray}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "00_api.json")
			original := []byte(`{"api":{"tag":"api","services":["RoutingService"]}}`)
			if err := os.WriteFile(path, original, 0600); err != nil { t.Fatal(err) }
			changed, err := EnsureAPIConfigForRuntime(tc.rt, PoolState{Enabled: true, APIFile: path}, "127.0.0.1:10085")
			if err != nil { t.Fatal(err) }
			if changed != tc.wantChange { t.Fatalf("changed=%v, want %v", changed, tc.wantChange) }
			got, err := os.ReadFile(path)
			if err != nil { t.Fatal(err) }
			if !tc.wantChange {
				if string(got) != string(original) { t.Fatal("inactive core config modified") }
				if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) { t.Fatalf("unexpected backup: %v", err) }
			} else {
				var cfg map[string]interface{}
				if err := ReadJSONC(path, &cfg); err != nil { t.Fatal(err) }
				api := cfg["api"].(map[string]interface{})
				if api["listen"] != "127.0.0.1:10085" || !hasService(api["services"], "HandlerService") { t.Fatal("Xray migration incomplete") }
			}
		})
	}
}
