package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"code.olipicus.com/line_file_catcher/internal/media"
	"code.olipicus.com/line_file_catcher/internal/utils"
)

// StatsResponse represents the response for the stats endpoint
type StatsResponse struct {
	Status        string                 `json:"status"`
	Uptime        string                 `json:"uptime"`
	FileStats     media.Stats            `json:"fileStats"`
	CloudStats    map[string]interface{} `json:"cloudStats"`
	MemoryStats   map[string]interface{} `json:"memoryStats"`
	ProcessUptime string                 `json:"processUptime"`
}

// StatsHandler struct to handle stats requests
type StatsHandler struct {
	startTime  time.Time
	logger     *utils.Logger
	mediaStore *media.MediaStore
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(logger *utils.Logger, mediaStore *media.MediaStore) *StatsHandler {
	return &StatsHandler{
		startTime:  time.Now(),
		logger:     logger,
		mediaStore: mediaStore,
	}
}

// HandleStats processes stats requests
func (h *StatsHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Received stats request from %s", r.RemoteAddr)

	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Format memory stats for human readability
	memoryStats := map[string]interface{}{
		"allocatedMB":    float64(memStats.Alloc) / 1024 / 1024,
		"totalAllocMB":   float64(memStats.TotalAlloc) / 1024 / 1024,
		"systemMemoryMB": float64(memStats.Sys) / 1024 / 1024,
		"numGC":          memStats.NumGC,
	}

	// Get cloud storage statistics
	cloudStats := h.mediaStore.GetCloudStats()

	// Create the response
	response := StatsResponse{
		Status:        "ok",
		Uptime:        time.Since(h.startTime).String(),
		FileStats:     h.mediaStore.GetStats(),
		CloudStats:    cloudStats,
		MemoryStats:   memoryStats,
		ProcessUptime: time.Since(h.startTime).String(),
	}

	// Set content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode stats response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Stats request processed successfully")
}
