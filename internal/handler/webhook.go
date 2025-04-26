package handler

import (
	"fmt"
	"net/http"
	"time"

	"code.olipicus.com/line_file_catcher/internal/lineapi"
	"code.olipicus.com/line_file_catcher/internal/media"
	"code.olipicus.com/line_file_catcher/internal/utils"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// WebhookHandler handles LINE webhook events
type WebhookHandler struct {
	lineClient  *lineapi.Client
	mediaStore  *media.MediaStore
	logger      *utils.Logger
	rateLimiter *utils.RateLimiter
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(lineClient *lineapi.Client, mediaStore *media.MediaStore, logger *utils.Logger) *WebhookHandler {
	// Create a rate limiter that allows 60 requests per minute (1 request per second on average)
	rateLimiter := utils.NewRateLimiter(60, time.Minute)

	return &WebhookHandler{
		lineClient:  lineClient,
		mediaStore:  mediaStore,
		logger:      logger,
		rateLimiter: rateLimiter,
	}
}

// HandleWebhook processes webhook requests from LINE
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Received webhook request from %s", r.RemoteAddr)

	// Apply rate limiting
	if !h.rateLimiter.Allow() {
		h.logger.Warning("Rate limit exceeded for request from %s", r.RemoteAddr)
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(h.rateLimiter.ResetInterval().Seconds())))
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Verify signature
	events, err := h.lineClient.GetBot().ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			h.logger.Error("Invalid signature in webhook request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.logger.Error("Error parsing webhook request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Info("Received %d events in webhook request", len(events))

	for i, event := range events {
		h.logger.Debug("Processing event %d of type %s", i+1, event.Type)
		if err := h.handleEvent(event); err != nil {
			h.logger.Error("Error handling event: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Webhook request processed successfully")
}

// handleEvent processes a single LINE event
func (h *WebhookHandler) handleEvent(event *linebot.Event) error {
	switch event.Type {
	case linebot.EventTypeMessage:
		return h.handleMessageEvent(event)
	default:
		// Ignore other event types
		h.logger.Debug("Ignoring non-message event type: %s", event.Type)
		return nil
	}
}

// handleMessageEvent processes a message event
func (h *WebhookHandler) handleMessageEvent(event *linebot.Event) error {
	// Since event.Message is an interface, we need to check its type
	if !lineapi.IsMedia(event.Message) {
		// Ignore non-media messages
		h.logger.Debug("Ignoring non-media message type")
		return nil
	}

	// Get media type and ID
	mediaType := lineapi.GetMediaType(event.Message)
	messageID := getMessageID(event.Message)

	h.logger.Info("Processing %s message with ID: %s from user: %s",
		mediaType, messageID, event.Source.UserID)

	// Get content directly using the LINE client
	content, err := h.lineClient.GetMessageContent(messageID)
	if err != nil {
		h.logger.Error("Failed to get message content: %v", err)
		return err
	}

	// Process the content using our MediaStore
	filePath, err := h.mediaStore.SaveMedia(messageID, mediaType, content)
	if err != nil {
		h.logger.Error("Failed to save media: %v", err)
		return err
	}

	h.logger.Info("Media saved to: %s", filePath)

	// Get user ID for sending follow-up messages
	userID := event.Source.UserID

	// Register a callback for when the file is uploaded to Google Drive
	h.mediaStore.RegisterUploadCallback(filePath, func(filename string, fileLink string) error {
		// Send a message with the Google Drive link
		return h.sendDriveLinkMessage(userID, filename, fileLink)
	})

	// Optional: Send a confirmation message back to the user
	if replyToken := event.ReplyToken; replyToken != "" {
		if err := h.sendConfirmationMessage(replyToken, mediaType); err != nil {
			h.logger.Error("Error sending confirmation: %v", err)
		}
	}

	return nil
}

// getMessageID extracts the message ID from the message interface
func getMessageID(message linebot.Message) string {
	switch m := message.(type) {
	case *linebot.TextMessage:
		return m.ID
	case *linebot.ImageMessage:
		return m.ID
	case *linebot.VideoMessage:
		return m.ID
	case *linebot.AudioMessage:
		return m.ID
	case *linebot.FileMessage:
		return m.ID
	case *linebot.LocationMessage:
		return m.ID
	case *linebot.StickerMessage:
		return m.ID
	default:
		return ""
	}
}

// sendConfirmationMessage sends a confirmation message back to the user
func (h *WebhookHandler) sendConfirmationMessage(replyToken, mediaType string) error {
	message := fmt.Sprintf("Thanks for sharing! Your %s file has been received and is being processed.", mediaType)

	h.logger.Debug("Sending confirmation message for %s", mediaType)

	if _, err := h.lineClient.GetBot().ReplyMessage(replyToken, linebot.NewTextMessage(message)).Do(); err != nil {
		return fmt.Errorf("error sending confirmation message: %v", err)
	}

	h.logger.Debug("Confirmation message sent successfully")
	return nil
}

// sendDriveLinkMessage sends a message with the Google Drive link back to the user
func (h *WebhookHandler) sendDriveLinkMessage(replyToken, filename, fileLink string) error {
	message := fmt.Sprintf("📁 Your file %s has been backed up to Google Drive and is available at: %s", filename, fileLink)

	h.logger.Debug("Sending Google Drive link message for %s", filename)

	if _, err := h.lineClient.GetBot().PushMessage(replyToken, linebot.NewTextMessage(message)).Do(); err != nil {
		return fmt.Errorf("error sending Google Drive link message: %v", err)
	}

	h.logger.Info("Google Drive link message sent successfully")
	return nil
}
