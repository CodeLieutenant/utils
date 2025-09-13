package httputils

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Oudwins/zog"
)

type ResponseEncoder interface {
	Encode(http.ResponseWriter, int, ...any)
}

type Response struct {
	w          http.ResponseWriter
	encoder    ResponseEncoder
	controller *http.ResponseController
	cookies    []*http.Cookie
}

func NewResponse(w http.ResponseWriter) Response {
	return Response{
		w:          w,
		controller: http.NewResponseController(w),
		encoder:    JSON{},
	}
}

func (r Response) JSON() Response {
	r.encoder = JSON{}

	return r
}

func (r Response) Text() Response {
	r.encoder = Text{}

	return r
}

func (r Response) SetEncoder(encoder ResponseEncoder) Response {
	r.encoder = encoder

	return r
}

func (r Response) OK(data ...any) {
	r.encoder.Encode(r.w, http.StatusOK, data...)
}

func (r Response) Created(data ...any) {
	r.encoder.Encode(r.w, http.StatusCreated, data...)
}

func (r Response) Error(status int, err string) {
	r.encoder.Encode(r.w, status, ErrorMessage{Message: err})
}

func (r Response) InvalidBodyError() {
	r.Error(http.StatusBadRequest, "invalid body")
}

func (r Response) Unauthorized() {
	r.Error(http.StatusUnauthorized, "unauthorized")
}

func (r Response) ForbiddenError() {
	r.Error(http.StatusForbidden, "forbidden")
}

func (r Response) NotFoundError() {
	r.Error(http.StatusNotFound, "requested resource not found")
}

func (r Response) ConflictError() {
	r.Error(http.StatusConflict, "resource already exists")
}

func (r Response) InternalServerError() {
	r.Error(http.StatusInternalServerError, "internal server error, please try again later")
}

func (r Response) ServiceUnavailableError() {
	r.Error(http.StatusServiceUnavailable, "service unavailable, please try again later")
}

func (r Response) BadRequest() {
	r.Error(http.StatusBadRequest, "bad request")
}

func (r Response) ValidationError(err zog.ZogIssueMap) {
	r.encoder.Encode(r.w, http.StatusUnprocessableEntity, map[string]any{
		"message": "validation error",
		"errors":  zog.Issues.SanitizeMapAndCollect(err),
	})
}

func (r Response) NoContent() {
	r.w.WriteHeader(http.StatusNoContent)
}

func (r Response) Redirect(url string) {
	r.w.Header().Set("Location", url)
	r.w.WriteHeader(http.StatusFound)
}

func (r Response) ContentType(contentType string) Response {
	r.w.Header().Set("Content-Type", contentType)

	return r
}

func (r Response) SetHeader(key, value string) Response {
	r.w.Header().Set(key, value)

	return r
}

func (r Response) ReadFrom(src io.Reader) (int64, error) {
	return r.w.(io.ReaderFrom).ReadFrom(src) //nolint:forcetypeassert
}

func (r Response) SetCookie(cookie *http.Cookie) Response {
	r.cookies = append(r.cookies, cookie)

	return r
}

func (r Response) Write(data []byte) (int, error) {
	if len(r.cookies) > 0 {
		for _, cookie := range r.cookies {
			http.SetCookie(r.w, cookie)
		}
	}

	return r.w.Write(data)
}

func (r Response) Flush() error {
	return r.controller.Flush()
}

func (r Response) SetReadDeadline(deadline time.Time) Response {
	_ = r.controller.SetReadDeadline(deadline)

	return r
}

func (r Response) SetWriteDeadline(deadline time.Time) Response {
	_ = r.controller.SetWriteDeadline(deadline)

	return r
}

func (r Response) EnableFullDuplex() Response {
	_ = r.controller.EnableFullDuplex()

	return r
}

func (r Response) SendFile(request *http.Request, name string, modTime time.Time, content io.ReadSeeker) {
	http.ServeContent(r.w, request, name, modTime, content)
}

type JSON struct{}

type Text struct{}

func (JSON) Encode(w http.ResponseWriter, status int, data ...any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if len(data) > 0 {
		bytes, err := json.Marshal(data[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("{}"))
			slog.Error("failed to encode response", "error", err)

			return
		}

		_, _ = w.Write(bytes)

		return
	}

	_, _ = w.Write([]byte("{}"))
}

func (Text) Encode(w http.ResponseWriter, status int, data ...any) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)

	if len(data) > 0 {
		_, _ = w.Write([]byte(data[0].(string))) //nolint:forcetypeassert

		return
	}
}
