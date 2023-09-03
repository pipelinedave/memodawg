package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"fmt"
)

// Create a struct that matches the JSON response structure
type AzureResponse struct {
	RecognitionStatus string `json:"RecognitionStatus"`
	Offset            int64  `json:"Offset"`
	Duration          int64  `json:"Duration"`
	DisplayText       string `json:"DisplayText"`
}

const tempAudioFile = "audio.wav"

func main() {
	log.Println("Starting memodawg-api...")
	http.HandleFunc("/transcribe", transcribeHandler)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func transcribeHandler(w http.ResponseWriter, r *http.Request) {
	gotifyToken := os.Getenv("GOTIFY_TOKEN")
	azureSubscriptionKey := os.Getenv("AZURE_KEY")
	azureSTTURL := os.Getenv("AZURE_STT_URL")
	azureTokenURL := os.Getenv("AZURE_TOKEN_URL")

	log.Println("Received request for transcription.")

	// Log only the headers and other metadata, not the body
	log.Printf("Request Method: %s\n", r.Method)
	log.Printf("Request URL: %s\n", r.URL.String())
	log.Printf("Request Headers: %v\n", r.Header)

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Println("No file provided:", err)
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save the file
	out, err := os.Create(tempAudioFile)
	if err != nil {
		log.Println("Error while creating audio file:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("Error while copying audio data:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get Azure token
	token, err := getAzureToken(azureSubscriptionKey, azureTokenURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transcribe using Azure
	transcription, err := transcribeWithAzure(token, azureSTTURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send Gotify notification
	if err := sendGotifyNotification(transcription, gotifyToken); err != nil {
		log.Printf("Failed to send Gotify notification: %s", err.Error())
	}

	// Return the transcription
	log.Println("Successfully processed the transcription.")
	json.NewEncoder(w).Encode(map[string]string{"transcription": transcription})
}

func getAzureToken(azureSubscriptionKey, azureTokenURL string) (string, error) {
	// Log the start of the function
	log.Println("Starting getAzureToken function...")

	// Create a new HTTP request to acquire the Azure token
	log.Printf("Creating a new POST request to %s", azureTokenURL)
	req, err := http.NewRequest("POST", azureTokenURL, nil)
	if err != nil {
		log.Printf("Failed to create a POST request: %v", err)
		return "", err
	}

	// Set the Azure subscription key in the request header
	req.Header.Set("Ocp-Apim-Subscription-Key", azureSubscriptionKey)
	log.Println("Added Azure subscription key to the request header.")

	// Perform the HTTP request
	log.Println("Performing the HTTP request...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Validate the HTTP response status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK HTTP status: %d", resp.StatusCode)
		return "", fmt.Errorf("received non-OK HTTP status: %d", resp.StatusCode)
	}

	// Read the HTTP response body
	log.Println("Reading the HTTP response body...")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read the response body: %v", err)
		return "", err
	}

	// DEBUG: Log the response body
	log.Printf("HTTP Response Body: %s", string(body))

	// As the response appears to be a JWT, no need for JSON deserialization
	token := string(body)
	if token == "" {
		log.Println("Token is empty.")
		return "", fmt.Errorf("Token is empty")
	}

	// Log the completion of the function (avoid logging the actual token in production)
	log.Println("Azure access token acquired successfully.")
	return token, nil
}



func transcribeWithAzure(token, azureSTTURL string) (string, error) {
	// Log the start of the function
	log.Println("Starting transcribeWithAzure function...")

	// Read audio file data
	log.Printf("Reading audio file from %s...", tempAudioFile)
	fileData, err := ioutil.ReadFile(tempAudioFile)
	if err != nil {
		log.Printf("Failed to read audio file: %s", err)
		return "", err
	}

	// Specify desired locale
	locale := "de-DE"
	fullURL := fmt.Sprintf("%s?language=%s", azureSTTURL, locale)

	// Create a new HTTP request for Azure Speech-to-Text
	log.Printf("Creating a new POST request to %s", fullURL)
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(fileData))
	if err != nil {
		log.Printf("Failed to create POST request: %s", err)
		return "", err
	}

	// Set headers: Authorization and Content-Type
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "audio/wav")
	log.Println("Added Authorization and Content-Type headers to the request.")

	// Perform the HTTP request
	log.Println("Performing the HTTP request...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("HTTP request failed: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the HTTP response body
	log.Println("Reading the HTTP response body...")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read the response body: %s", err)
		return "", err
	}

	// DEBUG: Log the response body
	log.Printf("HTTP Response Body: %s", string(body))

	// Deserialize JSON response body into a custom struct
	log.Println("Deserializing the JSON response body...")
	var data AzureResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Failed to deserialize JSON: %s", err)
		return "", err
	}

	// Retrieve the DisplayText field from the struct
	transcription := data.DisplayText
	if transcription == "" {
		log.Println("DisplayText field is missing or empty in the JSON response.")
		return "", fmt.Errorf("DisplayText field is missing or empty in the JSON response")
	}

	// Log the transcription and return it
	log.Printf("Transcription completed: %s", transcription)
	return transcription, nil
}


func sendGotifyNotification(message, gotifyToken string) error {
	// Log the start of the function
	log.Println("Starting sendGotifyNotification function...")

	// Retrieve the Gotify URL from environment variables
	gotifyBaseURL := os.Getenv("GOTIFY_URL")
	gotifyURL := fmt.Sprintf("%s?token=%s", gotifyBaseURL, gotifyToken)
	log.Printf("Full Gotify URL: %s", gotifyURL)

	// Prepare the payload for Gotify
	log.Println("Preparing the payload...")
	payload := map[string]interface{}{
		"message":  message,
		"title":    "Voice Memo Transcription",
		"priority": 10,
	}

	// Serialize the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to serialize payload to JSON: %s", err)
		return err
	}

	// Create a new HTTP request to send the notification
	log.Println("Creating a new POST request to Gotify...")
	req, err := http.NewRequest("POST", gotifyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create POST request: %s", err)
		return err
	}

	// Set headers for the Gotify request
	req.Header.Set("Content-Type", "application/json")
	log.Println("Added Content-Type headers to the request.")

	// Perform the HTTP request to send the notification
	log.Println("Performing the HTTP request to Gotify...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("HTTP request to Gotify failed: %s", err)
		return err
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read the response body: %s", err)
		return err
	}
	log.Printf("HTTP Response Body: %s", string(body))

	// Check if Gotify responded with a success status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Gotify responded with non-OK status: %d", resp.StatusCode)
		return fmt.Errorf("Gotify responded with non-OK status: %d", resp.StatusCode)
	}

	// Successfully sent the notification
	log.Println("Successfully sent the Gotify notification.")
	return nil
}


