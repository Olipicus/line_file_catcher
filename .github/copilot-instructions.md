# GitHub Copilot Custom Instructions for LineFileCatcher

- Use the Go programming language for all code examples.
- Utilize the official LINE Messaging API SDK for Go.
- Implement a webhook server to receive image and video files from LINE messages.
- Save received media files to local storage with unique filenames to prevent overwriting.
- Organize saved files into directories based on the date of receipt.
- Ensure the server can handle concurrent requests efficiently.
- Include error handling for file I/O and network operations.
- Provide comments for complex code sections to enhance readability.
- Follow Go's standard formatting and naming conventions.
- Implement Google Drive integration to backup received media files to the cloud.
- Use Google Drive API with OAuth 2.0 authentication for secure cloud storage.
- Add configurable options to enable/disable Google Drive backup.
- Implement retry mechanisms for failed Google Drive uploads.
- Create a folder structure in Google Drive that mirrors the local storage organization.
- Include logging of successful and failed Google Drive operations.


line_file_catcher/
├── cmd/
│   └── linefilecatcher/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration loading and management
│   ├── handler/              # HTTP handlers for webhook endpoints
│   ├── lineapi/              # Interactions with LINE Messaging API
│   ├── media/                # Media processing and storage logic
│   ├── cloud/
│   │   ├── drive/            # Google Drive integration components
│   │   └── common/           # Common cloud storage interfaces and utilities
│   └── utils/                # Utility functions
├── pkg/                      # Exportable packages (if any)
├── scripts/                  # Helper scripts (e.g., for setup or deployment)
├── test/                     # Integration and end-to-end tests by do component test
├── .env                      # Environment variables (excluded from version control)
├── .gitignore
├── Dockerfile                # Docker configuration
├── go.mod
├── go.sum
└── README.md
