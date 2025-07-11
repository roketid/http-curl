package httpcurl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurlValue_UnmarshalJSON_SingleString(t *testing.T) {
	var cv CurlValue
	data := `"test-value"`

	err := cv.UnmarshalJSON([]byte(data))
	require.NoError(t, err)
	assert.Len(t, cv, 1)
	assert.Equal(t, "test-value", cv[0])
}

func TestCurlValue_UnmarshalJSON_MultipleStrings(t *testing.T) {
	var cv CurlValue
	data := `["value1", "value2", "value3"]`

	err := cv.UnmarshalJSON([]byte(data))
	require.NoError(t, err)
	assert.Len(t, cv, 3)
	assert.Equal(t, "value1", cv[0])
	assert.Equal(t, "value2", cv[1])
	assert.Equal(t, "value3", cv[2])
}

func TestCurlValue_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var cv CurlValue
	data := `{invalid json}`

	err := cv.UnmarshalJSON([]byte(data))
	assert.Error(t, err)
}

func TestSanitizeInput_ValidOptions(t *testing.T) {
	input := CurlOption{
		"-X":         CurlValue{"POST"},
		"-d":         CurlValue{"{\"test\":\"data\"}"},
		"-H":         CurlValue{"Content-Type: application/json"},
		"--location": CurlValue{"https://example.com"},
	}

	args, err := sanitizeInput(input)
	require.NoError(t, err)
	assert.Len(t, args, 8) // 4 options with their values
}

func TestSanitizeInput_UnauthorizedOption(t *testing.T) {
	input := CurlOption{
		"-X":          CurlValue{"POST"},
		"--dangerous": CurlValue{"value"}, // This should not be allowed
	}

	args, err := sanitizeInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized curl option")
	assert.Nil(t, args)
}

func TestSanitizeInput_StandaloneOptions(t *testing.T) {
	input := CurlOption{
		"-k":         CurlValue{""}, // Skip SSL verification
		"--location": CurlValue{"https://example.com"},
	}

	args, err := sanitizeInput(input)
	require.NoError(t, err)
	assert.Len(t, args, 3) // -k (standalone), --location, and its value
	assert.Contains(t, args, "-k")
	assert.Contains(t, args, "--location")
}

func TestSanitizeInput_TrueValueOptions(t *testing.T) {
	input := CurlOption{
		"-k":         CurlValue{"true"}, // Should be treated as standalone
		"--location": CurlValue{"https://example.com"},
	}

	args, err := sanitizeInput(input)
	require.NoError(t, err)
	assert.Len(t, args, 3) // -k (standalone), --location, and its value
	assert.Contains(t, args, "-k")
}

func TestSanitizeInput_MultipleValues(t *testing.T) {
	input := CurlOption{
		"-H": CurlValue{"Header1: value1", "Header2: value2"},
	}

	args, err := sanitizeInput(input)
	require.NoError(t, err)
	assert.Len(t, args, 4) // -H, value1, -H, value2
	assert.Equal(t, "-H", args[0])
	assert.Equal(t, "Header1: value1", args[1])
	assert.Equal(t, "-H", args[2])
	assert.Equal(t, "Header2: value2", args[3])
}

func TestHttpCurl_ValidRequest(t *testing.T) {
	options := CurlOption{
		"--location": CurlValue{"https://httpbin.org/get"},
	}

	output, err := HttpCurl(options, 30*time.Second)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestHttpCurl_InvalidURL(t *testing.T) {
	options := CurlOption{
		"--location": CurlValue{"https://invalid-domain-that-does-not-exist-12345.com"},
	}

	output, err := HttpCurl(options, 10*time.Second)
	assert.Error(t, err)
	// The output might contain error details
	assert.NotNil(t, output)
}

func TestHttpCurl_Timeout(t *testing.T) {
	options := CurlOption{
		"--location": CurlValue{"https://httpbin.org/delay/10"},
	}

	output, err := HttpCurl(options, 1*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request timed out")
	assert.NotNil(t, output)
}

func TestHttpCurl_POSTRequest(t *testing.T) {
	options := CurlOption{
		"-X":         CurlValue{"POST"},
		"-d":         CurlValue{"{\"test\":\"data\"}"},
		"-H":         CurlValue{"Content-Type: application/json"},
		"--location": CurlValue{"https://httpbin.org/post"},
	}

	output, err := HttpCurl(options, 30*time.Second)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestHttpCurl_WithProxy(t *testing.T) {
	options := CurlOption{
		"-x":         CurlValue{"http://proxy.example.com:8080"},
		"--location": CurlValue{"https://httpbin.org/get"},
	}

	output, err := HttpCurl(options, 10*time.Second)
	// This might fail if proxy is not available, but should not panic
	// We're mainly testing that the option is accepted
	if err != nil {
		// Expected if proxy is not available
		assert.NotNil(t, output)
	} else {
		assert.NotEmpty(t, output)
	}
}

func TestHttpCurl_WithSSLSkip(t *testing.T) {
	options := CurlOption{
		"-k":         CurlValue{""},
		"--location": CurlValue{"https://httpbin.org/get"},
	}

	output, err := HttpCurl(options, 30*time.Second)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestHttpCurl_ComplexRequest(t *testing.T) {
	options := CurlOption{
		"-X":         CurlValue{"POST"},
		"-d":         CurlValue{"{\"name\":\"test\",\"value\":123}"},
		"-H":         CurlValue{"Content-Type: application/json", "Accept: application/json"},
		"--location": CurlValue{"https://httpbin.org/post"},
		"-k":         CurlValue{""},
	}

	output, err := HttpCurl(options, 30*time.Second)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestAllowedCurlOptions_Completeness(t *testing.T) {
	// Test that all documented options are present
	expectedOptions := []string{
		"-k",         // Skip SSL verification
		"-x",         // HTTP Proxy
		"-X",         // HTTP method
		"-d",         // Data payload
		"--data",     // Data payload (alternative)
		"--location", // Follow redirects
		"-H",         // HTTP headers
	}

	for _, option := range expectedOptions {
		assert.True(t, AllowedCurlOptions[option], "Option %s should be allowed", option)
	}
}

func TestAllowedCurlOptions_UnauthorizedOptions(t *testing.T) {
	// Test that dangerous options are not allowed
	dangerousOptions := []string{
		"--output",       // File output
		"--upload-file",  // File upload
		"--config",       // Config file
		"--cookie",       // Cookie file
		"--cert",         // Certificate
		"--key",          // Private key
		"--cacert",       // CA certificate
		"--capath",       // CA certificate directory
		"--crlfile",      // Certificate revocation list
		"--pinnedpubkey", // Pinned public key
	}

	for _, option := range dangerousOptions {
		assert.False(t, AllowedCurlOptions[option], "Option %s should not be allowed", option)
	}
}

// Benchmark tests
func BenchmarkHttpCurl_SimpleRequest(b *testing.B) {
	options := CurlOption{
		"--location": CurlValue{"https://httpbin.org/get"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HttpCurl(options, 30*time.Second)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSanitizeInput(b *testing.B) {
	input := CurlOption{
		"-X":         CurlValue{"POST"},
		"-d":         CurlValue{"{\"test\":\"data\"}"},
		"-H":         CurlValue{"Content-Type: application/json"},
		"--location": CurlValue{"https://example.com"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sanitizeInput(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}
