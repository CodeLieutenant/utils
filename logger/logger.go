package logger

import (
	"bufio"
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"sync"

	"github.com/CodeLieutenant/utils"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level          string // debug, info, warn, error
	Format         string // json, text
	Output         string // stdout, file
	FilePath       string // path to log file when output is file
	AddSource      bool
	AddStacktrace  bool // Add stacktrace to error logs
	AddRuntimeInfo bool // Add runtime info like memory stats, goroutine count
}

// StacktraceHandler wraps another slog.Handler and adds stacktraces to error logs
type StacktraceHandler struct {
	handler        slog.Handler
	hostname       string
	version        string
	pid            int
	addStacktrace  bool
	addRuntimeInfo bool
}

// NewStacktraceHandler creates a new StacktraceHandler
func NewStacktraceHandler(version string, handler slog.Handler, addStacktrace, addRuntimeInfo bool) *StacktraceHandler {
	hostname, err := os.Hostname()
	if err != nil {
		panic("failed to get hostname: " + err.Error())
	}

	pid := os.Getpid()

	return &StacktraceHandler{
		hostname:       hostname,
		pid:            pid,
		handler:        handler,
		addStacktrace:  addStacktrace,
		addRuntimeInfo: addRuntimeInfo,
		version:        version,
	}
}

func (h *StacktraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *StacktraceHandler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(slog.String("hostname", h.hostname),
		slog.String("version", h.version),
		slog.Int("pid", h.pid),
	)

	// Add stacktrace for error level logs
	if h.addStacktrace && record.Level >= slog.LevelError {
		// Skip frames: runtime.Callers, this method, Handle, and the logging call
		stack := make([]byte, 4096)
		n := runtime.Stack(stack, false)
		record.AddAttrs(slog.String("stacktrace", string(stack[:n])))
	}

	// Add runtime information for all logs if enabled
	if h.addRuntimeInfo {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		record.AddAttrs(
			slog.Int("goroutines", runtime.NumGoroutine()),
			slog.String("goVersion", runtime.Version()),
			slog.Uint64("memAlloc", m.Alloc),
			slog.Uint64("memTotalAlloc", m.TotalAlloc),
			slog.Uint64("memSys", m.Sys),
			slog.Uint64("numGC", uint64(m.NumGC)),
		)
	}

	return h.handler.Handle(ctx, record)
}

func (h *StacktraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewStacktraceHandler(h.version, h.handler.WithAttrs(attrs), h.addStacktrace, h.addRuntimeInfo)
}

func (h *StacktraceHandler) WithGroup(name string) slog.Handler {
	return NewStacktraceHandler(h.version, h.handler.WithGroup(name), h.addStacktrace, h.addRuntimeInfo)
}

func (l *LogConfig) Setup(version string) (*slog.Logger, *log.Logger, func() error, error) {
	var level slog.Level
	switch l.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var (
		writer io.Writer
		closer = func() error { return nil }
	)
	switch l.Output {
	case "file":
		absolutePath, err := utils.CreateDirectoryFromFile(l.FilePath, 0o744)
		if err != nil {
			return nil, nil, nil, err
		}

		file, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, nil, err
		}

		buffer := bufio.NewWriterSize(file, 32*1024)
		var once sync.Once
		closer = func() error {
			var closeErr error

			once.Do(func() {
				if err = buffer.Flush(); err != nil {
					slog.Error("failed to flush log buffer: ",
						"error", err,
						slog.String("file", l.FilePath),
					)
				}
				closeErr = file.Close()
			})

			return closeErr
		}
		writer = buffer
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	}

	opts := &slog.HandlerOptions{
		AddSource: l.AddSource,
		Level:     level,
	}

	var handler slog.Handler

	switch l.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	handler = NewStacktraceHandler(version, handler, l.AddStacktrace, l.AddRuntimeInfo)

	stdLogger := slog.NewLogLogger(handler, level)
	logger := slog.New(handler)

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	slog.SetDefault(logger)

	return logger, stdLogger, closer, nil
}
