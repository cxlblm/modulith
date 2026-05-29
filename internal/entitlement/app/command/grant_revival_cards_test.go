package command

import (
	"context"
	"errors"
	"testing"
)

func TestGrantRevivalCardsHandler_RejectsInvalidCommand(t *testing.T) {
	handler := GrantRevivalCardsHandler{RevivalCards: &fakeRevivalCards{}}

	tests := []struct {
		name string
		cmd  GrantRevivalCards
	}{
		{name: "missing user", cmd: GrantRevivalCards{Count: 1}},
		{name: "zero count", cmd: GrantRevivalCards{UserID: "user-1"}},
		{name: "negative count", cmd: GrantRevivalCards{UserID: "user-1", Count: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.Background(), tt.cmd)
			if !errors.Is(err, ErrInvalidCommand) {
				t.Fatalf("Handle() error = %v, want %v", err, ErrInvalidCommand)
			}
		})
	}
}

func TestGrantRevivalCardsHandler_GrantsCards(t *testing.T) {
	revivals := &fakeRevivalCards{}
	handler := GrantRevivalCardsHandler{RevivalCards: revivals}

	if err := handler.Handle(context.Background(), GrantRevivalCards{UserID: "user-1", Count: 2}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if revivals.grantedUserID != "user-1" || revivals.grantedCount != 2 {
		t.Fatalf("grant = (%q, %d), want (user-1, 2)", revivals.grantedUserID, revivals.grantedCount)
	}
}

func TestTryConsumeRevivalCardHandler_ReturnsConsumedState(t *testing.T) {
	tests := []struct {
		name         string
		consumeReply bool
		wantConsumed bool
	}{
		{name: "consumed", consumeReply: true, wantConsumed: true},
		{name: "no balance", consumeReply: false, wantConsumed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			revivals := &fakeRevivalCards{consumeReply: tt.consumeReply}
			handler := TryConsumeRevivalCardHandler{RevivalCards: revivals}

			result, err := handler.Handle(context.Background(), TryConsumeRevivalCard{UserID: "user-1"})
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if result.Consumed != tt.wantConsumed {
				t.Fatalf("Consumed = %v, want %v", result.Consumed, tt.wantConsumed)
			}
		})
	}
}

func TestTryConsumeRevivalCardHandler_RejectsInvalidCommand(t *testing.T) {
	handler := TryConsumeRevivalCardHandler{RevivalCards: &fakeRevivalCards{}}

	_, err := handler.Handle(context.Background(), TryConsumeRevivalCard{})
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrInvalidCommand)
	}
}

type fakeRevivalCards struct {
	grantedUserID string
	grantedCount  int
	consumeReply  bool
}

func (f *fakeRevivalCards) Grant(_ context.Context, userID string, count int) error {
	f.grantedUserID = userID
	f.grantedCount = count
	return nil
}

func (f *fakeRevivalCards) TryConsumeOne(context.Context, string) (bool, error) {
	return f.consumeReply, nil
}
