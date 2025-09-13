# HTTP Utils Package

The httputils package provides HTTP utilities for building web applications with the Chi router, including middleware configuration, request/response helpers, and JSON handling.

## Features

- Chi router setup with configurable middleware
- Type-safe JSON request/response handling
- Fluent response API
- Production-ready middleware configuration
- Request body parsing with validation
- Error response formatting

## Installation

```bash
go get github.com/CodeLieutenant/utils/httputils
```

## Quick Start

### Basic Router Setup

```go
package main

import (
    "net/http"
    
    "github.com/CodeLieutenant/utils/httputils"
    "github.com/go-chi/chi/v5"
)

func main() {
    // Production middleware configuration
    config := httputils.ProductionMiddlewareConfig()
    
    opts := &httputils.RouterSetupOptions{
        LoggerColor: false,
        Middleware:  config,
    }
    
    r := httputils.SetupRouter(opts)
    
    r.Get("/api/health", healthHandler)
    
    http.ListenAndServe(":8080", r)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    response := map[string]string{"status": "ok"}
    httputils.NewResponse(w).OK(response)
}
```

## Middleware Configuration

### Production Configuration

The package provides a production-ready middleware configuration:

```go
func ProductionMiddlewareConfig() *MiddlewareConfig {
    return &MiddlewareConfig{
        CleanPath:        true,  // Clean double slashes
        StripSlashes:     true,  // Remove trailing slashes
        Recoverer:        true,  // Panic recovery
        RealIP:           true,  // Real IP detection
        AllowContentType: true,  // Content-Type validation
        Compress:         true,  // Response compression
        RequestSize:      true,  // Request size limiting (20MB)
    }
}
```

### Custom Configuration

```go
config := &httputils.MiddlewareConfig{
    CleanPath:        true,
    StripSlashes:     true,
    Recoverer:        true,
    RealIP:           true,
    RequestLogger:    true,  // Enable request logging
    AllowContentType: false, // Disable content type validation
    Compress:         true,
    RequestSize:      true,
}

opts := &httputils.RouterSetupOptions{
    LoggerColor: true,  // Colored logs for development
    Logger:      log.New(os.Stdout, "", log.LstdFlags),
    Middleware:  config,
}

r := httputils.SetupRouter(opts)
```

## Request Handling

### JSON Request Body Parsing

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    // Parse JSON body with type safety
    body, ok := httputils.GetBody[CreateUserRequest](w, r)
    if !ok {
        // Error response already written
        return
    }
    
    // Process the request
    user := createUser(body.Name, body.Email)
    
    // Return success response
    httputils.NewResponse(w).OK(user)
}
```

### Manual JSON Reading

```go
func manualJSONHandler(w http.ResponseWriter, r *http.Request) {
    data, err := httputils.ReadJSON[CreateUserRequest](r.Body)
    if err != nil {
        httputils.NewResponse(w).BadRequest("Invalid JSON")
        return
    }
    
    // Process data...
}
```

## Response Handling

### Fluent Response API

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    resp := httputils.NewResponse(w)
    
    // JSON responses
    resp.OK(map[string]string{"message": "success"})
    resp.Created(user)
    resp.BadRequest("Invalid input")
    resp.Unauthorized("Token required")
    resp.Forbidden("Access denied")
    resp.NotFound("User not found")
    resp.InternalServerError("Something went wrong")
    
    // Text responses
    resp.Text().OK("Plain text response")
    
    // Custom status codes
    resp.Status(http.StatusTeapot).OK("I'm a teapot")
}
```

### Cookie Management

```go
func setCookieHandler(w http.ResponseWriter, r *http.Request) {
    cookie := &http.Cookie{
        Name:     "session",
        Value:    "abc123",
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
    }
    
    httputils.NewResponse(w).
        Cookie(cookie).
        OK(map[string]string{"message": "Cookie set"})
}
```

### Response Timeouts

```go
func slowHandler(w http.ResponseWriter, r *http.Request) {
    resp := httputils.NewResponse(w)
    
    // Set response timeout
    resp.SetWriteDeadline(time.Now().Add(5 * time.Second))
    
    // Long operation...
    time.Sleep(3 * time.Second)
    
    resp.OK("Completed")
}
```

## Custom Encoders

### Creating Custom Encoders

