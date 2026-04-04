package web

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// handleConsoleProxy reverse-proxies requests to an instance's OpenClaw Gateway
// web UI. This is a fallback for remote Dashboard access (e.g. via SSH tunnel)
// where the Gateway port is not directly reachable. For local access, the
// Dashboard opens the Gateway's native port directly (see app.js onConsole).
//
// Route: /console/{name}/* → http://127.0.0.1:{gateway_port}/*
func (s *Server) handleConsoleProxy(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	// Redirect /console/{name} → /console/{name}/ so relative paths resolve
	// correctly (e.g. ./assets/xxx.js → /console/{name}/assets/xxx.js).
	// Skip redirect for WebSocket upgrades — they don't follow redirects.
	isWS := strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
	if !isWS && !strings.HasSuffix(r.URL.Path, "/") && !strings.Contains(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:], ".") {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
		return
	}

	store, err := s.loadStore()
	if err != nil {
		http.Error(w, "failed to load state", http.StatusInternalServerError)
		return
	}

	inst := store.Get(name)
	if inst == nil {
		http.Error(w, fmt.Sprintf("instance %q not found", name), http.StatusNotFound)
		return
	}
	if inst.Status != "running" {
		http.Error(w, fmt.Sprintf("instance %q is not running", name), http.StatusServiceUnavailable)
		return
	}

	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", inst.Ports.Gateway))

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			// Strip the /console/{name} prefix from the path
			prefix := fmt.Sprintf("/console/%s", name)
			req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
			req.URL.RawPath = ""
		},
	}

	proxy.ServeHTTP(w, r)
}
