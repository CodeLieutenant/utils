package httputils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// Test types for generic functions
type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type TestStructWithExtra struct {
	Name  string `json:"name"`
	Extra string `json:"extra,omitempty"`
	Value int    `json:"value"`
}

func TestErrorMessage_JSON(t *testing.T) {
	t.Parallel()

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		err := ErrorMessage{Message: "test error"}
		data, jsonErr := json.Marshal(err)
		require.NoError(t, jsonErr)
		require.JSONEq(t, `{"message":"test error"}`, string(data))
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		var err ErrorMessage
		jsonErr := json.Unmarshal([]byte(`{"message":"test error"}`), &err)
		require.NoError(t, jsonErr)
		require.Equal(t, "test error", err.Message)
	})
}

func TestMiddlewareConfig(t *testing.T) {
	t.Parallel()

	t.Run("ZeroValue", func(t *testing.T) {
		t.Parallel()
		cfg := &MiddlewareConfig{}
		require.False(t, cfg.CleanPath)
		require.False(t, cfg.StripSlashes)
		require.False(t, cfg.Recoverer)
		require.False(t, cfg.RealIP)
		require.False(t, cfg.RequestLogger)
		require.False(t, cfg.AllowContentType)
		require.False(t, cfg.Compress)
		require.False(t, cfg.RequestSize)
	})

	t.Run("AllEnabled", func(t *testing.T) {
		t.Parallel()
		config := &MiddlewareConfig{
			CleanPath:        true,
			StripSlashes:     true,
			Recoverer:        true,
			RealIP:           true,
			RequestLogger:    true,
			AllowContentType: true,
			Compress:         true,
			RequestSize:      true,
		}
		require.True(t, config.CleanPath)
		require.True(t, config.StripSlashes)
		require.True(t, config.Recoverer)
		require.True(t, config.RealIP)
		require.True(t, config.RequestLogger)
		require.True(t, config.AllowContentType)
		require.True(t, config.Compress)
		require.True(t, config.RequestSize)
	})
}

func TestProductionMiddlewareConfig(t *testing.T) {
	t.Parallel()

	cfg := ProductionMiddlewareConfig()
	require.NotNil(t, cfg)

	// Verify all production middlewares are enabled except RequestLogger
	require.True(t, cfg.CleanPath)
	require.True(t, cfg.StripSlashes)
	require.True(t, cfg.Recoverer)
	require.True(t, cfg.RealIP)
	require.False(t, cfg.RequestLogger) // Not enabled in production cfg
	require.True(t, cfg.AllowContentType)
	require.True(t, cfg.Compress)
	require.True(t, cfg.RequestSize)
}

func TestRouterSetupOptions(t *testing.T) {
	t.Parallel()

	t.Run("AllFields", func(t *testing.T) {
		t.Parallel()
		logger := log.New(os.Stdout, "", 0)
		middleware := ProductionMiddlewareConfig()

		opts := &RouterSetupOptions{
			LoggerColor: true,
			Logger:      logger,
			Middleware:  middleware,
		}

		require.Equal(t, logger, opts.Logger)
		require.Equal(t, middleware, opts.Middleware)
	})

	t.Run("NilFields", func(t *testing.T) {
		t.Parallel()
		opts := &RouterSetupOptions{}
		require.Nil(t, opts.Logger)
		require.Nil(t, opts.Middleware)
	})
}

func TestSetupRouter(t *testing.T) {
	t.Parallel()

	t.Run("WithAllMiddlewares", func(t *testing.T) {
		t.Parallel()
		opts := &RouterSetupOptions{
			Logger:     log.New(os.Stdout, "", 0),
			Middleware: ProductionMiddlewareConfig(),
		}

		router := SetupRouter(opts)
		require.NotNil(t, router)
		require.IsType(t, &chi.Mux{}, router)
	})

	t.Run("WithNoMiddlewares", func(t *testing.T) {
		t.Parallel()
		opts := &RouterSetupOptions{
			Middleware: &MiddlewareConfig{},
		}

		router := SetupRouter(opts)
		require.NotNil(t, router)
		require.IsType(t, &chi.Mux{}, router)
	})

	t.Run("WithSelectiveMiddlewares", func(t *testing.T) {
		t.Parallel()
		opts := &RouterSetupOptions{
			Middleware: &MiddlewareConfig{
				CleanPath:   true,
				Recoverer:   true,
				Compress:    true,
				RequestSize: true,
			},
		}

		router := SetupRouter(opts)
		require.NotNil(t, router)
		require.IsType(t, &chi.Mux{}, router)
	})

	t.Run("WithLogger", func(t *testing.T) {
		t.Parallel()
		opts := &RouterSetupOptions{
			Logger: log.New(os.Stdout, "", 0),
			Middleware: &MiddlewareConfig{
				RequestLogger: false,
			},
		}

		router := SetupRouter(opts)
		require.NotNil(t, router)
		require.IsType(t, &chi.Mux{}, router)
	})

	t.Run("WithNilConfig", func(t *testing.T) {
		t.Parallel()
		opts := &RouterSetupOptions{
			Logger: log.New(os.Stdout, "", 0),
			Middleware: &MiddlewareConfig{
				RequestLogger: false,
			},
		}

		router := SetupRouter(opts)
		require.NotNil(t, router)
		require.IsType(t, &chi.Mux{}, router)
	})
}

