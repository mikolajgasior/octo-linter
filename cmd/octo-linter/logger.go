package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
)

type customHandler struct {
	opts  *slog.HandlerOptions
	attrs []slog.Attr
	group string
}

func (h *customHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("level=%s\n", r.Level.String()))
	buf.WriteString(fmt.Sprintf("msg=%q\n", r.Message))
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(fmt.Sprintf("%s=%q\n", a.Key, a.Value.String()))
		return true
	})
	buf.WriteString("---\n")

	fmt.Fprint(os.Stderr, buf.String())
	return nil
}

func (h *customHandler) clone() *customHandler {
	return &customHandler{
		opts:  h.opts,
		attrs: slices.Clone(h.attrs),
		group: h.group,
	}
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := h.clone()
	c.attrs = append(c.attrs, attrs...)
	return c
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	c := h.clone()
	c.group = name
	return c
}
