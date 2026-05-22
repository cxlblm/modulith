package eventbus

import (
	"context"
	"encoding/json"
	"testing"
)

type testPayload struct {
	ID string `json:"id"`
}

func TestBus_PublishDeliversEnvelopeToSubscribers(t *testing.T) {
	bus := New()
	var got Envelope

	bus.Subscribe(EventType("test.created"), HandlerFunc(func(ctx context.Context, env Envelope) error {
		got = env
		return nil
	}))

	err := bus.Publish(context.Background(), EventType("test.created"), testPayload{ID: "one"})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if got.EventID == "" {
		t.Fatal("EventID is empty")
	}
	if got.EventType != EventType("test.created") {
		t.Fatalf("EventType = %q, want %q", got.EventType, EventType("test.created"))
	}

	var payload testPayload
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.ID != "one" {
		t.Fatalf("payload.ID = %q, want %q", payload.ID, "one")
	}
}

func TestEnvelope_Decode(t *testing.T) {
	env, err := NewEnvelope(EventType("test.created"), testPayload{ID: "two"})
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	var payload testPayload
	if err := env.Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if payload.ID != "two" {
		t.Fatalf("payload.ID = %q, want %q", payload.ID, "two")
	}
}
