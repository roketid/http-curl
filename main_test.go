package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Add the same routes as main
	e.POST("/curl", handleCurl)
	e.Any("/waiting/:milli", func(c echo.Context) error {
		milliStr := c.Param("milli")
		milli, err := strconv.Atoi(milliStr)
		if err != nil || milli < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid milliseconds"})
		}

		// Convert milliseconds to duration
		time.Sleep(time.Duration(milli) * time.Millisecond)

		return c.String(http.StatusOK, "Ok")
	})

	return e
}

func TestHandleCurl_ValidRequest(t *testing.T) {
	e := setupTestServer()

	// Test data
	requestData := map[string]interface{}{
		"--location": "https://httpbin.org/get",
	}

	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/curl", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "result")
}

func TestHandleCurl_InvalidContentType(t *testing.T) {
	e := setupTestServer()

	requestData := map[string]interface{}{
		"--location": "https://httpbin.org/get",
	}

	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/curl", bytes.NewBuffer(jsonData))
	// Don't set Content-Type header
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Content-Type must be application/json")
}

func TestHandleCurl_InvalidJSON(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/curl", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid JSON input")
}

func TestHandleCurl_WithTimeout(t *testing.T) {
	e := setupTestServer()

	requestData := map[string]interface{}{
		"--location": "https://httpbin.org/delay/1",
	}

	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/curl?timeout=5s", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleCurl_InvalidTimeout(t *testing.T) {
	e := setupTestServer()

	requestData := map[string]interface{}{
		"--location": "https://httpbin.org/get",
	}

	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/curl?timeout=invalid", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Error parsing timeout duration")
}

func TestHandleCurl_Base64Response(t *testing.T) {
	e := setupTestServer()

	requestData := map[string]interface{}{
		"--location": "https://httpbin.org/json",
	}

	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/curl?base64=true", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "result")
	// The result should be base64 encoded
	assert.NotEmpty(t, response["result"])
}

func TestHandleCurl_PlainTextResponse(t *testing.T) {
	e := setupTestServer()

	requestData := map[string]interface{}{
		"--location": "https://httpbin.org/plain",
	}

	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/curl?plain=true", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain", rec.Header().Get("Content-Type"))
}

func TestHandleCurl_POSTRequest(t *testing.T) {
	e := setupTestServer()

	requestData := map[string]interface{}{
		"-X":         "POST",
		"-d":         `{"test": "data"}`,
		"-H":         "Content-Type: application/json",
		"--location": "https://httpbin.org/post",
	}

	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/curl", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "result")
}

func TestWaitingEndpoint_ValidMilliseconds(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/waiting/100", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	e.ServeHTTP(rec, req)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok", rec.Body.String())
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}

func TestWaitingEndpoint_InvalidMilliseconds(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/waiting/invalid", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid milliseconds")
}

func TestWaitingEndpoint_NegativeMilliseconds(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/waiting/-100", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid milliseconds")
}

func TestWaitingEndpoint_ZeroMilliseconds(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/waiting/0", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Ok", rec.Body.String())
}

// Benchmark tests
func BenchmarkHandleCurl_SimpleRequest(b *testing.B) {
	e := setupTestServer()

	requestData := map[string]interface{}{
		"--location": "https://httpbin.org/get",
	}

	jsonData, _ := json.Marshal(requestData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/curl", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
	}
}

func BenchmarkWaitingEndpoint(b *testing.B) {
	e := setupTestServer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/waiting/10", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
	}
}
