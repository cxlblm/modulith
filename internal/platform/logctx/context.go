package logctx

import (
	"context"
	"log/slog"
)

type requestIDKey struct{}
type userUUIDKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func WithUserUUID(ctx context.Context, userUUID string) context.Context {
	if userUUID == "" {
		return ctx
	}
	return context.WithValue(ctx, userUUIDKey{}, userUUID)
}

func Attrs(ctx context.Context) []slog.Attr {
	attrs := make([]slog.Attr, 0, 2)
	if requestID, ok := ctx.Value(requestIDKey{}).(string); ok && requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}
	if userUUID, ok := ctx.Value(userUUIDKey{}).(string); ok && userUUID != "" {
		attrs = append(attrs, slog.String("user_uuid", userUUID))
	}
	return attrs
}
