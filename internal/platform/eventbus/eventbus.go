package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type EventID string

type EventType string

type Envelope struct {
	EventID   EventID         `json:"event_id"`
	EventType EventType       `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

func (e Envelope) Decode(dst any) error {
	if err := json.Unmarshal(e.Payload, dst); err != nil {
		return fmt.Errorf("decode event payload: %w", err)
	}
	return nil
}

type Handler interface {
	Handle(ctx context.Context, env Envelope) error
}

type HandlerFunc func(ctx context.Context, env Envelope) error

func (fn HandlerFunc) Handle(ctx context.Context, env Envelope) error {
	return fn(ctx, env)
}

type Bus interface {
	Publish(ctx context.Context, eventType EventType, payload any) error
	Subscribe(eventType EventType, handler Handler)
}

type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
}

func New() *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[EventType][]Handler),
	}
}

func (b *InMemoryBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *InMemoryBus) Publish(ctx context.Context, eventType EventType, payload any) error {
	env, err := NewEnvelope(eventType, payload)
	if err != nil {
		return err
	}

	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[eventType]...)
	b.mu.RUnlock()

	var errs []error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, env); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func NewEnvelope(eventType EventType, payload any) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("encode event payload: %w", err)
	}
	return Envelope{
		EventID:   nextEventID(),
		EventType: eventType,
		Payload:   raw,
	}, nil
}

var eventSeq uint64

func nextEventID() EventID {
	seq := atomic.AddUint64(&eventSeq, 1)
	return EventID(fmt.Sprintf("evt-%d-%d", time.Now().UnixNano(), seq))
}
