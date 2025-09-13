# Signals Package

The signals package provides cross-platform OS signal handling utilities for Go applications, abstracting platform-specific signal differences between Unix-like systems and Windows.

## Features

- Cross-platform signal definitions
- String-to-signal conversion
- Safe signal retrieval with error handling
- Panic-based signal retrieval for must-have scenarios
- Support for common signals across Unix and Windows

## Installation

```bash
go get github.com/CodeLieutenant/utils/signals
```

## Supported Signals

### Unix/Linux/macOS
- `SIGHUP` - Hangup
- `SIGINT` - Interrupt (Ctrl+C)
- `SIGQUIT` - Quit (Ctrl+\)
- `SIGKILL` - Kill (cannot be caught)
- `SIGALRM` - Alarm clock
- `SIGTERM` - Termination
- `SIGUSR1` - User-defined signal 1
- `SIGUSR2` - User-defined signal 2

### Windows
- `SIGHUP` - Hangup
- `SIGINT` - Interrupt (Ctrl+C)
- `SIGQUIT` - Quit
- `SIGKILL` - Kill
- `SIGALRM` - Alarm clock
- `SIGTERM` - Termination

Note: `SIGUSR1` and `SIGUSR2` are not available on Windows.

## Basic Usage

### Safe Signal Retrieval

```go
package main

import (
    "fmt"
    "os"
    "os/signal"

    "github.com/CodeLieutenant/utils/signals"
)

func main() {
    // Safe signal retrieval with error handling
    sigterm, err := signals.Get("SIGTERM")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    sigint, err := signals.Get("SIGINT")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // Setup signal channel
    c := make(chan os.Signal, 1)
    signal.Notify(c, sigterm, sigint)

    fmt.Println("Waiting for signal...")
    sig := <-c
    fmt.Printf("Received signal: %v\n", sig)
}
```

### Must-Get Signal Retrieval

```go
package main

import (
    "fmt"
    "os"
    "os/signal"

    "github.com/CodeLieutenant/utils/signals"
)

func main() {
    // Panic-based retrieval for critical signals
    sigterm := signals.MustGet("SIGTERM")
    sigint := signals.MustGet("SIGINT")

    c := make(chan os.Signal, 1)
    signal.Notify(c, sigterm, sigint)

    fmt.Println("Server starting... Press Ctrl+C to stop")
    <-c
    fmt.Println("Shutting down gracefully...")
}
```

## Graceful Shutdown Example

### HTTP Server with Graceful Shutdown

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/CodeLieutenant/utils/signals"
)

func main() {
    // Create HTTP server
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // Setup signal handling
    sigterm := signals.MustGet("SIGTERM")
    sigint := signals.MustGet("SIGINT")

    c := make(chan os.Signal, 1)
    signal.Notify(c, sigterm, sigint)

    // Start server in goroutine
    go func() {
        fmt.Println("Server starting on :8080")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            fmt.Printf("Server error: %v\n", err)
        }
    }()

    // Wait for signal
    <-c
    fmt.Println("Received shutdown signal")

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        fmt.Printf("Server shutdown error: %v\n", err)
    } else {
        fmt.Println("Server shutdown complete")
    }
}
```

### Worker Pool Shutdown

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "sync"
    "time"

    "github.com/CodeLieutenant/utils/signals"
)

func worker(ctx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for {
        select {
        case job := <-jobs:
            fmt.Printf("Worker %d processing job %d\n", id, job)
            time.Sleep(time.Second) // Simulate work
        case <-ctx.Done():
            fmt.Printf("Worker %d shutting down\n", id)
            return
        }
    }
}

func main() {
    const numWorkers = 3
    jobs := make(chan int, 10)

    // Create context for cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start workers
    var wg sync.WaitGroup
    for i := 1; i <= numWorkers; i++ {
        wg.Add(1)
        go worker(ctx, i, jobs, &wg)
    }

    // Send some jobs
    go func() {
        for i := 1; i <= 5; i++ {
            jobs <- i
            time.Sleep(500 * time.Millisecond)
        }
        close(jobs)
    }()

    // Setup signal handling
    sigterm := signals.MustGet("SIGTERM")
    sigint := signals.MustGet("SIGINT")

    c := make(chan os.Signal, 1)
    signal.Notify(c, sigterm, sigint)

    // Wait for signal
    <-c
    fmt.Println("Received shutdown signal, stopping workers...")

    // Cancel context to stop workers
    cancel()

    // Wait for workers to finish
    wg.Wait()
    fmt.Println("All workers stopped")
}
```

## Advanced Usage

### Signal-based Configuration Reload

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "sync"

    "github.com/CodeLieutenant/utils/signals"
)

type Config struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Config) Reload() {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Simulate config reload
    c.data = map[string]string{
        "version": fmt.Sprintf("v1.0.%d", len(c.data)+1),
    }
    fmt.Printf("Config reloaded: %+v\n", c.data)
}

func (c *Config) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

