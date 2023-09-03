package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TranscriptionResponse to map the JSON response
type TranscriptionResponse struct {
	Transcription string `json:"transcription"`
}

func TestTranscribeHandler(t *testing.T) {
	// Prepare a form that you will submit to the URL.
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	formFile, err := writer.CreateFormFile("file", "test.wav")
	if err != nil {
		t.Errorf("CreateFormFile failed: %v", err)
		return
	}

	// Read the test audio file from disk and write to form
	filePath := "testdata/test.wav"
	file, err := os.Open(filePath)
	if err != nil {
		t.Errorf("Could not open file: %v", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(formFile, file)
	if err != nil {
		t.Errorf("Could not write to form: %v", err)
		return
	}

	writer.Close()

	// Create a request to use against our handler.
	req, err := http.NewRequest("POST", "/transcribe", &buffer)
	if err != nil {
		t.Fatal(err)
	}

	// Setting header and API key
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", "your_test_api_key_here")

	// Create a ResponseRecorder to record the response.
	recorder := httptest.NewRecorder()

	// Create an HTTP handler from our handler function.
	handler := http.HandlerFunc(apiKeyMiddleware(transcribeHandler))

	// Serve the HTTP request to our recorder.
	handler.ServeHTTP(recorder, req)

	// Check the status code is what we expect.
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Parse the JSON response
	var resp TranscriptionResponse
	err = json.NewDecoder(recorder.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Check the transcription, this will depend on the audio file used for testing
	if resp.Transcription != "expected_transcription_here" {
		t.Errorf("Unexpected transcription: got %v want %v",
			resp.Transcription, "expected_transcription_here")
	}
}
