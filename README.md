# HTTP-CURL
**HTTP cURL** is a web service built on Golang Echo framework that allows you to execute cURL commands via HTTP requests. It provides a secure way to make HTTP requests through a REST API interface.

## Features
- Execute cURL commands via HTTP POST requests
- Configurable timeout support
- Base64 encoding option for responses
- Plain text output support
- Built-in waiting endpoint for testing
- Docker containerization support
- Security restrictions on allowed cURL options
- Comprehensive test suite with unit, integration, and benchmark tests
- Configurable debug output

## Pre-requisites
- Go 1.21 or higher
- cURL installed in your system

## Quick Start

### Local Development
```bash
# Clone the repository
git clone <repository-url>
cd http-curl

# Run the service
go run main.go

# Or specify a custom port
PORT=8081 go run main.go

# Hide curl arguments in production
HIDE_CURL_OPTIONS=true go run main.go
```

### Docker Deployment
```bash
# Build the Docker image
docker build -t http-curl .

# Run the container
docker run -p 8080:8080 http-curl

# Run with custom environment variables
docker run -p 8080:8081-e PORT=8081 -e HIDE_CURL_OPTIONS=true http-curl
```

## API Endpoints

### POST /curl
Execute cURL commands via HTTP POST request.

**Headers:**
- `Content-Type: application/json` (required)

**Query Parameters:**
- `timeout` (optional): Request timeout duration (e.g., "30s", "2m", "1h")
- `base64` (optional): Set to "true" to return response in base64 encoding
- `plain` (optional): Set to "true" to return raw response without JSON wrapper

**Request Body:**
```json
{
  "-X": "POST",
  "-d": "{\"foo\":\"bar\"}",
  "--location": "https://api.example.com/endpoint",
  "-H": "Content-Type: application/json"
}
```

**Response (Success):**
```json
{
  "result": "<output of the curl command>"
}
```

**Response (Error):**
```json
{
  "error": "Error executing curl command: <error details>",
  "details": "<stderr output>"
}
```

### ANY /waiting/:milli
Utility endpoint for testing timeouts and delays.

**Parameters:**
- `milli`: Number of milliseconds to wait before responding

**Response:**
```
Ok
```

## Allowed cURL Options
For security reasons, only the following cURL options are allowed:
- `-k`: Skip SSL verification
- `-x`: HTTP Proxy
- `-X`: HTTP method
- `-d`, `--data`: Data payload
- `--location`: Follow redirects
- `-H`: HTTP headers

## Examples

### Basic GET Request
```bash
curl -X POST http://localhost:8080/curl \
  -H "Content-Type: application/json" \
  -d '{
    "--location": "https://httpbin.org/get"
  }'
```

### POST Request with Data
```bash
curl -X POST http://localhost:8080/curl \
  -H "Content-Type: application/json" \
  -d '{
    "-X": "POST",
    "-d": "{\"name\":\"John\",\"age\":30}",
    "-H": "Content-Type: application/json",
    "--location": "https://httpbin.org/post"
  }'
```

### Request with Timeout
```bash
curl -X POST "http://localhost:8080/curl?timeout=30s" \
  -H "Content-Type: application/json" \
  -d '{
    "--location": "https://httpbin.org/delay/10"
  }'
```

### Base64 Encoded Response
```bash
curl -X POST "http://localhost:8080/curl?base64=true" \
  -H "Content-Type: application/json" \
  -d '{
    "--location": "https://httpbin.org/json"
  }'
```

### Plain Text Response
```bash
curl -X POST "http://localhost:8080/curl?plain=true" \
  -H "Content-Type: application/json" \
  -H "Accept: text/plain" \
  -d '{
    "--location": "https://httpbin.org/plain"
  }'
```

## Building

### Binary Build
```bash
# Build the binary
go build -o http-curl

# Run the binary
./http-curl
```

### Docker Build
```bash
# Build Docker image
docker build -t http-curl .

# Run with custom port
docker run -p 8080:8080 http-curl
```

## Using as a Library

The `lib/curl.go` package can be imported and used in other Go projects to execute cURL commands programmatically.

### Installation

Add the dependency to your `go.mod`:

```bash
go get github.com/roketid/http-curl/lib
```

Or add it manually to your `go.mod`:

```go
require (
    github.com/roketid/http-curl/lib v0.0.0
)
```

### Usage Example

```go
package main

import (
    "fmt"
    "time"
    
    httpcurl "github.com/roketid/http-curl/lib"
)

func main() {
    // Configure curl options
    options := httpcurl.CurlOption{
        "-X":         httpcurl.CurlValue{"GET"},
        "--location": httpcurl.CurlValue{"https://httpbin.org/get"},
        "-H":         httpcurl.CurlValue{"Content-Type: application/json"},
    }

    // Execute the curl command with timeout
    output, err := httpcurl.HttpCurl(options, 30*time.Second)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Response: %s\n", string(output))
}
```

### Advanced Usage

