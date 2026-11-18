package mihomo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestControllerBaseURL(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", "http://mihomo.local"},
		{"192.168.2.1:9090", "http://192.168.2.1:9090"},
		{"0.0.0.0:9090", "http://127.0.0.1:9090"},
		{"[::]:9090", "http://127.0.0.1:9090"},
	} {
		got, err := controllerBaseURL(tc.in)
		if err != nil || got != tc.want { t.Fatalf("input=%q got=%q err=%v", tc.in, got, err) }
	}
	for _, in := range []string{"user:pass@host:9090", "host:9090/path", "host:9090?q=x", "host:0", "host:65536", "http://host:9090"} {
		if _, err := controllerBaseURL(in); err == nil { t.Fatalf("accepted invalid authority %q", in) }
	}
}

func TestUnixControllerConfiguredHostAndForbidden(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "mh-host-")
	if err != nil { t.Fatal(err) }
	t.Cleanup(func(){ _ = os.RemoveAll(dir) })
	socket := filepath.Join(dir, "api.sock")
	ln, err := net.Listen("unix", socket)
	if err != nil { t.Fatal(err) }
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "192.168.2.1:9090" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "<html>fake-sensitive-body</html>")
			return
		}
		switch r.URL.Path {
		case "/proxies": fmt.Fprint(w, `{"proxies":{}}`)
		case "/providers/proxies": fmt.Fprint(w, `{"providers":{}}`)
		default: w.WriteHeader(http.StatusNotFound)
		}
	})}
	go func(){ _ = server.Serve(ln) }()
	t.Cleanup(func(){ _ = server.Close() })
	config := filepath.Join(dir, "config.yaml")
	for _, authority := range []string{"192.168.2.1:9090", "localhost:9090"} {
		data := fmt.Sprintf("external-controller-unix: %q\nexternal-controller: %q\n", socket, authority)
		if err := os.WriteFile(config, []byte(data), 0600); err != nil { t.Fatal(err) }
		c, err := OpenController(config)
		if err != nil { t.Fatal(err) }
		_, err = c.Snapshot(context.Background())
		if authority == "192.168.2.1:9090" {
			if err != nil { t.Fatal(err) }
		} else if err == nil || !strings.Contains(err.Error(), "HTTP 403") || strings.Contains(err.Error(), "fake-sensitive-body") {
			t.Fatalf("incorrect forbidden handling: %v", err)
		}
	}
	if err := os.Remove(socket); err != nil { t.Fatal(err) }
	c, err := OpenController(config)
	if err != nil { t.Fatal(err) }
	if _, err := c.Snapshot(context.Background()); err == nil { t.Fatal("missing Unix socket accepted") }
}