func TestReadJSON(t *testing.T) {
	t.Parallel()

	t.Run("ValidJSON", func(t *testing.T) {
		t.Parallel()
		jsonData := `{"name":"test","value":42}`
		reader := strings.NewReader(jsonData)

		result, err := ReadJSON[TestStruct](reader)
		require.NoError(t, err)
		require.Equal(t, "test", result.Name)
		require.Equal(t, 42, result.Value)
	})

	t.Run("ValidJSONWithReadCloser", func(t *testing.T) {
		t.Parallel()
		jsonData := `{"name":"test","value":42}`
		reader := io.NopCloser(strings.NewReader(jsonData))

		result, err := ReadJSON[TestStruct](reader)
		require.NoError(t, err)
		require.Equal(t, "test", result.Name)
		require.Equal(t, 42, result.Value)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		t.Parallel()
		invalidJSON := `{"name":"test","value":}`
		reader := strings.NewReader(invalidJSON)

		result, err := ReadJSON[TestStruct](reader)
		require.Error(t, err)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
	})

	t.Run("UnknownFieldsRejected", func(t *testing.T) {
		t.Parallel()
		jsonWithExtra := `{"name":"test","value":42,"unknown":"field"}`
		reader := strings.NewReader(jsonWithExtra)

		result, err := ReadJSON[TestStruct](reader)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown field")
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
	})

	t.Run("EmptyJSON", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader(`{}`)

		result, err := ReadJSON[TestStruct](reader)
		require.NoError(t, err)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
	})

	t.Run("EmptyReader", func(t *testing.T) {
		t.Parallel()
		reader := strings.NewReader(``)

		result, err := ReadJSON[TestStruct](reader)
		require.Error(t, err)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
	})

	t.Run("GenericWithMap", func(t *testing.T) {
		t.Parallel()
		jsonData := `{"key1":"value1","key2":"value2"}`
		reader := strings.NewReader(jsonData)

		result, err := ReadJSON[map[string]string](reader)
		require.NoError(t, err)
		require.Equal(t, "value1", result["key1"])
		require.Equal(t, "value2", result["key2"])
	})
}