```go
package main

import (
    "fmt"
    "time"
    
    httpcurl "github.com/roketid/http-curl/lib"
)

func main() {
    // POST request with JSON data
    options := httpcurl.CurlOption{
        "-X":         httpcurl.CurlValue{"POST"},
        "-d":         httpcurl.CurlValue{`{"name":"John","age":30}`},
        "-H":         httpcurl.CurlValue{"Content-Type: application/json"},
        "--location": httpcurl.CurlValue{"https://httpbin.org/post"},
        "-k":         httpcurl.CurlValue{""}, // Skip SSL verification
    }

    // Set timeout and disable debug output
    httpcurl.SetPrintArgs(false)
    output, err := httpcurl.HttpCurl(options, 10*time.Second)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Response: %s\n", string(output))
}
```

### Available Functions

- `HttpCurl(options CurlOption, timeout time.Duration) ([]byte, error)`: Execute curl command
- `SetPrintArgs(print bool)`: Control debug output
- `AllowedCurlOptions`: Map of allowed curl options

### Supported cURL Options

The library supports the same restricted set of cURL options as the web service:
- `-k`: Skip SSL verification
- `-x`: HTTP Proxy
- `-X`: HTTP method
- `-d`, `--data`: Data payload
- `--location`: Follow redirects
- `-H`: HTTP headers

### Error Handling

```go
output, err := httpcurl.HttpCurl(options, 5*time.Second)
if err != nil {
    switch {
    case err.Error() == "request timed out":
        fmt.Println("Request timed out")
    case err.Error() == "unauthorized curl option":
        fmt.Println("Invalid curl option provided")
    default:
        fmt.Printf("Curl error: %v\n", err)
    }
    return
}
```

## Deployment

### Docker Deployment
The project includes a Dockerfile for containerized deployment:

```bash
# Build and push to registry
docker build -t your-registry/http-curl .
docker push your-registry/http-curl
```

### GitHub Actions (CI/CD)
This repository includes a GitHub Actions workflow that automatically builds and pushes Docker images to the GitHub Container Registry (GHCR).

#### Features:
- **Automatic builds**: Triggers on pushes to main/master branch and on tags
- **Pull request builds**: Builds on PRs for testing (without pushing)
- **Multi-platform support**: Uses Docker Buildx for efficient builds
- **Caching**: Uses GitHub Actions cache for faster builds
- **Smart tagging**: Automatically tags images based on git refs and semantic versions

#### Usage:
The workflow automatically runs when you:
1. Push to `main` or `master` branch
2. Create a tag (e.g., `v1.0.0`)
3. Create a pull request

#### Accessing the Docker Image:
```bash
# Pull the latest image
docker pull ghcr.io/roketid/http-curl:latest

# Pull a specific version
docker pull ghcr.io/roketid/http-curl:v1.0.0

# Run the container
docker run -p 8080:8080 ghcr.io/roketid/http-curl:latest
```

#### Available Tags:
- `latest`: Latest commit on main branch
- `v*`: Semantic version tags (e.g., `v1.0.0`, `v1.0`)
- `main-*`: Branch-specific tags with commit SHA
- `pr-*`: Pull request tags

### Cloud Platforms
This service can be deployed on:
- AWS (ECS, EKS, Lambda)
- Google Cloud Platform (GKE, Cloud Run)
- Azure (AKS, Container Instances)
- Any Kubernetes cluster
- Traditional VPS or dedicated servers

### Environment Variables
- `PORT`: Server port (default: 8080)
- `HIDE_CURL_OPTIONS`: Set to "true" to disable curl argument logging (useful for production)

## Testing

### Running Tests
```bash
# Run all tests
go test -v

# Run with coverage
go test -v -cover

# Run specific test categories
go test -v -run "^TestHandleCurl" ./...     # Unit tests
go test -v -run "^TestIntegration" ./...    # Integration tests
go test -v ./lib                            # Library tests

# Run benchmark tests
go test -v -bench=. -benchmem ./...
```

### Test Categories
- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test complete HTTP endpoints and workflows
- **Library Tests**: Test the core curl library functionality
- **Benchmark Tests**: Performance testing for critical paths

### Test Coverage
The test suite covers:
- ✅ HTTP endpoint functionality
- ✅ Request validation and error handling
- ✅ Timeout mechanisms
- ✅ Security restrictions on curl options
- ✅ Base64 and plain text response formats
- ✅ JSON parsing and validation
- ✅ Concurrent request handling
- ✅ Edge cases and error scenarios

## Security Considerations
- Only whitelisted cURL options are allowed
- Request timeouts prevent long-running commands
- Input validation and sanitization
- No direct command injection vulnerabilities
- Configurable debug output for production environments

## Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Write tests for new features
- Ensure all tests pass before submitting PR
- Follow Go coding conventions
- Update documentation for API changes

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Support
For issues and questions:
1. Check existing issues
2. Create a new issue with detailed description
3. Include steps to reproduce
4. Provide expected vs actual behavior
