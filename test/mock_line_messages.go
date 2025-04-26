package test

import (
	"bytes"
	"io"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// MockContentResponse implements the same interface as linebot.MessageContentResponse
// for testing purposes
type MockContentResponse struct {
	ContentData []byte
	ContentType string
}

// GetContent returns a ReadCloser for the content
func (m *MockContentResponse) GetContent() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(m.ContentData))
}

// GetContentType returns the content type
func (m *MockContentResponse) GetContentType() string {
	return m.ContentType
}

// CreateMockEvent creates a mock LINE event for testing
func CreateMockEvent(messageType string, messageID string) *linebot.Event {
	var message linebot.Message

	switch messageType {
	case "image":
		message = &linebot.ImageMessage{
			ID: messageID,
		}
	case "video":
		message = &linebot.VideoMessage{
			ID: messageID,
		}
	case "audio":
		message = &linebot.AudioMessage{
			ID: messageID,
		}
	case "file":
		message = &linebot.FileMessage{
			ID: messageID,
		}
	default:
		// Default to image if unknown type
		message = &linebot.ImageMessage{
			ID: messageID,
		}
	}

	return &linebot.Event{
		ReplyToken: "mock_reply_token",
		Type:       linebot.EventTypeMessage,
		Timestamp:  time.Now(),
		Source: &linebot.EventSource{
			Type:   linebot.EventSourceTypeUser,
			UserID: "mock_user_id",
		},
		Message: message,
	}
}