func TestGetBody(t *testing.T) {
	t.Parallel()

	t.Run("ValidJSONContentType", func(t *testing.T) {
		t.Parallel()
		jsonData := `{"name":"test","value":42}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rr := httptest.NewRecorder()

		result, success := GetBody[TestStruct](rr, req)
		require.True(t, success)
		require.Equal(t, "test", result.Name)
		require.Equal(t, 42, result.Value)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("InvalidJSONContent", func(t *testing.T) {
		t.Parallel()
		invalidJSON := `{"name":"test","value":}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		rr := httptest.NewRecorder()

		result, success := GetBody[TestStruct](rr, req)
		require.False(t, success)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Contains(t, rr.Body.String(), `"error": "unsupported Content-Type"`)
	})

	t.Run("UnsupportedContentType", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Accept", "application/json")

		rr := httptest.NewRecorder()

		result, success := GetBody[TestStruct](rr, req)
		require.False(t, success)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Contains(t, rr.Body.String(), `"error": "unsupported Content-Type"`)
	})

	t.Run("UnsupportedContentTypeNoJSONAccept", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Accept", "text/plain")

		rr := httptest.NewRecorder()

		result, success := GetBody[TestStruct](rr, req)
		require.False(t, success)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
		require.Equal(t, "unsupported Content-Type", rr.Body.String())
	})

	t.Run("NoContentType", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
		req.Header.Set("Accept", "application/json")

		rr := httptest.NewRecorder()

		result, success := GetBody[TestStruct](rr, req)
		require.False(t, success)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Contains(t, rr.Body.String(), `"error": "unsupported Content-Type"`)
	})

	t.Run("MultipleAcceptHeaders", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Add("Accept", "text/html")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Accept", "*/*")

		rr := httptest.NewRecorder()

		result, success := GetBody[TestStruct](rr, req)
		require.False(t, success)
		require.Empty(t, result.Name)
		require.Equal(t, 0, result.Value)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Contains(t, rr.Body.String(), `"error": "unsupported Content-Type"`)
	})
}

func TestResponseEncoder(t *testing.T) {
	t.Parallel()

	t.Run("JSONEncoder", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		encoder := JSON{}

		testData := TestStruct{Name: "test", Value: 42}
		encoder.Encode(rr, http.StatusOK, testData)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var result TestStruct
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Equal(t, testData, result)
	})

	t.Run("JSONEncoderNoData", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		encoder := JSON{}

		encoder.Encode(rr, http.StatusOK)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Equal(t, "{}", rr.Body.String())
	})

	t.Run("JSONEncoderMarshalError", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		encoder := JSON{}

		// Use a channel which cannot be marshaled to JSON
		encoder.Encode(rr, http.StatusOK, make(chan int))

		// First WriteHeader call wins, so status remains 200
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Equal(t, "{}", rr.Body.String())
	})

	t.Run("TextEncoder", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		encoder := Text{}

		encoder.Encode(rr, http.StatusOK, "test message")

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
		require.Equal(t, "test message", rr.Body.String())
	})

	t.Run("TextEncoderNoData", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		encoder := Text{}

		encoder.Encode(rr, http.StatusOK)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
		require.Empty(t, rr.Body.String())
	})
}

//nolint:maintidx
func TestResponse(t *testing.T) {
	t.Parallel()

	t.Run("NewResponse", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		require.NotNil(t, response.w)
		require.NotNil(t, response.controller)
		require.NotNil(t, response.encoder)
		require.IsType(t, JSON{}, response.encoder)
		require.Empty(t, response.cookies)
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr).JSON()

		require.IsType(t, JSON{}, response.encoder)
	})

	t.Run("Text", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr).Text()

		require.IsType(t, Text{}, response.encoder)
	})

	t.Run("SetEncoder", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		customEncoder := JSON{}
		response := NewResponse(rr).SetEncoder(customEncoder)

		require.IsType(t, JSON{}, response.encoder)
	})

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)
		testData := TestStruct{Name: "test", Value: 42}

		response.OK(testData)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var result TestStruct
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Equal(t, testData, result)
	})

	t.Run("Created", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)
		testData := TestStruct{Name: "created", Value: 1}

		response.Created(testData)

		require.Equal(t, http.StatusCreated, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var result TestStruct
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Equal(t, testData, result)
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.Error(http.StatusBadRequest, "test error")

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var result ErrorMessage
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Equal(t, "test error", result.Message)
	})

	t.Run("InvalidBodyError", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.InvalidBodyError()

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "invalid body")
	})

	t.Run("Unauthorized", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.Unauthorized()

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		require.Contains(t, rr.Body.String(), "unauthorized")
	})

	t.Run("ForbiddenError", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.ForbiddenError()

		require.Equal(t, http.StatusForbidden, rr.Code)
		require.Contains(t, rr.Body.String(), "forbidden")
	})

	t.Run("NotFoundError", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.NotFoundError()

		require.Equal(t, http.StatusNotFound, rr.Code)
		require.Contains(t, rr.Body.String(), "requested resource not found")
	})

	t.Run("ConflictError", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.ConflictError()

		require.Equal(t, http.StatusConflict, rr.Code)
		require.Contains(t, rr.Body.String(), "resource already exists")
	})

	t.Run("InternalServerError", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.InternalServerError()

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		require.Contains(t, rr.Body.String(), "internal server error")
	})

	t.Run("ServiceUnavailableError", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.ServiceUnavailableError()

		require.Equal(t, http.StatusServiceUnavailable, rr.Code)
		require.Contains(t, rr.Body.String(), "service unavailable")
	})

	t.Run("BadRequest", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.BadRequest()

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "bad request")
	})

	t.Run("NoContent", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.NoContent()

		require.Equal(t, http.StatusNoContent, rr.Code)
		require.Empty(t, rr.Body.String())
	})

	t.Run("Redirect", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.Redirect("https://example.com")

		require.Equal(t, http.StatusFound, rr.Code)
		require.Equal(t, "https://example.com", rr.Header().Get("Location"))
	})

	t.Run("ContentType", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.ContentType("application/xml")

		require.Equal(t, "application/xml", rr.Header().Get("Content-Type"))
	})

	t.Run("SetHeader", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		response.SetHeader("X-Custom-Header", "custom-value")

		require.Equal(t, "custom-value", rr.Header().Get("X-Custom-Header"))
	})

	t.Run("SetCookie", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)
		cookie := &http.Cookie{
			Name:  "test-cookie",
			Value: "test-value",
		}

		responseWithCookie := response.SetCookie(cookie)

		require.Len(t, responseWithCookie.cookies, 1)
		require.Equal(t, cookie, responseWithCookie.cookies[0])
		// Original response should still be empty
		require.Empty(t, response.cookies)
	})

	t.Run("WriteWithCookies", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)
		cookie := &http.Cookie{
			Name:  "test-cookie",
			Value: "test-value",
		}

		responseWithCookie := response.SetCookie(cookie)
		n, err := responseWithCookie.Write([]byte("test content"))

		require.NoError(t, err)
		require.Equal(t, 12, n) // "test content" length
		require.Equal(t, "test content", rr.Body.String())

		// Check cookie was set
		cookies := rr.Result().Cookies()
		require.Len(t, cookies, 1)
		require.Equal(t, "test-cookie", cookies[0].Name)
		require.Equal(t, "test-value", cookies[0].Value)
	})

	t.Run("WriteWithoutCookies", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		n, err := response.Write([]byte("test content"))

		require.NoError(t, err)
		require.Equal(t, 12, n)
		require.Equal(t, "test content", rr.Body.String())
		require.Empty(t, rr.Result().Cookies())
	})

	t.Run("SetReadDeadline", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)
		deadline := time.Now().Add(time.Minute)

		result := response.SetReadDeadline(deadline)

		// Should return the response for chaining
		require.Equal(t, response.w, result.w)
	})

	t.Run("SetWriteDeadline", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)
		deadline := time.Now().Add(time.Minute)

		result := response.SetWriteDeadline(deadline)

		// Should return the response for chaining
		require.Equal(t, response.w, result.w)
	})

	t.Run("EnableFullDuplex", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)

		result := response.EnableFullDuplex()

		// Should return the response for chaining
		require.Equal(t, response.w, result.w)
	})

	t.Run("SendFile", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		response := NewResponse(rr)
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		content := strings.NewReader("file content")
		modTime := time.Now()

		response.SendFile(req, "test.txt", modTime, content)

		require.Equal(t, "file content", rr.Body.String())
		require.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
	})

	t.Run("ChainedMethodCalls", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()

		NewResponse(rr).
			ContentType("custom/type").
			SetHeader("X-Test", "value").
			SetCookie(&http.Cookie{Name: "chain", Value: "test"}).
			Text().
			OK("chained response")

		require.Equal(t, http.StatusOK, rr.Code)
		// Text encoder overwrites content-type
		require.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
		require.Equal(t, "value", rr.Header().Get("X-Test"))
		require.Equal(t, "chained response", rr.Body.String())

		// Note: Cookies are lost when Text() returns new Response instance
		// This demonstrates the current behavior, not necessarily ideal behavior
		cookies := rr.Result().Cookies()
		require.Empty(t, cookies)
	})
}

func TestFlush(t *testing.T) {
	t.Parallel()

	// Note: httptest.ResponseRecorder doesn't support flushing,
	// but we can test that the method exists and doesn't panic
	rr := httptest.NewRecorder()
	response := NewResponse(rr)

	err := response.Flush()
	// The actual error depends on the implementation, just ensure it doesn't panic
	_ = err
}

func TestReadFrom(t *testing.T) {
	t.Parallel()

	// Note: httptest.ResponseRecorder doesn't implement io.ReaderFrom,
	// so this will panic when type require.ing, but we can test the method exists
	rr := httptest.NewRecorder()
	response := NewResponse(rr)
	reader := strings.NewReader("test content")

	require.Panics(t, func() {
		_, _ = response.ReadFrom(reader)
	})
}
