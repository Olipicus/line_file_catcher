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

## Google Drive Integration

LineFileCatcher can automatically backup your media files to Google Drive. This feature is disabled by default and requires additional setup.

### Prerequisites for Google Drive Integration

- A Google account
- A Google Cloud project with the Google Drive API enabled
- OAuth 2.0 credentials for authentication

### Setting Up Google Drive Integration

#### 1. Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Create Project" or select an existing project
3. Give your project a name and click "Create"

#### 2. Enable the Google Drive API

1. In your Google Cloud project, go to "APIs & Services" > "Library"
2. Search for "Google Drive API" and select it
3. Click "Enable"

#### 3. Create OAuth 2.0 Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" and select "OAuth client ID"
3. If this is your first time, you'll need to configure the OAuth consent screen:
   - Choose "External" or "Internal" depending on your use case
   - Fill in the required information (app name, user support email, etc.)
   - Add the necessary scopes (at minimum, `.../auth/drive.file`)
   - Add test users if using External
   - Save and continue
4. For the OAuth client ID:
   - Select "Desktop app" as the application type
   - Give your client a name
   - Click "Create"
5. Download the JSON credentials file by clicking the download icon
6. Move this file to `./bin/credentials.json` in your project directory (or update the path in your `.env` file)

#### 4. Generate the Google Drive Access Token

The project includes a utility to generate the necessary access token for Google Drive:

1. Navigate to the utility directory:
   ```bash
   cd cli/gcp_gen_token
   ```

2. Run the token generation utility:
   ```bash
   go run main.go
   ```
   
   If your credentials file is in a different location, you can edit the path in the utility or copy it to the expected location.

3. The utility will output a URL. Copy and paste this URL in your browser.

4. Sign in with your Google account and grant the requested permissions.

5. After authentication, you'll receive an authorization code. Copy this code.

6. Return to the terminal and paste the code when prompted.

7. The utility will generate a `token.json` file and save it to the current directory.

8. Move this token file to `./bin/token.json` (or update the path in your `.env` file).

#### 5. Configure the Environment Variables

Edit your `.env` file to enable Google Drive integration:

```
# Google Drive Integration
DRIVE_ENABLED=true
DRIVE_CREDENTIALS=./bin/credentials.json
DRIVE_TOKEN_FILE=./bin/token.json
DRIVE_FOLDER=LineFileCatcher
DRIVE_RETRY_COUNT=3
```

### How It Works

When enabled:

1. Files saved locally will also be uploaded to Google Drive
2. The same directory structure (organized by date) will be maintained in Google Drive
3. Files are uploaded asynchronously to avoid slowing down the response times
4. Failed uploads will be retried according to the configured retry count
5. Detailed logs of upload success/failure are maintained

### Troubleshooting Google Drive Integration

Common issues:

1. **Authentication errors**: Ensure your credentials.json and token.json files are valid and accessible to the application

2. **Permission errors**: Check that your Google account has the necessary permissions and the correct scopes were requested

3. **Expired tokens**: Token may expire after some time. If you experience authentication issues, regenerate the token using the included utility

4. **Rate limiting**: Google Drive API has quotas. Check the logs for any rate limiting errors if you're processing many files

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