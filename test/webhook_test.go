package test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"code.olipicus.com/line_file_catcher/internal/config"
	"code.olipicus.com/line_file_catcher/internal/handler"
	"code.olipicus.com/line_file_catcher/internal/lineapi"
	"code.olipicus.com/line_file_catcher/internal/media"
	"code.olipicus.com/line_file_catcher/internal/utils"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

const (
	// Test configuration
	testChannelSecret = "test_channel_secret"
	testChannelToken  = "test_channel_token"
	testStorageDir    = "/tmp/line_file_catcher_test"
	testLogDir        = "/tmp/line_file_catcher_test/logs"
)

// mockLineServer creates a mock LINE API server for testing
type mockLineServer struct {
	server            *httptest.Server
	messageContentMap map[string][]byte
	contentTypeMap    map[string]string
	repliesReceived   []linebot.Message
}

// newMockLineServer creates a new mock LINE API server
func newMockLineServer() *mockLineServer {
	mock := &mockLineServer{
		messageContentMap: make(map[string][]byte),
		contentTypeMap:    make(map[string]string),
		repliesReceived:   make([]linebot.Message, 0),
	}

	// Create a test server
	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Mock server received request: %s %s\n", r.Method, r.URL.Path)

		// Match message content endpoint: "/v2/bot/message/%s/content"
		if contentRegex := regexp.MustCompile(`/v2/bot/message/([^/]+)/content`); contentRegex.MatchString(r.URL.Path) {
			matches := contentRegex.FindStringSubmatch(r.URL.Path)
			if len(matches) >= 2 {
				messageID := matches[1]
				fmt.Printf("Handling content request for message ID: %s\n", messageID)
				mock.handleContentRequest(w, r, messageID)
				return
			}
		}

		// Handle other LINE API endpoints based on exact path
		switch r.URL.Path {
		// Message API endpoints
		case "/v2/bot/message/reply":
			fmt.Printf("Handling reply message request\n")
			mock.handleReplyRequest(w, r)
		case "/v2/bot/message/push":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/multicast":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/broadcast":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/narrowcast":
			mock.handleDefaultSuccess(w, r)

		// Message validation endpoints
		case "/v2/bot/message/validate/push":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/validate/reply":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/validate/broadcast":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/validate/multicast":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/validate/narrowcast":
			mock.handleDefaultSuccess(w, r)

		// Message quota endpoints
		case "/v2/bot/message/quota":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/message/quota/consumption":
			mock.handleDefaultSuccess(w, r)

		// Profile-related endpoints
		case "/v2/bot/profile/":
			mock.handleDefaultSuccess(w, r)
		case "/v2/bot/followers/ids":
			mock.handleDefaultSuccess(w, r)

		// Bot info endpoint
		case "/v2/bot/info":
			mock.handleDefaultSuccess(w, r)

		// Default handler for any unhandled paths
		default:
			// Check for regex patterns for endpoints with parameters
			switch {
			// Group-related endpoints
			case regexp.MustCompile(`/v2/bot/group/[^/]+/leave`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/group/[^/]+/members/ids`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/group/[^/]+/members/count`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/group/[^/]+/member/[^/]+`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/group/[^/]+/summary`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)

			// Room-related endpoints
			case regexp.MustCompile(`/v2/bot/room/[^/]+/leave`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/room/[^/]+/members/ids`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/room/[^/]+/members/count`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/room/[^/]+/member/[^/]+`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)

			// Rich menu-related endpoints
			case regexp.MustCompile(`/v2/bot/richmenu/[^/]+`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/richmenu/[^/]+/content`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/user/[^/]+/richmenu`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)
			case regexp.MustCompile(`/v2/bot/user/[^/]+/richmenu/[^/]+`).MatchString(r.URL.Path):
				mock.handleDefaultSuccess(w, r)

			// Default response for any unhandled endpoint
			default:
				fmt.Printf("Unhandled request path: %s\n", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			}
		}
	}))

	return mock
}

// handleContentRequest handles requests for message content
func (m *mockLineServer) handleContentRequest(w http.ResponseWriter, r *http.Request, messageID string) {
	// Check if content exists for this message ID
	content, exists := m.messageContentMap[messageID]
	if !exists {
		fmt.Printf("Content not found for message ID: %s\n", messageID)
		fmt.Printf("Available message IDs: %v\n", getMapKeys(m.messageContentMap))
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	// Set content type
	contentType, exists := m.contentTypeMap[messageID]
	if exists {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	fmt.Printf("Serving content for message ID %s with type %s and length %d bytes\n",
		messageID, w.Header().Get("Content-Type"), len(content))

	// Write content
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// handleReplyRequest handles reply message requests
func (m *mockLineServer) handleReplyRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the reply request
	var replyRequest struct {
		ReplyToken string            `json:"replyToken"`
		Messages   []json.RawMessage `json:"messages"`
	}

	body, _ := io.ReadAll(r.Body)
	fmt.Printf("Reply request body: %s\n", string(body))

	if err := json.Unmarshal(body, &replyRequest); err != nil {
		fmt.Printf("Failed to parse reply request: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// For each message, try to parse it as a text message
	for _, msgJSON := range replyRequest.Messages {
		var textMsg struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}

		if err := json.Unmarshal(msgJSON, &textMsg); err == nil && textMsg.Type == "text" {
			m.repliesReceived = append(m.repliesReceived, linebot.NewTextMessage(textMsg.Text))
			fmt.Printf("Received reply message: %s\n", textMsg.Text)
		}
	}

	// Respond with success (as per LINE API documentation)
	m.handleDefaultSuccess(w, r)
}

// handleDefaultSuccess responds with a standard success response
func (m *mockLineServer) handleDefaultSuccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// addTestContent adds test content to the mock server
func (m *mockLineServer) addTestContent(messageID, contentType string, content []byte) {
	m.messageContentMap[messageID] = content
	m.contentTypeMap[messageID] = contentType
	fmt.Printf("Added test content for message ID %s with type %s and length %d bytes\n",
		messageID, contentType, len(content))
}

// close closes the mock server
func (m *mockLineServer) close() {
	m.server.Close()
}

// getEndpointURL returns the URL for the mock server
func (m *mockLineServer) getEndpointURL() string {
	return m.server.URL
}

// createSignature creates a valid LINE signature for a request
func createSignature(channelSecret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// setup sets up the test environment
func setup(t *testing.T) (*mockLineServer, *handler.WebhookHandler, *config.Config, *media.MediaStore, func()) {
	// Create a mock LINE server
	mockServer := newMockLineServer()

	// Set environment variable to point to the mock server
	os.Setenv("LINE_API_ENDPOINT", mockServer.getEndpointURL())

	// Create a test config
	cfg := &config.Config{
		ChannelSecret: testChannelSecret,
		ChannelToken:  testChannelToken,
		StorageDir:    testStorageDir,
		LogDir:        testLogDir,
		Debug:         true,
		Port:          "8080",
	}

	// Create directories if they don't exist
	os.MkdirAll(testStorageDir, 0755)
	os.MkdirAll(testLogDir, 0755)

	// Create a logger
	logger, err := utils.NewLogger(testLogDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create a LINE client
	lineClient, err := lineapi.NewClient(testChannelSecret, testChannelToken)
	if err != nil {
		t.Fatalf("Failed to create LINE client: %v", err)
	}

	// Create a media store
	mediaStore := media.NewMediaStore(cfg, logger)

	// Create a webhook handler
	webhookHandler := handler.NewWebhookHandler(lineClient, mediaStore, logger)

	// Return a cleanup function
	cleanup := func() {
		mockServer.close()
		logger.Close()
		os.RemoveAll(testStorageDir)
		os.Unsetenv("LINE_API_ENDPOINT")
	}

	return mockServer, webhookHandler, cfg, mediaStore, cleanup
}

// setupTestData creates the test data directory if it doesn't exist
func setupTestData(t *testing.T) {
	testDataDir := "../test_data"

	// Create test data directory if it doesn't exist
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// Create sample image if it doesn't exist
	imagePath := filepath.Join(testDataDir, "sample_image.jpg")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		// Create a minimal JPEG file
		file, err := os.Create(imagePath)
		if err != nil {
			t.Fatalf("Failed to create sample image: %v", err)
		}
		// Write minimal JPEG header
		file.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xFF, 0xD9})
		file.Close()
	}

	// Create sample video if it doesn't exist
	videoPath := filepath.Join(testDataDir, "sample_video.mp4")
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		// Create a minimal MP4 file
		file, err := os.Create(videoPath)
		if err != nil {
			t.Fatalf("Failed to create sample video: %v", err)
		}
		// Write minimal MP4 header (FTYP box)
		file.Write([]byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32, 0x00, 0x00, 0x00, 0x00, 0x6D, 0x70, 0x34, 0x32, 0x69, 0x73, 0x6F, 0x6D})
		file.Close()
	}
}

// TestWebhookHandlerWithImageMessage tests the webhook handler with an image message
func TestWebhookHandlerWithImageMessage(t *testing.T) {
	// Set up test data
	setupTestData(t)

	// Set up the test environment
	mockServer, webhookHandler, _, mediaStore, cleanup := setup(t)
	defer cleanup()

	// Read the sample image file
	imageID := "image123"
	imageContent, err := os.ReadFile("../test_data/sample_image.jpg")
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	// Add test content to the mock server
	mockServer.addTestContent(imageID, "image/jpeg", imageContent)

	// Create a webhook request with an image message
	webhookRequest := createImageMessageWebhook(imageID)
	body, _ := json.Marshal(webhookRequest)

	// Create a signature
	signature := createSignature(testChannelSecret, body)

	// Create a test HTTP request
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	res := httptest.NewRecorder()

	// Handle the request
	webhookHandler.HandleWebhook(res, req)

	// Check the response
	if res.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
	}

	// Wait for any downloads to complete
	mediaStore.WaitForDownloads()

	// Verify that a file was saved
	// Note: We can't check the exact filename as it's generated with a UUID
	filesExist := false
	currentDate := time.Now().Format("2006-01-02")
	dateDir := filepath.Join(testStorageDir, currentDate)

	// Check if the date directory was created
	if _, err := os.Stat(dateDir); err == nil {
		// List files in the directory
		files, err := os.ReadDir(dateDir)
		if err == nil && len(files) > 0 {
			for _, file := range files {
				if strings.Contains(file.Name(), "image_") {
					filesExist = true
					break
				}
			}
		}
	}

	if !filesExist {
		t.Errorf("Expected image file to be saved in %s", dateDir)
	}

	// Check if a reply was sent
	if len(mockServer.repliesReceived) == 0 {
		t.Errorf("Expected a reply message to be sent")
	} else {
		// Check if the reply contains the expected text
		textMsg, ok := mockServer.repliesReceived[0].(*linebot.TextMessage)
		if !ok {
			t.Errorf("Expected a text message reply")
		} else if !strings.Contains(textMsg.Text, "image") {
			t.Errorf("Expected reply to mention 'image', got: %s", textMsg.Text)
		}
	}
}

// TestWebhookHandlerWithVideoMessage tests the webhook handler with a video message
func TestWebhookHandlerWithVideoMessage(t *testing.T) {
	// Set up test data
	setupTestData(t)

	// Set up the test environment
	mockServer, webhookHandler, _, mediaStore, cleanup := setup(t)
	defer cleanup()

	// Read the sample video file
	videoID := "video456"
	videoContent, err := os.ReadFile("../test_data/sample_video.mp4")
	if err != nil {
		t.Fatalf("Failed to read test video: %v", err)
	}

	// Add test content to the mock server
	mockServer.addTestContent(videoID, "video/mp4", videoContent)

	// Create a webhook request with a video message
	webhookRequest := createVideoMessageWebhook(videoID)
	body, _ := json.Marshal(webhookRequest)

	// Create a signature
	signature := createSignature(testChannelSecret, body)

	// Create a test HTTP request
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	res := httptest.NewRecorder()

	// Handle the request
	webhookHandler.HandleWebhook(res, req)

	// Check the response
	if res.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
	}

	// Wait for any downloads to complete
	mediaStore.WaitForDownloads()

	// Verify that a file was saved
	filesExist := false
	currentDate := time.Now().Format("2006-01-02")
	dateDir := filepath.Join(testStorageDir, currentDate)

	// Check if the date directory was created
	if _, err := os.Stat(dateDir); err == nil {
		// List files in the directory
		files, err := os.ReadDir(dateDir)
		if err == nil && len(files) > 0 {
			for _, file := range files {
				if strings.Contains(file.Name(), "video_") {
					filesExist = true
					break
				}
			}
		}
	}

	if !filesExist {
		t.Errorf("Expected video file to be saved in %s", dateDir)
	}

	// Check if a reply was sent
	if len(mockServer.repliesReceived) == 0 {
		t.Errorf("Expected a reply message to be sent")
	} else {
		// Check if the reply contains the expected text
		textMsg, ok := mockServer.repliesReceived[0].(*linebot.TextMessage)
		if !ok {
			t.Errorf("Expected a text message reply")
		} else if !strings.Contains(textMsg.Text, "video") {
			t.Errorf("Expected reply to mention 'video', got: %s", textMsg.Text)
		}
	}
}

// TestWebhookHandlerWithInvalidSignature tests the webhook handler with an invalid signature
func TestWebhookHandlerWithInvalidSignature(t *testing.T) {
	// Set up test data
	setupTestData(t)

	// Set up the test environment
	_, webhookHandler, _, _, cleanup := setup(t)
	defer cleanup()

	// Create a webhook request
	webhookRequest := createImageMessageWebhook("image123")
	body, _ := json.Marshal(webhookRequest)

	// Create an invalid signature
	signature := "invalid_signature"

	// Create a test HTTP request
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Line-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	res := httptest.NewRecorder()

	// Handle the request
	webhookHandler.HandleWebhook(res, req)

	// Check the response - should be 400 Bad Request due to invalid signature
	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, res.Code)
	}
}

// Helper function to create a webhook request with an image message
func createImageMessageWebhook(imageID string) map[string]interface{} {
	return map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"type":       "message",
				"replyToken": "reply123",
				"source": map[string]interface{}{
					"type":   "user",
					"userId": "user123",
				},
				"timestamp": time.Now().Unix() * 1000,
				"message": map[string]interface{}{
					"id":   imageID,
					"type": "image",
				},
			},
		},
	}
}

// Helper function to create a webhook request with a video message
func createVideoMessageWebhook(videoID string) map[string]interface{} {
	return map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"type":       "message",
				"replyToken": "reply456",
				"source": map[string]interface{}{
					"type":   "user",
					"userId": "user456",
				},
				"timestamp": time.Now().Unix() * 1000,
				"message": map[string]interface{}{
					"id":   videoID,
					"type": "video",
				},
			},
		},
	}
}
