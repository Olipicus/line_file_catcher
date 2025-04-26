package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"code.olipicus.com/line_file_catcher/internal/media"
	"code.olipicus.com/line_file_catcher/internal/utils"
)

// StatsHandler handles statistics requests
type StatsHandler struct {
	logger     *utils.Logger
	mediaStore *media.MediaStore
}

// StatsResponse represents the statistics response
type StatsResponse struct {
	MediaStats    media.Stats `json:"mediaStats"`
	ProcessedRate float64     `json:"processedPerMinute"` // Files processed per minute
	StorageRate   float64     `json:"storageBytesPerMin"` // Bytes stored per minute
	Timestamp     time.Time   `json:"timestamp"`
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(logger *utils.Logger, mediaStore *media.MediaStore) *StatsHandler {
	return &StatsHandler{
		logger:     logger,
		mediaStore: mediaStore,
	}
}

// HandleStats processes statistics requests
func (h *StatsHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Received stats request from %s", r.RemoteAddr)

	// Get media stats
	stats := h.mediaStore.GetStats()

	// Calculate processing rates
	var processedRate float64
	var storageRate float64

	// Duration since start
	uptime := time.Since(stats.StartTime)
	uptimeMinutes := uptime.Minutes()

	if uptimeMinutes > 0 {
		totalProcessed := stats.ImageCount + stats.VideoCount + stats.AudioCount + stats.FileCount
		processedRate = float64(totalProcessed) / uptimeMinutes
		storageRate = float64(stats.TotalBytes) / uptimeMinutes
	}

	response := StatsResponse{
		MediaStats:    stats,
		ProcessedRate: processedRate,
		StorageRate:   storageRate,
		Timestamp:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode stats response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Stats request processed successfully")
}
