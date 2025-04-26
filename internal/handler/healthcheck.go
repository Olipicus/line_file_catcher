package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"code.olipicus.com/line_file_catcher/internal/media"
	"code.olipicus.com/line_file_catcher/internal/utils"
)

// HealthCheckHandler handles health check requests
type HealthCheckHandler struct {
	startTime  time.Time
	logger     *utils.Logger
	mediaStore *media.MediaStore
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status    string      `json:"status"`
	Uptime    string      `json:"uptime"`
	GoVersion string      `json:"goVersion"`
	Memory    MemStats    `json:"memory"`
	Stats     media.Stats `json:"stats"`
	Timestamp time.Time   `json:"timestamp"`
}

// MemStats represents memory statistics
type MemStats struct {
	Alloc      uint64 `json:"alloc"`      // bytes allocated and not yet freed
	TotalAlloc uint64 `json:"totalAlloc"` // bytes allocated (even if freed)
	Sys        uint64 `json:"sys"`        // bytes obtained from system
	NumGC      uint32 `json:"numGC"`      // number of completed GC cycles
}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler(logger *utils.Logger, mediaStore *media.MediaStore) *HealthCheckHandler {
	return &HealthCheckHandler{
		startTime:  time.Now(),
		logger:     logger,
		mediaStore: mediaStore,
	}
}

// HandleHealthCheck processes health check requests
func (h *HealthCheckHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Received health check request from %s", r.RemoteAddr)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := HealthCheckResponse{
		Status:    "OK",
		Uptime:    time.Since(h.startTime).String(),
		GoVersion: runtime.Version(),
		Memory: MemStats{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
		Stats:     h.mediaStore.GetStats(), // Include media processing statistics
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode health check response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Health check request processed successfully")
}
