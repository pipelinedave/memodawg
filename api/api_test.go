package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTranscribeHandler(t *testing.T) {
	payload := []byte(`{"locale": "en-US", "data": "test"}`)
	req, err := http.NewRequest("POST", "/transcribe", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(transcribeHandler)

	handler.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", res.Status)
	}

	data, _ := ioutil.ReadAll(res.Body)

	if data == nil {
		t.Errorf("Expected non-empty response")
	}
}
