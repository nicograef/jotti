// Package logging carries the request logger through the context.
package logging

import (
	"context"
	"io"
	"log/slog"
)

type contextKey struct{}

var discard = slog.New(slog.DiscardHandler)

// New returns a JSON logger that writes every level from debug up to w.
func New(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// WithContext returns a copy of ctx that carries logger.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext returns the logger WithContext stored in ctx; without one it discards every record.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return logger
	}
	return discard
}
