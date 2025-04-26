#!/bin/bash

# This script runs the mock LINE content API and the LineFileCatcher service for testing

# Function to cleanup on exit
cleanup() {
    echo "Cleaning up..."
    kill $MOCK_API_PID $SERVICE_PID 2>/dev/null
    exit
}

# Set up trap to ensure cleanup
trap cleanup INT TERM EXIT

# Ensure storage directory exists
mkdir -p storage

# Build the application
echo "Building LineFileCatcher..."
go build -o bin/linefilecatcher cmd/linefilecatcher/main.go

if [ $? -ne 0 ]; then
    echo "Failed to build application"
    exit 1
fi

# Start the mock LINE content API server
echo "Starting mock LINE API server..."
go run test/mock_content_api.go 9000 &
MOCK_API_PID=$!

# Wait for the mock API to start
echo "Waiting for mock API to start..."
sleep 2

# Create .env file for testing if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating test .env file..."
    cat > .env << EOF
LINE_CHANNEL_SECRET=test_channel_secret
LINE_CHANNEL_TOKEN=test_channel_token
PORT=8080
STORAGE_DIR=./storage
LOG_DIR=./logs
DEBUG=true
LINE_API_ENDPOINT=http://localhost:9000/v2/bot
EOF
    echo "Created test .env file"
fi

# Override the LINE API endpoint for testing
export LINE_API_ENDPOINT=http://localhost:9000/v2/bot

# Start the LineFileCatcher service
echo "Starting LineFileCatcher service..."
./bin/linefilecatcher &
SERVICE_PID=$!

# Wait for the service to start
echo "Waiting for service to start..."
sleep 2

# Run the webhook simulator
echo "Sending test webhook request..."
go run test/webhook_simulator.go

# Keep services running to allow manual testing
echo ""
echo "Services are running for manual testing:"
echo "- Mock LINE API server: http://localhost:9000"
echo "- LineFileCatcher service: http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop and cleanup"

# Wait for interrupt
wait