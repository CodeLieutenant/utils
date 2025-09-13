# URL Signer Package

The urlsigner package provides HMAC-based URL signing capabilities for creating secure, time-limited URLs. This is useful for protecting downloads, API endpoints, or any URLs that should only be accessible for a limited time or by authorized parties.

## Features

- HMAC-based URL signing with multiple hash algorithms
- Time-based URL expiration
- Signature verification
- Support for SHA256, SHA512, SHA3, and BLAKE2B algorithms
- Base64 URL-safe encoding
- Custom time providers for testing

## Installation

```bash
go get github.com/CodeLieutenant/utils/urlsigner
```

## Quick Start

### Basic URL Signing

```go
package main

import (
    "fmt"
    "net/url"
    "time"

    "github.com/CodeLieutenant/utils"
    "github.com/CodeLieutenant/utils/urlsigner"
)

func main() {
    // Create signing key (32-64 bytes)
    key, _ := utils.ParseKey("deadbeefcafebabe1234567890abcdef12345678")

    // Create HMAC signer with SHA256
    signer := urlsigner.New("sha256", key)

    // Sign a URL with 1 hour expiration
    originalURL := "https://example.com/api/download?file=document.pdf"
    signedURL, err := signer.Sign(originalURL, time.Hour)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Signed URL: %s\n", signedURL)

    // Verify the signed URL
    parsedURL, _ := url.Parse(signedURL)
    if err := signer.Verify(parsedURL); err != nil {
        fmt.Printf("Verification failed: %v\n", err)
    } else {
        fmt.Println("URL signature is valid!")
    }
}
```

## Supported Hash Algorithms

The package supports multiple hash algorithms through the utils package:

- `sha256` - SHA-256 (recommended)
- `sha512/256` - SHA-512/256
- `sha3-256` - SHA3-256
- `sha3-512` - SHA3-512
- `blake2b-256` - BLAKE2B-256
- `blake2b-512` - BLAKE2B-512

```go
// Different hash algorithms
sha256Signer := urlsigner.New("sha256", key)
sha3Signer := urlsigner.New("sha3-256", key)
blake2bSigner := urlsigner.New("blake2b-256", key)
```

## Key Management

### Key Formats

Keys can be provided in hex or base64 format:

```go
// Hex format
hexKey := "deadbeefcafebabe1234567890abcdef12345678"
key1, _ := utils.ParseKey(hexKey)

// Base64 format
base64Key := "base64:3q2+78r+uro="
key2, _ := utils.ParseKey(base64Key)

signer := urlsigner.New("sha256", key1)
```

### Key Requirements

- Keys must be between 32 and 64 bytes
- Use cryptographically secure random keys in production
- Rotate keys regularly for security

```go
// Generate a secure key
import "crypto/rand"

func generateKey() []byte {
    key := make([]byte, 32) // 256-bit key
    rand.Read(key)
    return key
}
```

## Time-Based Expiration

### Expiration Durations

```go
signer := urlsigner.New("sha256", key)

// 1 hour expiration
shortURL, _ := signer.Sign("https://example.com/temp", time.Hour)

// 24 hours expiration
dayURL, _ := signer.Sign("https://example.com/daily", 24*time.Hour)

// 30 days expiration
monthURL, _ := signer.Sign("https://example.com/monthly", 30*24*time.Hour)

// No expiration (permanent signature)
permanentURL, _ := signer.Sign("https://example.com/permanent", 0)
```

### Custom Time Provider

For testing or custom time logic:

```go
// Custom time provider for testing
mockTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
signer := urlsigner.New("sha256", key, func() time.Time {
    return mockTime
})

// This URL will be considered expired immediately
expiredURL, _ := signer.Sign("https://example.com/test", -time.Hour)
```

## URL Verification

### Verification Process

```go
func verifyURL(signer *urlsigner.HMACSigner, urlString string) error {
    parsedURL, err := url.Parse(urlString)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    if err := signer.Verify(parsedURL); err != nil {
        switch err {
        case urlsigner.ErrMissingSignature:
            return fmt.Errorf("URL not signed")
        case urlsigner.ErrInvalidSignature:
            return fmt.Errorf("invalid signature")
        case urlsigner.ErrExpired:
            return fmt.Errorf("URL expired")
        default:
            return fmt.Errorf("verification failed: %w", err)
        }
    }

    return nil
}
```

