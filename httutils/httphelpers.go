package httputils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

// MiddlewareConfig controls which middlewares are applied
type MiddlewareConfig struct {
	// Core middlewares
	CleanPath     bool
	StripSlashes  bool
	Recoverer     bool
	RealIP        bool
	RequestLogger bool

	// Content middlewares
	AllowContentType bool
	Compress         bool
	RequestSize      bool
}

// ProductionMiddlewareConfig returns the production middleware configuration
func ProductionMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		CleanPath:        true,
		StripSlashes:     true,
		Recoverer:        true,
		RealIP:           true,
		AllowContentType: true,
		Compress:         true,
		RequestSize:      true,
	}
}

// RouterSetupOptions configures the router setup
type RouterSetupOptions struct {
	LoggerColor bool
	Logger      *log.Logger
	Middleware  *MiddlewareConfig

	BeforeRun func(r *chi.Mux)
}

// SetupRouter creates a chi router with configurable middlewares
func SetupRouter(opts *RouterSetupOptions) *chi.Mux {
	r := chi.NewRouter()

	if opts.BeforeRun != nil {
		opts.BeforeRun(r)
	}

	// Set up logger if provided
	if opts.Logger != nil && opts.Middleware.RequestLogger {
		middleware.DefaultLogger = middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger:  opts.Logger,
			NoColor: opts.LoggerColor,
		})
	}

	// Apply middlewares based on configuration
	if opts.Middleware.CleanPath {
		r.Use(middleware.CleanPath)
	}
	if opts.Middleware.StripSlashes {
		r.Use(middleware.StripSlashes)
	}
	if opts.Middleware.Recoverer {
		r.Use(middleware.Recoverer)
	}
	if opts.Middleware.RequestLogger {
		r.Use(middleware.Logger)
	}
	if opts.Middleware.RealIP {
		r.Use(middleware.RealIP)
	}
	if opts.Middleware.AllowContentType {
		r.Use(middleware.AllowContentType("application/json", "multipart/form-data"))
	}
	if opts.Middleware.Compress {
		r.Use(middleware.Compress(5, "brotli", "gzip", "deflate"))
	}
	if opts.Middleware.RequestSize {
		r.Use(middleware.RequestSize(20 * 1024 * 1024))
	}

	return r
}

func ReadJSON[T any](r io.Reader) (T, error) {
	var data T

	if closer, ok := r.(io.ReadCloser); ok {
		defer func() {
			_ = closer.Close()
		}()
	}

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields() // approximate RejectUnknownMembers
	if err := dec.Decode(&data); err != nil {
		var empty T

		return empty, err
	}

	return data, nil
}

func GetBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		defer func(body io.ReadCloser) {
			_ = body.Close()
		}(r.Body)

		value, err := ReadJSON[T](r.Body)
		if err != nil {
			goto errors
		}

		return value, true
	default:
		goto errors
	}

errors:
	allowsJSON := slices.Contains(r.Header.Values("Accept"), "application/json")

	if allowsJSON {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "unsupported Content-Type"}`))
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`unsupported Content-Type`))
	}

	var t T

	return t, false
}
