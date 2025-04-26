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
├── test/                     # Integration and end-to-end tests by do component test
├── .env                      # Environment variables (excluded from version control)
├── .gitignore
├── Dockerfile                # Docker configuration
├── go.mod
├── go.sum
└── README.md