### Error Types

The package provides specific error types:

```go
var (
    ErrMissingSignature = errors.New("missing signature")
    ErrInvalidSignature = errors.New("invalid signature")
    ErrExpired          = errors.New("url expired")
)
```

## HTTP Integration

### File Download Protection

```go
package main

import (
    "net/http"
    "net/url"
    "time"

    "github.com/CodeLieutenant/utils/urlsigner"
)

func protectedDownloadHandler(signer *urlsigner.HMACSigner) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Verify the URL signature
        if err := signer.Verify(r.URL); err != nil {
            http.Error(w, "Access denied: "+err.Error(), http.StatusForbidden)
            return
        }

        // Extract original file parameter
        filename := r.URL.Query().Get("file")
        if filename == "" {
            http.Error(w, "Missing file parameter", http.StatusBadRequest)
            return
        }

        // Serve the file
        http.ServeFile(w, r, "/secure/files/"+filename)
    }
}

func generateDownloadLink(signer *urlsigner.HMACSigner, filename string) (string, error) {
    baseURL := "https://example.com/download?file=" + url.QueryEscape(filename)
    return signer.Sign(baseURL, 15*time.Minute) // 15-minute download window
}

func main() {
    key := []byte("your-secret-key-here-32-bytes-min")
    signer := urlsigner.New("sha256", key)

    http.HandleFunc("/download", protectedDownloadHandler(signer))
    http.HandleFunc("/generate-link", func(w http.ResponseWriter, r *http.Request) {
        filename := r.URL.Query().Get("file")
        signedURL, err := generateDownloadLink(signer, filename)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"download_url": "%s"}`, signedURL)
    })

    http.ListenAndServe(":8080", nil)
}
```

### API Endpoint Protection

```go
func protectedAPIHandler(signer *urlsigner.HMACSigner) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Verify signature
        if err := signer.Verify(r.URL); err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusForbidden)
            fmt.Fprintf(w, `{"error": "access denied", "reason": "%s"}`, err.Error())
            return
        }

        // Process API request
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"status": "success", "data": "sensitive information"}`)
    }
}

func createAPILink(signer *urlsigner.HMACSigner, endpoint string, params map[string]string) (string, error) {
    u, err := url.Parse(endpoint)
    if err != nil {
        return "", err
    }

    // Add parameters
    query := u.Query()
    for key, value := range params {
        query.Set(key, value)
    }
    u.RawQuery = query.Encode()

    // Sign with 5-minute expiration
    return signer.Sign(u.String(), 5*time.Minute)
}
```

## Middleware Integration

### Chi Router Middleware

```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/CodeLieutenant/utils/urlsigner"
)

func URLSignatureMiddleware(signer *urlsigner.HMACSigner) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if err := signer.Verify(r.URL); err != nil {
                http.Error(w, "Signature verification failed: "+err.Error(), http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func main() {
    key := []byte("your-secret-key-here-32-bytes-min")
    signer := urlsigner.New("sha256", key)

    r := chi.NewRouter()

    // Protected routes
    r.Route("/api/protected", func(r chi.Router) {
        r.Use(URLSignatureMiddleware(signer))
        r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("Protected data"))
        })
    })

    // Public routes
    r.Get("/api/public", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Public data"))
    })

    http.ListenAndServe(":8080", r)
}
```

## Advanced Usage

### Multi-Key Support

```go
type MultiKeySigner struct {
    signers map[string]*urlsigner.HMACSigner
}

func NewMultiKeySigner() *MultiKeySigner {
    return &MultiKeySigner{
        signers: make(map[string]*urlsigner.HMACSigner),
    }
}

func (m *MultiKeySigner) AddKey(keyID string, key []byte, algo string) {
    m.signers[keyID] = urlsigner.New(algo, key)
}

func (m *MultiKeySigner) Sign(keyID, urlString string, duration time.Duration) (string, error) {
    signer, exists := m.signers[keyID]
    if !exists {
        return "", fmt.Errorf("key not found: %s", keyID)
    }

    // Add key ID to URL
    u, err := url.Parse(urlString)
    if err != nil {
        return "", err
    }

    query := u.Query()
    query.Set("key_id", keyID)
    u.RawQuery = query.Encode()

    return signer.Sign(u.String(), duration)
}

func (m *MultiKeySigner) Verify(u *url.URL) error {
    keyID := u.Query().Get("key_id")
    if keyID == "" {
        return fmt.Errorf("missing key ID")
    }

    signer, exists := m.signers[keyID]
    if !exists {
        return fmt.Errorf("unknown key ID: %s", keyID)
    }

    return signer.Verify(u)
}
```

