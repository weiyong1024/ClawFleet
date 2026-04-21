package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clawfleet/clawfleet/internal/container"
	"github.com/clawfleet/clawfleet/internal/version"
)

// handleImageStatus reports whether the sandbox Docker image has been built.
func (s *Server) handleImageStatus(w http.ResponseWriter, r *http.Request) {
	openclawExists, err := container.ImageExists(s.docker, s.config.ImageRef())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	hermesExists, _ := container.ImageExists(s.docker, s.config.HermesImageRef())
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"image":        s.config.ImageRef(),
			"built":        openclawExists,
			"hermes_image": s.config.HermesImageRef(),
			"hermes_built": hermesExists,
		},
	})
}

// handleOpenClawVersions queries the npm registry and returns available OpenClaw versions.
func (s *Server) handleOpenClawVersions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://registry.npmjs.org/openclaw", nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create request")
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Registry unreachable — return only the recommended version.
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"recommended": version.RecommendedOpenClawVersion,
				"versions":    []string{version.RecommendedOpenClawVersion},
				"latest":      "",
				"error":       "npm registry unreachable",
			},
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reading registry response")
		return
	}

	var registry struct {
		Versions map[string]json.RawMessage `json:"versions"`
		DistTags map[string]string          `json:"dist-tags"`
	}
	if err := json.Unmarshal(body, &registry); err != nil {
		writeError(w, http.StatusInternalServerError, "parsing registry response")
		return
	}

	versions := make([]string, 0, len(registry.Versions))
	for v := range registry.Versions {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool {
		return versionLess(versions[j], versions[i]) // descending
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"recommended": version.RecommendedOpenClawVersion,
			"versions":    versions,
			"latest":      registry.DistTags["latest"],
		},
	})
}

// versionLess compares two dot-separated version strings numerically.
// Works for both semver (1.2.3) and date-based (2026.3.24) versioning.
// Pre-release versions (e.g. 2026.3.24-beta.1) sort before their release.
func versionLess(a, b string) bool {
	aBase, aIsPrerelease := splitPrerelease(a)
	bBase, bIsPrerelease := splitPrerelease(b)

	pa := parseVersionParts(aBase)
	pb := parseVersionParts(bBase)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		va, vb := 0, 0
		if i < len(pa) {
			va = pa[i]
		}
		if i < len(pb) {
			vb = pb[i]
		}
		if va != vb {
			return va < vb
		}
	}
	// Same base version — pre-release < release.
	if aIsPrerelease != bIsPrerelease {
		return aIsPrerelease
	}
	return a < b // both pre-release: lexicographic fallback
}

// splitPrerelease splits "2026.3.24-beta.1" into ("2026.3.24", true).
func splitPrerelease(v string) (string, bool) {
	v = strings.TrimPrefix(v, "v")
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		return v[:idx], true
	}
	return v, false
}

func parseVersionParts(v string) []int {
	parts := strings.Split(v, ".")
	nums := make([]int, len(parts))
	for i, p := range parts {
		nums[i], _ = strconv.Atoi(p)
	}
	return nums
}

// handleImageBuild triggers a Docker image build and streams progress via SSE.
func (s *Server) handleImageBuild(w http.ResponseWriter, r *http.Request) {
	// Parse optional JSON body for openclaw_version.
	openclawVersion := version.RecommendedOpenClawVersion
	if r.Body != nil && r.ContentLength != 0 {
		var req struct {
			OpenClawVersion string `json:"openclaw_version"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err == nil && req.OpenClawVersion != "" {
			openclawVersion = req.OpenClawVersion
		}
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	writeSSE(w, "log", fmt.Sprintf("Building with OpenClaw %s", openclawVersion))
	flusher.Flush()

	imageRef := s.config.ImageRef()
	pr, pw := newLineWriter(r.Context())

	done := make(chan error, 1)
	go func() {
		done <- container.Build(s.docker, imageRef, openclawVersion, pw)
		pw.Close()
	}()

	for line := range pr {
		writeSSE(w, "log", line)
		flusher.Flush()
	}

	if err := <-done; err != nil {
		writeSSE(w, "error", err.Error())
	} else {
		writeSSE(w, "done", "image built successfully")
	}
	flusher.Flush()
}

// handleImagePull pulls the Docker image from the registry and streams progress via SSE.
func (s *Server) handleImagePull(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	repo := s.config.Image.Name
	tag := s.config.Image.Tag
	pr, pw := newLineWriter(r.Context())

	done := make(chan error, 1)
	go func() {
		done <- container.PullImage(s.docker, repo, tag, pw)
		pw.Close()
	}()

	for line := range pr {
		writeSSE(w, "log", line)
		flusher.Flush()
	}

	if err := <-done; err != nil {
		writeSSE(w, "error", err.Error())
	} else {
		writeSSE(w, "done", fmt.Sprintf("image %s:%s pulled successfully", repo, tag))
	}
	flusher.Flush()
}

// writeSSE writes a single Server-Sent Event, handling multi-line data correctly.
func writeSSE(w http.ResponseWriter, event, data string) {
	w.Write([]byte("event: " + event + "\n"))
	for _, line := range strings.Split(data, "\n") {
		w.Write([]byte("data: " + line + "\n"))
	}
	w.Write([]byte("\n"))
}

// newLineWriter returns a channel that receives lines as they are written.
// It respects context cancellation to avoid blocking the build goroutine.
func newLineWriter(ctx context.Context) (<-chan string, *lineWriter) {
	ch := make(chan string, 64)
	return ch, &lineWriter{ch: ch, ctx: ctx}
}

type lineWriter struct {
	ch  chan string
	ctx context.Context
	buf []byte
}

func (lw *lineWriter) Write(p []byte) (int, error) {
	lw.buf = append(lw.buf, p...)
	for {
		idx := -1
		for i, b := range lw.buf {
			if b == '\n' {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}
		line := string(lw.buf[:idx])
		lw.buf = lw.buf[idx+1:]
		select {
		case lw.ch <- line:
		case <-lw.ctx.Done():
			return len(p), lw.ctx.Err()
		}
	}
	return len(p), nil
}

func (lw *lineWriter) Close() error {
	if len(lw.buf) > 0 {
		select {
		case lw.ch <- string(lw.buf):
		case <-lw.ctx.Done():
		}
		lw.buf = nil
	}
	close(lw.ch)
	return nil
}
