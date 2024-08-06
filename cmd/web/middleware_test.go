package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"snippetbox.i4o.dev/internal/assert"
	"testing"
)

func TestCommonHeaders(t *testing.T) {
	responseRecorder := httptest.NewRecorder()

	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("OK"))
	})
	commonHeaders(next).ServeHTTP(responseRecorder, request)
	response := responseRecorder.Result()

	expectedValue := "default-src 'self'; style-src 'self' 'unsafe-inline' fonts.bunny.net; font-src fonts.bunny.net;"
	assert.Equal(t, response.Header.Get("Content-Security-Policy"), expectedValue)

	expectedValue = "origin-when-cross-origin"
	assert.Equal(t, response.Header.Get("Referrer-Policy"), expectedValue)

	expectedValue = "nosniff"
	assert.Equal(t, response.Header.Get("X-Content-Type-Options"), expectedValue)

	expectedValue = "deny"
	assert.Equal(t, response.Header.Get("X-Frame-Options"), expectedValue)

	expectedValue = "0"
	assert.Equal(t, response.Header.Get("X-XSS-Protection"), expectedValue)

	expectedValue = "Go"
	assert.Equal(t, response.Header.Get("Server"), expectedValue)

	assert.Equal(t, response.StatusCode, http.StatusOK)

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	assert.Equal(t, string(body), "OK")
}
