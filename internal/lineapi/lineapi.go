package lineapi

import (
	"fmt"
	"io"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// Client encapsulates functionality for interacting with the LINE API
type Client struct {
	bot         *linebot.Client
	apiEndpoint string
}

// MockContentResponse is a test helper that implements the same interface
// as linebot.MessageContentResponse for testing purposes
type MockContentResponse struct {
	Content     io.ReadCloser
	ContentType string
}

// GetContent returns the content of the mock response
func (m *MockContentResponse) GetContent() io.ReadCloser {
	return m.Content
}

// GetContentType returns the content type of the mock response
func (m *MockContentResponse) GetContentType() string {
	return m.ContentType
}

// NewClient creates a new instance of the LINE API client
func NewClient(channelSecret, channelToken string) (*Client, error) {
	// Allow overriding the API endpoint for testing
	apiEndpoint := os.Getenv("LINE_API_ENDPOINT")

	// Create LINE bot client with options
	var bot *linebot.Client
	var err error

	if apiEndpoint != "" {
		// Use custom endpoint for testing
		bot, err = linebot.New(
			channelSecret,
			channelToken,
			linebot.WithEndpointBase(apiEndpoint),
		)
	} else {
		// Use default endpoint
		bot, err = linebot.New(channelSecret, channelToken)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create LINE bot client: %v", err)
	}

	return &Client{
		bot:         bot,
		apiEndpoint: apiEndpoint,
	}, nil
}

// GetBot returns the underlying linebot client
func (c *Client) GetBot() *linebot.Client {
	return c.bot
}

// GetMessageContent retrieves content for a specific message
func (c *Client) GetMessageContent(messageID string) (*linebot.MessageContentResponse, error) {
	content, err := c.bot.GetMessageContent(messageID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message content: %v", err)
	}

	return content, nil
}

// IsMedia checks if a message is a media type that can be downloaded
func IsMedia(message linebot.Message) bool {
	switch message.(type) {
	case *linebot.ImageMessage,
		*linebot.VideoMessage,
		*linebot.AudioMessage,
		*linebot.FileMessage:
		return true
	default:
		return false
	}
}

// GetMediaType returns a string representation of the media type
func GetMediaType(message linebot.Message) string {
	switch message.(type) {
	case *linebot.ImageMessage:
		return "image"
	case *linebot.VideoMessage:
		return "video"
	case *linebot.AudioMessage:
		return "audio"
	case *linebot.FileMessage:
		return "file"
	default:
		return "unknown"
	}
}
