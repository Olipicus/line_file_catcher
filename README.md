# LineFileCatcher

LineFileCatcher is a Go-based service that receives image and video files from LINE messages via a webhook and stores them in a local directory. The service organizes saved files into directories based on the date of receipt and ensures unique filenames to prevent overwriting.

## Features

- Webhook server for receiving LINE message events
- Automatic download and storage of media files (images, videos, audio, and other files)
- Concurrent processing of file downloads for improved performance
- Organization of files into date-based directories
- Unique filename generation to prevent overwriting
- Graceful shutdown to ensure all pending downloads complete
- Health check endpoint for monitoring service status
- Comprehensive logging system

## Prerequisites

- Go 1.16 or higher
- A LINE Bot account with channel secret and access token
- A publicly accessible URL for webhook callbacks (you can use ngrok for development)

## Installation

### Option 1: From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/line_file_catcher.git
   cd line_file_catcher
   ```

2. Copy the example environment file and update with your LINE credentials:
   ```bash
   cp .env.example .env
   # Edit .env with your LINE channel secret and token
   ```

3. Build and run the application:
   ```bash
   go build -o linefilecatcher ./cmd/linefilecatcher
   ./linefilecatcher
   ```

### Option 2: Using Docker

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/line_file_catcher.git
   cd line_file_catcher
   ```

2. Build the Docker image:
   ```bash
   docker build -t line-file-catcher .
   ```

3. Run the Docker container:
   ```bash
   docker run -p 8080:8080 \
     -e LINE_CHANNEL_SECRET=your_channel_secret \
     -e LINE_CHANNEL_TOKEN=your_channel_token \
     -v $(pwd)/storage:/app/storage \
     -v $(pwd)/logs:/app/logs \
     line-file-catcher
   ```

## Configuration

LineFileCatcher can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| LINE_CHANNEL_SECRET | Your LINE channel secret | (required) |
| LINE_CHANNEL_TOKEN | Your LINE channel access token | (required) |
| PORT | Port for the webhook server | 8080 |
| STORAGE_DIR | Directory where files will be stored | ./storage |
| LOG_DIR | Directory where logs will be stored | ./logs |
| DEBUG | Enable debug logging | false |

## Setting Up Your LINE Bot

1. Create a LINE Developer account and create a new provider and channel at [LINE Developers Console](https://developers.line.biz/console/)

2. Create a new Messaging API channel:
   - Go to the Providers List and select your provider
   - Click "Create a new channel"
   - Select "Messaging API"
   - Fill in the required information and create the channel

3. Configure your channel settings:
   - From your channel page, go to the "Messaging API" tab
   - Scroll down to "Webhook settings" and enable "Use webhook"
   - Set the Webhook URL to `https://your-domain.com/webhook`
   - Under "Bot settings", issue a Channel Access Token with proper permissions

4. Set necessary permissions:
   - Turn on "Allow bot to join group chats"
   - Enable "Use webhook" in Webhook settings
   - Set the appropriate permission for your bot

5. Test your webhook with the LINE webhook simulator:
   - In the LINE Developers Console, use the webhook testing tool to send test events to your service

## Using the Service

Once the service is set up:

1. Users can send images, videos, and files to your LINE bot
2. The service will automatically save these files to the configured storage directory
3. Files are organized by date in the format `YYYY-MM-DD/`
4. Each file has a unique name containing the media type, timestamp, and random string to prevent collisions

### Health Checking

The service provides a health check endpoint at `/health` that returns JSON with service status information:

```
GET http://your-server:8080/health
```

The response includes uptime, memory usage, and other diagnostics information.

## Directory Structure

Files are saved in the following structure:

```
storage/
  ├── YYYY-MM-DD/
  │   ├── image_timestamp_randomString.jpg
  │   ├── video_timestamp_randomString.mp4
  │   └── ...
  └── ...
```

## Development

The project follows a standard Go project layout:

```
line_file_catcher/
├── cmd/
│   └── linefilecatcher/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration loading and management
│   ├── handler/              # HTTP handlers for webhook endpoints
│   ├── lineapi/              # Interactions with LINE Messaging API
│   ├── media/                # Media processing and storage logic
│   └── utils/                # Utility functions
├── pkg/                      # Exportable packages (if any)
├── scripts/                  # Helper scripts (e.g., for setup or deployment)
├── test/                     # Integration and end-to-end tests
├── .env                      # Environment variables (excluded from version control)
├── .gitignore
├── Dockerfile                # Docker configuration
├── go.mod
├── go.sum
└── README.md
```

### Testing

The project includes integration testing tools:

1. `test_integration.sh` - Runs the service with a mock LINE API server for testing
2. `mock_content_api.go` - Mock implementation of LINE's Content API for testing
3. `webhook_simulator.go` - Simulates LINE webhook events for testing

Run the integration tests with:

```bash
./scripts/test_integration.sh
```

## Logs

Logs are stored in the configured log directory with the naming pattern `linefilecatcher_YYYY-MM-DD.log`. 
When debug mode is enabled, more detailed logs are generated.

## Troubleshooting

Common issues:

1. **Webhook validation errors**: Ensure your LINE Channel Secret is correct and the webhook URL is publicly accessible

2. **File access errors**: Check that the application has write permissions to the storage directory

3. **Missing media files**: Verify that your LINE Bot has the necessary permissions to access message content

## Disclaimer

This project was developed with assistance from GitHub Copilot, an AI-powered code generation tool. Please be aware of the following:

- The code generated with AI assistance may vary in quality and completeness
- This software is provided "as is", without warranty of any kind, express or implied
- No guarantees are made regarding the reliability, security, or performance of this application
- Users should thoroughly test and review the code before implementing in production environments
- The maintainers and contributors are not responsible for any issues that may arise from using this software
- LINE's APIs and services may change over time, potentially affecting this application's functionality

While every effort has been made to ensure the code is functional and follows best practices, AI-generated code may contain unexpected issues or edge cases that haven't been addressed. Users are encouraged to review, test, and modify the code to suit their specific requirements.

## License

[MIT License](LICENSE)