package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"snippetbox.i4o.dev/internal/assert"
	"testing"
)

func TestPing(t *testing.T) {
	responseRecorder := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	ping(responseRecorder, request)
	response := responseRecorder.Result()
	assert.Equal(t, response.StatusCode, http.StatusOK)

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	assert.Equal(t, string(body), "OK")
}