### URL Builder Pattern

```go
type URLBuilder struct {
    signer   *urlsigner.HMACSigner
    baseURL  string
    params   map[string]string
    duration time.Duration
}

func NewURLBuilder(signer *urlsigner.HMACSigner, baseURL string) *URLBuilder {
    return &URLBuilder{
        signer:  signer,
        baseURL: baseURL,
        params:  make(map[string]string),
    }
}

func (b *URLBuilder) Param(key, value string) *URLBuilder {
    b.params[key] = value
    return b
}

func (b *URLBuilder) Expires(duration time.Duration) *URLBuilder {
    b.duration = duration
    return b
}

func (b *URLBuilder) Build() (string, error) {
    u, err := url.Parse(b.baseURL)
    if err != nil {
        return "", err
    }

    query := u.Query()
    for key, value := range b.params {
        query.Set(key, value)
    }
    u.RawQuery = query.Encode()

    return b.signer.Sign(u.String(), b.duration)
}

// Usage
func example() {
    key := []byte("your-secret-key-here-32-bytes-min")
    signer := urlsigner.New("sha256", key)

    signedURL, err := NewURLBuilder(signer, "https://api.example.com/data").
        Param("user_id", "12345").
        Param("format", "json").
        Expires(time.Hour).
        Build()

    if err != nil {
        panic(err)
    }

    fmt.Printf("Signed URL: %s\n", signedURL)
}
```

## Testing

### Unit Tests

```go
package main

import (
    "net/url"
    "testing"
    "time"

    "github.com/CodeLieutenant/utils/urlsigner"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestURLSigning(t *testing.T) {
    key := []byte("test-key-32-bytes-long-for-security")
    signer := urlsigner.New("sha256", key)

    originalURL := "https://example.com/test?param=value"

    // Test signing
    signedURL, err := signer.Sign(originalURL, time.Hour)
    require.NoError(t, err)
    assert.Contains(t, signedURL, "signature=")
    assert.Contains(t, signedURL, "expires=")

    // Test verification
    parsedURL, err := url.Parse(signedURL)
    require.NoError(t, err)

    err = signer.Verify(parsedURL)
    assert.NoError(t, err)
}

func TestExpiredURL(t *testing.T) {
    // Use custom time provider
    mockTime := time.Now()
    signer := urlsigner.New("sha256", []byte("test-key-32-bytes-long-for-security"),
        func() time.Time { return mockTime })

    // Sign URL with past expiration
    signedURL, err := signer.Sign("https://example.com/test", -time.Hour)
    require.NoError(t, err)

    // Verification should fail
    parsedURL, _ := url.Parse(signedURL)
    err = signer.Verify(parsedURL)
    assert.Equal(t, urlsigner.ErrExpired, err)
}

func TestInvalidSignature(t *testing.T) {
    key := []byte("test-key-32-bytes-long-for-security")
    signer := urlsigner.New("sha256", key)

    // Create a URL with invalid signature
    invalidURL := "https://example.com/test?signature=invalid"
    parsedURL, _ := url.Parse(invalidURL)

    err := signer.Verify(parsedURL)
    assert.Equal(t, urlsigner.ErrInvalidSignature, err)
}
```

## Security Considerations

1. **Key Management**: Use cryptographically secure keys of adequate length (32+ bytes)
2. **Key Rotation**: Regularly rotate signing keys
3. **Time Synchronization**: Ensure server clocks are synchronized when using expiration
4. **HTTPS Only**: Always use HTTPS to prevent signature interception
5. **Short Expiration**: Use the shortest practical expiration times
6. **Rate Limiting**: Implement rate limiting on signature generation endpoints

## Best Practices

1. **Use SHA-256 or stronger** hash algorithms
2. **Set reasonable expiration times** based on use case
3. **Validate URLs** before signing
4. **Log signature failures** for security monitoring
5. **Test expiration logic** thoroughly
6. **Use middleware** for consistent verification
7. **Handle errors gracefully** in user-facing applications

## Back to Main Documentation

← [Back to main README](../README.md)
