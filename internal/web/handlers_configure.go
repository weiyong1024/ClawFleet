package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/weiyong1024/clawsandbox/internal/container"
)

// configureRequest is the JSON body for POST /api/v1/instances/{name}/configure.
type configureRequest struct {
	Provider     string `json:"provider"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	Channel      string `json:"channel"`
	ChannelToken string `json:"channel_token"`
}

// handleConfigureInstance configures an OpenClaw instance via docker exec.
func (s *Server) handleConfigureInstance(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var req configureRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Provider == "" || req.APIKey == "" {
		writeError(w, http.StatusBadRequest, "provider and api_key are required")
		return
	}

	store, err := s.loadStore()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	inst := store.Get(name)
	if inst == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("instance %s not found", name))
		return
	}

	// Ensure instance is running
	status, _, _ := container.Status(s.docker, inst.ContainerID)
	if status != "running" {
		if err := container.Start(s.docker, inst.ContainerID); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("starting instance: %v", err))
			return
		}
		store.SetStatus(name, "running")
		_ = store.Save()
	}

	if err := container.Configure(s.docker, container.ConfigureParams{
		ContainerID:  inst.ContainerID,
		Provider:     req.Provider,
		APIKey:       req.APIKey,
		Model:        req.Model,
		Channel:      req.Channel,
		ChannelToken: req.ChannelToken,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("configure failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"status":  "configured",
			"message": fmt.Sprintf("Instance %s configured successfully", name),
		},
	})
}

// handleConfigureStatus returns the configuration status of an instance.
func (s *Server) handleConfigureStatus(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	store, err := s.loadStore()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	inst := store.Get(name)
	if inst == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("instance %s not found", name))
		return
	}

	status, _, _ := container.Status(s.docker, inst.ContainerID)
	if status != "running" {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": &container.ConfigInfo{Configured: false},
		})
		return
	}

	info, err := container.ConfigStatus(s.docker, inst.ContainerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": info})
}
