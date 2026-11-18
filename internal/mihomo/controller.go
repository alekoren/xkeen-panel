package mihomo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type DelaySample struct {
	Time  string `json:"time"`
	Delay int    `json:"delay"`
}

type RuntimeProxy struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Now     string        `json:"now,omitempty"`
	All     []string      `json:"all,omitempty"`
	Alive   bool          `json:"alive"`
	History []DelaySample `json:"history,omitempty"`
}

type RuntimeProvider struct {
	Type    string         `json:"type"`
	Vehicle string         `json:"vehicleType"`
	Proxies []RuntimeProxy `json:"proxies"`
}

type RuntimeSnapshot struct {
	Proxies   map[string]RuntimeProxy    `json:"proxies"`
	Providers map[string]RuntimeProvider `json:"providers"`
}

type Controller struct {
	client *http.Client
	secret string
}

// OpenController reads the socket and credentials from the local Mihomo config.
// No caller-supplied host, TCP fallback, or redirect is allowed.
func OpenController(configPath string) (*Controller, error) {
	var cfg struct {
		Socket string `yaml:"external-controller-unix"`
		Secret string `yaml:"secret"`
	}
	body, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read Mihomo configuration")
	}
	if yaml.Unmarshal(body, &cfg) != nil || !filepath.IsAbs(cfg.Socket) {
		return nil, fmt.Errorf("Mihomo requires an absolute external-controller-unix path")
	}
	transport := &http.Transport{DisableKeepAlives: true, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", cfg.Socket)
	}}
	return &Controller{client: &http.Client{Transport: transport, Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}, secret: cfg.Secret}, nil
}

func (c *Controller) request(ctx context.Context, method, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, "http://mihomo.local"+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("cannot create Mihomo request")
	}
	if c.secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.secret)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Mihomo controller is unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Mihomo controller returned HTTP %d", resp.StatusCode)
	}
	if out != nil && json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(out) != nil {
		return fmt.Errorf("invalid Mihomo controller response")
	}
	return nil
}

// Snapshot only reads current state; it never starts health checks.
func (c *Controller) Snapshot(ctx context.Context) (RuntimeSnapshot, error) {
	var snapshot RuntimeSnapshot
	if err := c.request(ctx, http.MethodGet, "/proxies", nil, &snapshot); err != nil {
		return RuntimeSnapshot{}, err
	}
	if err := c.request(ctx, http.MethodGet, "/providers/proxies", nil, &snapshot); err != nil {
		return RuntimeSnapshot{}, err
	}
	return snapshot, nil
}

// Select chooses a member of an existing manual selector without rewriting YAML.
func (c *Controller) Select(ctx context.Context, group, member string) error {
	if group == "" || group == "." || group == ".." || strings.Contains(group, "/") || member == "" {
		return fmt.Errorf("invalid selector or member")
	}
	path := "/proxies/" + url.PathEscape(group)
	var proxy RuntimeProxy
	if err := c.request(ctx, http.MethodGet, path, nil, &proxy); err != nil {
		return err
	}
	if !strings.EqualFold(proxy.Type, "Selector") || !slices.Contains(proxy.All, member) {
		return fmt.Errorf("target is not a member of a Mihomo selector")
	}
	body, _ := json.Marshal(map[string]string{"name": member})
	return c.request(ctx, http.MethodPut, path, body, nil)
}
