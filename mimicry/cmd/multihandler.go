package main

import (
	"context"
	"errors"
	"log/slog"
)

// MultiHandler fans a single slog.Record out to several sub-handlers, each of which keeps its
// own level filter and formatting.
type MultiHandler []slog.Handler

func (h MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, sub := range h {
		if sub.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, sub := range h {
		if !sub.Enabled(ctx, r.Level) {
			continue
		}
		// Clone per sub: handlers are free to mutate the record via AddAttrs.
		if err := sub.Handle(ctx, r.Clone()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make(MultiHandler, len(h))
	for i, sub := range h {
		out[i] = sub.WithAttrs(attrs)
	}
	return out
}

func (h MultiHandler) WithGroup(name string) slog.Handler {
	out := make(MultiHandler, len(h))
	for i, sub := range h {
		out[i] = sub.WithGroup(name)
	}
	return out
}
