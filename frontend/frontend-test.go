package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFormHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(formHandler)

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Upload Audio for Transcription" // part of the expected HTML content
	actual := recorder.Body.String()

	if !contains(actual, expected) {
		t.Errorf("Handler returned unexpected body: got %v want %v", actual, expected)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