func main() {
    config := &Config{
        data: map[string]string{"version": "v1.0.0"},
    }

    // Setup signal handling for config reload
    sighup := signals.MustGet("SIGHUP")   // Reload config
    sigterm := signals.MustGet("SIGTERM") // Shutdown
    sigint := signals.MustGet("SIGINT")   // Shutdown

    c := make(chan os.Signal, 1)
    signal.Notify(c, sighup, sigterm, sigint)

    fmt.Println("Application started. Send SIGHUP to reload config, SIGTERM/SIGINT to exit.")

    for {
        sig := <-c

        switch sig {
        case sighup:
            config.Reload()
        case sigterm, sigint:
            fmt.Println("Shutting down...")
            return
        }
    }
}
```

### Platform-Specific Signal Handling

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "runtime"

    "github.com/CodeLieutenant/utils/signals"
)

func main() {
    // Common signals available on all platforms
    sigterm := signals.MustGet("SIGTERM")
    sigint := signals.MustGet("SIGINT")

    c := make(chan os.Signal, 1)
    signal.Notify(c, sigterm, sigint)

    // Try to get Unix-specific signals
    if runtime.GOOS != "windows" {
        if sigusr1, err := signals.Get("SIGUSR1"); err == nil {
            signal.Notify(c, sigusr1)
            fmt.Println("SIGUSR1 handling enabled (Unix only)")
        }

        if sigusr2, err := signals.Get("SIGUSR2"); err == nil {
            signal.Notify(c, sigusr2)
            fmt.Println("SIGUSR2 handling enabled (Unix only)")
        }
    }

    fmt.Printf("Running on %s, waiting for signals...\n", runtime.GOOS)

    for {
        sig := <-c
        fmt.Printf("Received signal: %v\n", sig)

        // Handle platform-specific behavior
        if sig.String() == "user defined signal 1" ||
           sig.String() == "user defined signal 2" {
            fmt.Println("User signal received - performing custom action")
            continue
        }

        // Exit on termination signals
        fmt.Println("Exiting...")
        break
    }
}
```

## Error Handling

### Safe Signal Operations

```go
package main

import (
    "fmt"
    "os"
    "os/signal"

    "github.com/CodeLieutenant/utils/signals"
)

func setupSignalHandler(signalNames []string) {
    var validSignals []os.Signal

    for _, name := range signalNames {
        if sig, err := signals.Get(name); err != nil {
            fmt.Printf("Warning: Signal %s not available: %v\n", name, err)
        } else {
            validSignals = append(validSignals, sig)
            fmt.Printf("Registered signal: %s\n", name)
        }
    }

    if len(validSignals) == 0 {
        fmt.Println("No valid signals to handle")
        return
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, validSignals...)

    go func() {
        sig := <-c
        fmt.Printf("Received signal: %v\n", sig)
        os.Exit(0)
    }()
}

func main() {
    // Try to register both Unix and Windows signals
    signalNames := []string{
        "SIGTERM", "SIGINT", "SIGHUP",
        "SIGUSR1", "SIGUSR2", // These will fail on Windows
    }

    setupSignalHandler(signalNames)

    fmt.Println("Application running... Send a signal to exit")
    select {} // Block forever
}
```

## Testing

### Testing Signal Handlers

```go
package main

import (
    "os"
    "testing"
    "time"

    "github.com/CodeLieutenant/utils/signals"
)

func TestSignalHandling(t *testing.T) {
    // Test signal retrieval
    sigint, err := signals.Get("SIGINT")
    if err != nil {
        t.Fatalf("Failed to get SIGINT: %v", err)
    }

    // Test invalid signal
    _, err = signals.Get("INVALID_SIGNAL")
    if err == nil {
        t.Fatal("Expected error for invalid signal")
    }

    // Test MustGet with valid signal
    sigterm := signals.MustGet("SIGTERM")
    if sigterm == nil {
        t.Fatal("SIGTERM should not be nil")
    }
}

func TestGracefulShutdown(t *testing.T) {
    done := make(chan bool)

    // Simulate a service that handles shutdown
    go func() {
        sigterm := signals.MustGet("SIGTERM")
        c := make(chan os.Signal, 1)
        signal.Notify(c, sigterm)

        <-c
        // Cleanup...
        done <- true
    }()

    // Send signal after a delay
    go func() {
        time.Sleep(100 * time.Millisecond)
        p, _ := os.FindProcess(os.Getpid())
        p.Signal(signals.MustGet("SIGTERM"))
    }()

    // Wait for shutdown
    select {
    case <-done:
        // Success
    case <-time.After(1 * time.Second):
        t.Fatal("Shutdown timeout")
    }
}
```

## Best Practices

1. **Always handle SIGTERM and SIGINT** for graceful shutdown
2. **Use context.Context** for coordinated cancellation
3. **Set timeouts** for shutdown operations
4. **Test signal handling** in your applications
5. **Handle platform differences** when using Unix-specific signals
6. **Use buffered channels** to avoid missing signals
7. **Clean up resources** during shutdown

## Common Patterns

### Service Manager Pattern

```go
type Service interface {
    Start() error
    Stop() error
}

type ServiceManager struct {
    services []Service
}

func (sm *ServiceManager) Start() error {
    for _, service := range sm.services {
        if err := service.Start(); err != nil {
            return err
        }
    }
    return nil
}

func (sm *ServiceManager) Stop() error {
    for i := len(sm.services) - 1; i >= 0; i-- {
        if err := sm.services[i].Stop(); err != nil {
            return err
        }
    }
    return nil
}

func (sm *ServiceManager) Run() error {
    if err := sm.Start(); err != nil {
        return err
    }

    sigterm := signals.MustGet("SIGTERM")
    sigint := signals.MustGet("SIGINT")

    c := make(chan os.Signal, 1)
    signal.Notify(c, sigterm, sigint)

    <-c
    return sm.Stop()
}
```

## Back to Main Documentation

← [Back to main README](../README.md)
