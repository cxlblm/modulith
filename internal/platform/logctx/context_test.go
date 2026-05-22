package logctx

import (
	"context"
	"log/slog"
	"testing"
)

func TestAttrs(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want map[string]string
	}{
		{
			name: "empty context",
			ctx:  context.Background(),
			want: map[string]string{},
		},
		{
			name: "request id",
			ctx:  WithRequestID(context.Background(), "req-123"),
			want: map[string]string{"request_id": "req-123"},
		},
		{
			name: "user uuid",
			ctx:  WithUserUUID(context.Background(), "user-123"),
			want: map[string]string{"user_uuid": "user-123"},
		},
		{
			name: "request id and user uuid",
			ctx:  WithUserUUID(WithRequestID(context.Background(), "req-123"), "user-123"),
			want: map[string]string{"request_id": "req-123", "user_uuid": "user-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attrsToMap(Attrs(tt.ctx))
			if len(got) != len(tt.want) {
				t.Fatalf("len(Attrs) = %d, want %d; attrs = %#v", len(got), len(tt.want), got)
			}
			for key, wantValue := range tt.want {
				if got[key] != wantValue {
					t.Fatalf("Attrs[%q] = %q, want %q", key, got[key], wantValue)
				}
			}
		})
	}
}

func attrsToMap(attrs []slog.Attr) map[string]string {
	result := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		result[attr.Key] = attr.Value.String()
	}
	return result
}
