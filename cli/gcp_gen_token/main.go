package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func main() {
	// Read credentials from file
	credentialsPath := "./credentials.json" // Update this path if needed
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Configure the OAuth2 config
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Generate an authentication URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n", authURL)

	// Get the authorization code from user input
	var authCode string
	fmt.Print("Enter the authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	// Exchange auth code for token
	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	// Save the token
	tokenPath := "./token.json" // Update this path if needed
	fmt.Printf("Saving token to: %s\n", tokenPath)

	// Ensure directory exists
	tokenDir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		log.Fatalf("Unable to create token directory: %v", err)
	}

	// Write token to file
	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()

	// Encode token to JSON
	if err := json.NewEncoder(f).Encode(tok); err != nil {
		log.Fatalf("Unable to encode token to JSON: %v", err)
	}

	fmt.Println("Token successfully generated and saved!")
}
