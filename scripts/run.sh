#!/bin/bash

# Create storage directory if it doesn't exist
mkdir -p storage

# Run the application
cd "$(dirname "$0")" || exit
go run cmd/linefilecatcher/main.go