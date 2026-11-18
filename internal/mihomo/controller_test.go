package mihomo

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestUnixController(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "mh-api-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	socket := filepath.Join(dir, "api.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-test-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.Method + " " + r.URL.Path {
		case "GET /proxies":
			fmt.Fprint(w, `{"proxies":{"AUTO":{"type":"URLTest","now":"Proton-025"}}}`)
		case "GET /providers/proxies":
			fmt.Fprint(w, `{"providers":{"PROTON-LOCAL":{"vehicleType":"File","proxies":[{"name":"Proton-025","history":[{"delay":304}],"private-key":"fake-test-key"}]}}}`)
		case "GET /proxies/Manual Group":
			fmt.Fprint(w, `{"type":"Selector","all":["AUTO","Proton-025"]}`)
		case "GET /proxies/AUTO":
			fmt.Fprint(w, `{"type":"URLTest","all":["Proton-025"]}`)
		case "PUT /proxies/Manual Group":
			var input map[string]string
			if json.NewDecoder(r.Body).Decode(&input) != nil || input["name"] != "Proton-025" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })
	config := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(config, []byte(fmt.Sprintf("{external-controller-unix: %q, secret: fake-test-secret}", socket)), 0600); err != nil {
		t.Fatal(err)
	}
	client, err := OpenController(config)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := client.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	nodes := snapshot.Providers["PROTON-LOCAL"].Proxies
	if snapshot.Proxies["AUTO"].Now != "Proton-025" || len(nodes) != 1 || len(nodes[0].History) != 1 || nodes[0].History[0].Delay != 304 {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	if err := client.Select(context.Background(), "Manual Group", "Proton-025"); err != nil {
		t.Fatal(err)
	}
	if client.Select(context.Background(), "Manual Group", "unknown") == nil || client.Select(context.Background(), "AUTO", "Proton-025") == nil {
		t.Fatal("invalid selection was accepted")
	}
}
