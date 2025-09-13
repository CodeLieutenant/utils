# Logger Package

The logger package provides structured logging capabilities with enhanced features like stacktraces, runtime information, and flexible output formats.

## Features

- Structured logging with slog
- Configurable output formats (JSON, Text)
- Stacktrace support for error logs
- Runtime information injection (memory stats, goroutine count)
- Hostname and process ID tracking
- File and stdout output options
- Configurable log levels

## Installation

```bash
go get github.com/CodeLieutenant/utils/logger
```

## Basic Usage

### Simple Setup

```go
package main

import (
    "context"
    "log/slog"

    "github.com/CodeLieutenant/utils/logger"
)

func main() {
    config := &logger.LogConfig{
        Level:  "info",
        Format: "json",
        Output: "stdout",
    }

    slogger, err := logger.SetupLogger("myapp", "v1.0.0", config)
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    slogger.InfoContext(ctx, "Application started")
}
```

### Advanced Configuration

```go
package main

import (
    "context"
    "log/slog"

    "github.com/CodeLieutenant/utils/logger"
)

func main() {
    config := &logger.LogConfig{
        Level:          "debug",
        Format:         "json",
        Output:         "file",
        FilePath:       "/var/log/myapp.log",
        AddSource:      true,
        AddStacktrace:  true,
        AddRuntimeInfo: true,
    }

    slogger, err := logger.SetupLogger("myapp", "v1.0.0", config)
    if err != nil {
        panic(err)
    }

    ctx := context.Background()

    // Info log with structured data
    slogger.InfoContext(ctx, "User logged in",
        slog.String("user_id", "12345"),
        slog.String("ip", "192.168.1.1"),
    )

    // Error log with stacktrace
    slogger.ErrorContext(ctx, "Database connection failed",
        slog.String("error", "connection timeout"),
        slog.String("database", "postgres"),
    )

    // Debug log with runtime info
    slogger.DebugContext(ctx, "Processing request",
        slog.Int("request_id", 42),
    )
}
```

## Configuration Options

### LogConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Level` | `string` | Log level: debug, info, warn, error | `info` |
| `Format` | `string` | Output format: json, text | `json` |
| `Output` | `string` | Output destination: stdout, file | `stdout` |
| `FilePath` | `string` | File path when output is file | `""` |
| `AddSource` | `bool` | Add source file and line number | `false` |
| `AddStacktrace` | `bool` | Add stacktrace to error logs | `false` |
| `AddRuntimeInfo` | `bool` | Add runtime stats to logs | `false` |

### Log Levels

- `debug` - Detailed information for debugging
- `info` - General information about application flow
- `warn` - Warning messages for potentially harmful situations
- `error` - Error events that might still allow the application to continue

## Enhanced Features

### Stacktrace Support

When `AddStacktrace` is enabled, error logs automatically include stack traces:

```go
config := &logger.LogConfig{
    Level:         "error",
    Format:        "json",
    Output:        "stdout",
    AddStacktrace: true,
}

slogger, _ := logger.SetupLogger("myapp", "v1.0.0", config)
slogger.ErrorContext(ctx, "Critical error occurred") // Includes stacktrace
```

### Runtime Information

When `AddRuntimeInfo` is enabled, logs include memory and goroutine statistics:

```go
config := &logger.LogConfig{
    Level:          "info",
    Format:         "json",
    Output:         "stdout",
    AddRuntimeInfo: true,
}

slogger, _ := logger.SetupLogger("myapp", "v1.0.0", config)
slogger.InfoContext(ctx, "Status check") // Includes memory stats, goroutine count
```

### Metadata Injection

The logger automatically adds metadata to all logs:

- `hostname` - Server hostname
- `pid` - Process ID
- `version` - Application version
- `service` - Service name

## File Logging

For production applications, you can log to files:

```go
config := &logger.LogConfig{
    Level:    "info",
    Format:   "json",
    Output:   "file",
    FilePath: "/var/log/myapp/application.log",
}

slogger, err := logger.SetupLogger("myapp", "v1.0.0", config)
if err != nil {
    // Handle file creation/permission errors
    panic(err)
}
```

## Integration Examples

### HTTP Middleware

```go
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            next.ServeHTTP(w, r)

            logger.InfoContext(r.Context(), "HTTP request",
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
                slog.Duration("duration", time.Since(start)),
            )
        })
    }
}
```

### Error Handling

```go
func handleError(logger *slog.Logger, ctx context.Context, err error, msg string) {
    if err != nil {
        logger.ErrorContext(ctx, msg,
            slog.String("error", err.Error()),
        )
    }
}
```

## Performance Considerations

- Use structured logging with slog attributes instead of string formatting
- Consider log levels carefully in production (avoid debug logs)
- File logging includes automatic directory creation
- JSON format is recommended for production environments
- Stacktraces have performance overhead, use judiciously

## Thread Safety

The logger is thread-safe and can be used concurrently from multiple goroutines.

## Testing

For testing, you can use a no-op logger or capture logs:

```go
func TestWithLogger(t *testing.T) {
    config := &logger.LogConfig{
        Level:  "debug",
        Format: "text",
        Output: "stdout",
    }

    testLogger, err := logger.SetupLogger("test", "v0.0.1", config)
    require.NoError(t, err)

    // Use testLogger in your tests
}
```

## Back to Main Documentation

← [Back to main README](../README.md)
