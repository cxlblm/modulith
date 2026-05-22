package httpserver

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/platform/logging"
)

const (
	testRequestID = "req-123"
	testUserUUID  = "user-123"
)

func TestServer_Healthz(t *testing.T) {
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("body.Status = %q, want %q", body.Status, "ok")
	}
}

func TestServer_ValidationErrorResponse(t *testing.T) {
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))

	srv.Echo().POST("/users", func(c *echo.Context) error {
		var req struct {
			Name string `json:"name" validate:"required"`
		}
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}
		return c.NoContent(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	var body ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error.Code != "validation_failed" {
		t.Fatalf("error code = %q, want %q", body.Error.Code, "validation_failed")
	}
	if len(body.Error.Fields) != 1 {
		t.Fatalf("len(fields) = %d, want %d", len(body.Error.Fields), 1)
	}
	if body.Error.Fields[0].Field != "name" {
		t.Fatalf("field = %q, want %q", body.Error.Fields[0].Field, "name")
	}
}

func TestServer_NotFoundErrorResponse(t *testing.T) {
	var logs bytes.Buffer
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(&logs, nil)))

	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}

	var body ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want %q", body.Error.Code, "not_found")
	}
	if body.Error.Message == "" {
		t.Fatal("error message is empty")
	}
	if bytes.Contains(logs.Bytes(), []byte("http request failed")) {
		t.Fatalf("logs contain server error entry: %s", logs.String())
	}
}

func TestServer_RequestIDHeader(t *testing.T) {
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	if rec.Header().Get(echo.HeaderXRequestID) == "" {
		t.Fatal("response X-Request-Id header is empty")
	}
}

func TestServer_PreservesRequestIDHeader(t *testing.T) {
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(echo.HeaderXRequestID, testRequestID)
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	if got := rec.Header().Get(echo.HeaderXRequestID); got != testRequestID {
		t.Fatalf("response X-Request-Id = %q, want %q", got, testRequestID)
	}
}

func TestServer_CleansMultipartFormOnRequestContextCopy(t *testing.T) {
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))

	var originalReq *http.Request
	var tempFileName string
	srv.Echo().POST("/upload", func(c *echo.Context) error {
		if c.Request() == originalReq {
			return echo.NewHTTPError(http.StatusInternalServerError, "request was not copied")
		}
		if err := c.Request().ParseMultipartForm(1); err != nil {
			return err
		}
		file, _, err := c.Request().FormFile("file")
		if err != nil {
			return err
		}
		defer file.Close()
		namedFile, ok := file.(interface{ Name() string })
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "multipart file is not backed by a temp file")
		}
		tempFileName = namedFile.Name()
		return c.NoContent(http.StatusNoContent)
	})

	body, contentType := multipartBody(t)
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	originalReq = req
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if tempFileName == "" {
		t.Fatal("multipart temp file name is empty")
	}
	if _, err := os.Stat(tempFileName); !os.IsNotExist(err) {
		t.Fatalf("multipart temp file still exists: %v", err)
	}
}

func TestServer_AccessLogIncludesRequestContext(t *testing.T) {
	var logs bytes.Buffer
	logger, err := logging.New(logging.Config{}, &logs)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, logger)
	srv.Echo().GET("/protected", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}, RequireUserAuth())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderXRequestID, testRequestID)
	req.Header.Set(HeaderUserUUID, testUserUUID)
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	entry := decodeAccessLogEntry(t, logs.Bytes())
	if entry["request_id"] != testRequestID {
		t.Fatalf("request_id = %v, want %q; entry = %#v", entry["request_id"], testRequestID, entry)
	}
	if entry["user_uuid"] != testUserUUID {
		t.Fatalf("user_uuid = %v, want %q; entry = %#v", entry["user_uuid"], testUserUUID, entry)
	}
	if entry["msg"] != "REQUEST" {
		t.Fatalf("msg = %v, want %q; entry = %#v", entry["msg"], "REQUEST", entry)
	}
	if entry["method"] != http.MethodGet {
		t.Fatalf("method = %v, want %q; entry = %#v", entry["method"], http.MethodGet, entry)
	}
	if entry["uri"] != "/protected" {
		t.Fatalf("uri = %v, want %q; entry = %#v", entry["uri"], "/protected", entry)
	}
	if entry["status"] != float64(http.StatusNoContent) {
		t.Fatalf("status = %v, want %d; entry = %#v", entry["status"], http.StatusNoContent, entry)
	}
	if entry["latency"] == "" {
		t.Fatalf("latency is empty; entry = %#v", entry)
	}
	if entry["host"] == "" {
		t.Fatalf("host is empty; entry = %#v", entry)
	}
	if entry["bytes_in"] != "" {
		t.Fatalf("bytes_in = %v, want empty string; entry = %#v", entry["bytes_in"], entry)
	}
	if entry["bytes_out"] != float64(0) {
		t.Fatalf("bytes_out = %v, want 0; entry = %#v", entry["bytes_out"], entry)
	}
	if entry["remote_ip"] == "" {
		t.Fatalf("remote_ip is empty; entry = %#v", entry)
	}
}

func TestServer_AccessLogHandlesErrorsWithHTTPErrorHandlerStatus(t *testing.T) {
	var logs bytes.Buffer
	logger, err := logging.New(logging.Config{}, &logs)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	srv := New(Config{
		Addr:            ":0",
		ReadTimeout:     time.Second,
		WriteTimeout:    time.Second,
		ShutdownTimeout: time.Second,
	}, logger)
	srv.Echo().GET("/broken", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusForbidden, "no access")
	})

	req := httptest.NewRequest(http.MethodGet, "/broken", nil)
	req.Header.Set(echo.HeaderXRequestID, testRequestID)
	rec := httptest.NewRecorder()

	srv.Echo().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	entries := decodeLogEntries(t, logs.Bytes())
	entry := findLogEntry(t, entries, "REQUEST_ERROR")
	if entry["request_id"] != testRequestID {
		t.Fatalf("request_id = %v, want %q; entry = %#v", entry["request_id"], testRequestID, entry)
	}
	if entry["status"] != float64(http.StatusForbidden) {
		t.Fatalf("status = %v, want %d; entry = %#v", entry["status"], http.StatusForbidden, entry)
	}
	if entry["method"] != http.MethodGet {
		t.Fatalf("method = %v, want %q; entry = %#v", entry["method"], http.MethodGet, entry)
	}
	if entry["uri"] != "/broken" {
		t.Fatalf("uri = %v, want %q; entry = %#v", entry["uri"], "/broken", entry)
	}
	if entry["error"] == "" {
		t.Fatalf("error is empty; entry = %#v", entry)
	}
}

func multipartBody(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := file.Write(bytes.Repeat([]byte("x"), 4096)); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return &body, writer.FormDataContentType()
}

func decodeAccessLogEntry(t *testing.T, data []byte) map[string]any {
	t.Helper()

	entries := decodeLogEntries(t, data)
	return findLogEntry(t, entries, "REQUEST")
}

func decodeLogEntries(t *testing.T, data []byte) []map[string]any {
	t.Helper()

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatalf("unmarshal log entry: %v; logs = %s", err, string(data))
		}
		entries = append(entries, entry)
	}
	return entries
}

func findLogEntry(t *testing.T, entries []map[string]any, msg string) map[string]any {
	t.Helper()

	for _, entry := range entries {
		if entry["msg"] == msg {
			return entry
		}
	}
	t.Fatalf("log entry %q not found; entries = %#v", msg, entries)
	return nil
}
