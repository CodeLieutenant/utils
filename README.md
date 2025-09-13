# Utils

A comprehensive Go utility library providing essential tools for common development tasks.

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/doc/devel/release.html)
[![Go Reference](https://pkg.go.dev/badge/github.com/CodeLieutenant/utils.svg)](https://pkg.go.dev/github.com/CodeLieutenant/utils)
[![Go Report Card](https://goreportcard.com/badge/github.com/CodeLieutenant/utils)](https://goreportcard.com/report/github.com/CodeLieutenant/utils)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![codecov](https://codecov.io/gh/CodeLieutenant/utils/graph/badge.svg?token=lqhmgWPlWJ)](https://codecov.io/gh/CodeLieutenant/utils)

## Installation

```bash
go get github.com/CodeLieutenant/utils
```

## Modules

This package contains several utility modules:

- **Core utilities** - File operations, path handling, password generation, memory formatting
- **Environment management** - Environment variable handling with type conversion
- **Network utilities** - IP address detection and validation
- **Cryptographic utilities** - Key parsing and hashing algorithms
- **[HTTP utilities](httputils/README.md)** - Chi router helpers and HTTP response utilities
- **[Logging](logger/README.md)** - Structured logging with stacktraces and runtime info
- **[URL signing](urlsigner/README.md)** - HMAC-based URL signing for secure links
- **[Signal handling](signals/README.md)** - Cross-platform OS signal management

## 📚 Detailed Documentation

Each subpackage has comprehensive documentation with examples:

- **[Logger Package](logger/README.md)** - Structured logging with slog, stacktraces, and runtime information
- **[HTTP Utils Package](httputils/README.md)** - Chi router setup, middleware, and HTTP helpers
- **[Signals Package](signals/README.md)** - Cross-platform OS signal handling utilities
- **[URL Signer Package](urlsigner/README.md)** - HMAC-based URL signing for secure time-limited links

## Core Utilities

### File Operations

```go
package main

import (
    "fmt"
    "os"

    "github.com/CodeLieutenant/utils"
)

func main() {
    // Create a directory with permissions
    dirPath, err := utils.CreateDirectory("/tmp/myapp", 0755)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Directory created: %s\n", dirPath)

    // Create a log file
    logFile, err := utils.CreateLogFile("/tmp/myapp/app.log")
    if err != nil {
        panic(err)
    }
    defer logFile.Close()

    // Check if file exists
    if utils.FileExists("/tmp/myapp/app.log") {
        fmt.Println("Log file exists!")
    }
}
```

### Password Generation

```go
package main

import (
    "fmt"

    "github.com/CodeLieutenant/utils"
)

func main() {
    // Generate a secure random password
    password, err := utils.GenerateRandomPassword(16)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Generated password: %s\n", password)
}
```

### Memory Size Formatting

```go
package main

import (
    "fmt"

    "github.com/CodeLieutenant/utils"
)

func main() {
    // Format memory sizes
    size := utils.MemorySize(1024 * 1024 * 1024) // 1 GiB
    fmt.Printf("Memory size: %s\n", size.String()) // Output: 1GiB

    // Different sizes
    fmt.Printf("1024 bytes: %s\n", utils.MemorySize(1024).String()) // 1KiB
    fmt.Printf("1MB: %s\n", utils.MemorySize(1024*1024).String())   // 1MiB
}
```

### Unsafe String/Bytes Conversion

```go
package main

import (
    "fmt"

    "github.com/CodeLieutenant/utils"
)

func main() {
    // Zero-copy string to bytes conversion
    str := "hello world"
    bytes := utils.UnsafeBytes(str)
    fmt.Printf("String as bytes: %v\n", bytes)

    // Zero-copy bytes to string conversion
    backToString := utils.UnsafeString(bytes)
    fmt.Printf("Bytes as string: %s\n", backToString)
}
```

## Environment Management

```go
package main

import (
    "fmt"

    "github.com/CodeLieutenant/utils"
)

func main() {
    // Initialize environment (loads .env file automatically)
    env := utils.NewEnv(false)

    // Get string with default
    dbHost := utils.GetStringEnv(env, "DB_HOST", "localhost")
    fmt.Printf("Database host: %s\n", dbHost)

    // Get integer with default
    dbPort := utils.GetIntEnv(env, "DB_PORT", 5432)
    fmt.Printf("Database port: %d\n", dbPort)

    // Get boolean with default
    debug := utils.GetBoolEnv(env, "DEBUG", false)
    fmt.Printf("Debug mode: %t\n", debug)

    // Get duration with default
    timeout := utils.GetDurationEnv(env, "REQUEST_TIMEOUT", "30s")
    fmt.Printf("Request timeout: %v\n", timeout)

    // Get string slice
    allowedHosts := utils.GetStringsEnv(env, "ALLOWED_HOSTS", []string{"localhost"})
    fmt.Printf("Allowed hosts: %v\n", allowedHosts)
}
```

## Network Utilities

```go
package main

import (
    "fmt"

    "github.com/CodeLieutenant/utils"
)

func main() {
    // Get local machine IP
    localIP := utils.GetLocalIP()
    fmt.Printf("Local IP: %s\n", localIP)

    // Get all local IPs
    localIPs := utils.GetLocalIPs()
    fmt.Printf("All local IPs: %v\n", localIPs)
}
```

## Cryptographic Utilities

```go
package main

import (
    "fmt"

    "github.com/CodeLieutenant/utils"
)

func main() {
    // Parse hex-encoded key
    hexKey := "deadbeefcafebabe1234567890abcdef12345678"
    key1, err := utils.ParseKey(hexKey)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Hex key length: %d bytes\n", len(key1))

    // Parse base64-encoded key
    base64Key := "base64:3q2+78r+uro="
    key2, err := utils.ParseKey(base64Key)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Base64 key length: %d bytes\n", len(key2))

    // Get hash function
    hasher := utils.ParseHasher("sha256")
    if hasher != nil {
        h := hasher()
        h.Write([]byte("hello world"))
        hash := h.Sum(nil)
        fmt.Printf("SHA256 hash length: %d bytes\n", len(hash))
    }
}
```

## HTTP Utilities

> 📖 **For comprehensive HTTP utilities documentation, see: [HTTP Utils Package Documentation](httputils/README.md)**

```go
package main

import (
    "net/http"

    "github.com/CodeLieutenant/utils/httputils"
    "github.com/go-chi/chi/v5"
)

func main() {
    // Setup router with production middleware
    config := httputils.ProductionMiddlewareConfig()
    options := &httputils.RouterSetupOptions{
        LoggerColor: false,
        Middleware:  config,
    }

    r := httputils.SetupRouter(options)

    // Add routes
    r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
        response := map[string]string{"status": "ok"}
        httputils.NewResponse(w).OK(response)
    })

    http.ListenAndServe(":8080", r)
}
```

## Logging

> 📖 **For comprehensive logging documentation, see: [Logger Package Documentation](logger/README.md)**

```go
package main

import (
    "context"
    "log/slog"

    "github.com/CodeLieutenant/utils/logger"
)

func main() {
    // Configure structured logging
    config := &logger.LogConfig{
        Level:          "info",
        Format:         "json",
        Output:         "stdout",
        AddSource:      true,
        AddStacktrace:  true,
        AddRuntimeInfo: true,
    }

    // Setup logger
    slogger, err := logger.SetupLogger("myapp", "v1.0.0", config)
    if err != nil {
        panic(err)
    }

    // Use the logger
    ctx := context.Background()
    slogger.InfoContext(ctx, "Application started",
        slog.String("version", "v1.0.0"),
        slog.Int("port", 8080),
    )

    slogger.ErrorContext(ctx, "Something went wrong",
        slog.String("error", "database connection failed"),
    )
}
```

## URL Signing

> 📖 **For comprehensive URL signing documentation, see: [URL Signer Package Documentation](urlsigner/README.md)**

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
    // Create HMAC signer with SHA256
    key, _ := utils.ParseKey("deadbeefcafebabe1234567890abcdef12345678")
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

## Signal Handling

> 📖 **For comprehensive signal handling documentation, see: [Signals Package Documentation](signals/README.md)**

```go
package main

import (
    "fmt"
    "os"

    "github.com/CodeLieutenant/utils/signals"
)

func main() {
    // Get signal by name
    sigterm, err := signals.Get("SIGTERM")
    if err != nil {
        panic(err)
    }
    fmt.Printf("SIGTERM signal: %v\n", sigterm)

    // Must get signal (panics if not found)
    sigint := signals.MustGet("SIGINT")
    fmt.Printf("SIGINT signal: %v\n", sigint)

    // Use in signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, sigint, sigterm)

    fmt.Println("Waiting for signal...")
    sig := <-sigChan
    fmt.Printf("Received signal: %v\n", sig)
}
```

## Environment Constants

The package provides predefined environment constants:

```go
const (
    EnvDev   = "development"
    EnvProd  = "production"
    EnvStage = "staging"
)
```

## Testing

Run the test suite:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# View coverage report
make coverage-html
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

- [github.com/go-chi/chi/v5](https://github.com/go-chi/chi) - HTTP router
- [github.com/joho/godotenv](https://github.com/joho/godotenv) - Environment file loading
- [github.com/stretchr/testify](https://github.com/stretchr/testify) - Testing toolkit
- [golang.org/x/crypto](https://golang.org/x/crypto) - Extended cryptography support