```go
type XMLEncoder struct{}

func (x XMLEncoder) Encode(w http.ResponseWriter, status int, data ...any) {
    w.Header().Set("Content-Type", "application/xml")
    w.WriteHeader(status)
    
    if len(data) > 0 {
        // Custom XML encoding logic
        xml.NewEncoder(w).Encode(data[0])
    }
}

func xmlHandler(w http.ResponseWriter, r *http.Request) {
    resp := httputils.NewResponse(w).SetEncoder(XMLEncoder{})
    resp.OK(map[string]string{"message": "XML response"})
}
```

## Error Handling

### Standardized Error Responses

```go
func errorHandler(w http.ResponseWriter, r *http.Request) {
    resp := httputils.NewResponse(w)
    
    // Standard error format
    resp.BadRequest("Validation failed")
    // Returns: {"error": "Validation failed"}
    
    // Custom error structure
    errorData := map[string]interface{}{
        "error":   "validation_failed",
        "details": []string{"name is required", "email is invalid"},
    }
    resp.BadRequest(errorData)
}
```

### Validation Integration

```go
func validateAndCreateUser(w http.ResponseWriter, r *http.Request) {
    body, ok := httputils.GetBody[CreateUserRequest](w, r)
    if !ok {
        return
    }
    
    // Validate using zog (built-in support)
    schema := zog.Struct(zog.Schema{
        "name":  zog.String().Required(),
        "email": zog.String().Email().Required(),
    })
    
    if err := schema.Parse(body); err != nil {
        httputils.NewResponse(w).BadRequest(err.Error())
        return
    }
    
    // Process validated data...
    httputils.NewResponse(w).Created(body)
}
```

## Advanced Features

### Middleware Chain Example

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            httputils.NewResponse(w).Unauthorized("Token required")
            return
        }
        
        // Validate token...
        next.ServeHTTP(w, r)
    })
}

func setupAPI() *chi.Mux {
    r := httputils.SetupRouter(&httputils.RouterSetupOptions{
        Middleware: httputils.ProductionMiddlewareConfig(),
    })
    
    r.Route("/api", func(r chi.Router) {
        r.Use(authMiddleware)
        r.Post("/users", createUserHandler)
        r.Get("/users/{id}", getUserHandler)
    })
    
    return r
}
```

### Content Negotiation

```go
func negotiatedHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{"message": "Hello, World!"}
    
    accept := r.Header.Get("Accept")
    resp := httputils.NewResponse(w)
    
    switch accept {
    case "application/json":
        resp.JSON().OK(data)
    case "text/plain":
        resp.Text().OK("Hello, World!")
    default:
        resp.JSON().OK(data) // Default to JSON
    }
}
```

## Testing

### Testing HTTP Handlers

```go
func TestCreateUser(t *testing.T) {
    body := CreateUserRequest{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    jsonBody, _ := json.Marshal(body)
    req := httptest.NewRequest("POST", "/users", bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    createUserHandler(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response User
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", response.Name)
}
```

## Configuration Reference

### MiddlewareConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `CleanPath` | `bool` | Remove double slashes from URLs | `false` |
| `StripSlashes` | `bool` | Remove trailing slashes | `false` |
| `Recoverer` | `bool` | Panic recovery middleware | `false` |
| `RealIP` | `bool` | Extract real IP from headers | `false` |
| `RequestLogger` | `bool` | Enable request logging | `false` |
| `AllowContentType` | `bool` | Validate Content-Type headers | `false` |
| `Compress` | `bool` | Enable response compression | `false` |
| `RequestSize` | `bool` | Limit request size to 20MB | `false` |

### RouterSetupOptions

| Field | Type | Description |
|-------|------|-------------|
| `LoggerColor` | `bool` | Enable colored logging |
| `Logger` | `*log.Logger` | Custom logger instance |
| `Middleware` | `*MiddlewareConfig` | Middleware configuration |
| `BeforeRun` | `func(*chi.Mux)` | Callback before applying middleware |

## Performance Tips

1. **Use production middleware config** for better performance
2. **Enable compression** for smaller response sizes
3. **Set request size limits** to prevent DoS attacks
4. **Use type-safe request parsing** to avoid runtime errors
5. **Leverage response caching** where appropriate

## Back to Main Documentation

← [Back to main README](../README.md)