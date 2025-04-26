package drive

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// GenerateToken creates a token for Google Drive API access
func GenerateToken(credentialsFile, tokenFile string) error {
	// Read the credentials file
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	// Parse the credentials
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file: %v", err)
	}

	// Generate an authorization URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser and authorize the application:\n%v\n", authURL)
	fmt.Println("Enter the authorization code:")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return fmt.Errorf("unable to read authorization code: %v", err)
	}

	// Exchange the auth code for a token
	token, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web: %v", err)
	}

	// Save the token to a file
	fmt.Printf("Saving token to: %s\n", tokenFile)
	f, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}

// StartTokenServer is a utility to help generate a token for Google Drive integration
// You can run this as a standalone program or call it from your main application
func StartTokenServer() {
	log.Println("Running Google Drive Token Server...")
	log.Println("Please provide the following environment variables:")
	log.Println("- DRIVE_CREDENTIALS: path to credentials.json file")
	log.Println("- DRIVE_TOKEN_FILE: path where token will be saved")

	credentialsFile := os.Getenv("DRIVE_CREDENTIALS")
	if credentialsFile == "" {
		credentialsFile = "credentials.json"
		log.Printf("DRIVE_CREDENTIALS not set, using default: %s", credentialsFile)
	}

	tokenFile := os.Getenv("DRIVE_TOKEN_FILE")
	if tokenFile == "" {
		tokenFile = "token.json"
		log.Printf("DRIVE_TOKEN_FILE not set, using default: %s", tokenFile)
	}

	if err := GenerateToken(credentialsFile, tokenFile); err != nil {
		log.Fatalf("Error generating token: %v", err)
	}

	log.Println("Token generated successfully!")
}
