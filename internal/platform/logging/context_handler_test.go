package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"modular_monolith/internal/platform/logctx"
)

func TestNewLogger_AddsContextAttrs(t *testing.T) {
	var logs bytes.Buffer
	logger, err := New(Config{}, &logs)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	ctx := logctx.WithUserUUID(logctx.WithRequestID(context.Background(), "req-123"), "user-123")
	logger.InfoContext(ctx, "request scoped log", "component", "test")

	entry := decodeLogEntry(t, logs.Bytes())
	if entry["request_id"] != "req-123" {
		t.Fatalf("request_id = %v, want %q; entry = %#v", entry["request_id"], "req-123", entry)
	}
	if entry["user_uuid"] != "user-123" {
		t.Fatalf("user_uuid = %v, want %q; entry = %#v", entry["user_uuid"], "user-123", entry)
	}
	if entry["component"] != "test" {
		t.Fatalf("component = %v, want %q; entry = %#v", entry["component"], "test", entry)
	}
}

func TestNewLogger_SkipsEmptyContextAttrs(t *testing.T) {
	var logs bytes.Buffer
	logger, err := New(Config{}, &logs)
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	logger.InfoContext(context.Background(), "process log")

	entry := decodeLogEntry(t, logs.Bytes())
	if _, ok := entry["request_id"]; ok {
		t.Fatalf("request_id present in process log: %#v", entry)
	}
	if _, ok := entry["user_uuid"]; ok {
		t.Fatalf("user_uuid present in process log: %#v", entry)
	}
}

func decodeLogEntry(t *testing.T, data []byte) map[string]any {
	t.Helper()

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v; logs = %s", err, string(data))
	}
	return entry
}
