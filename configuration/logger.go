package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config estructura las opciones de inicialización
type Config struct {
	LogPath string
	LokiURL string
	NodeID  string
	Debug   bool
}

var l *slog.Logger

func SetupLoggerTx(cfg Config) {
	var handlers []slog.Handler

	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	// 1. ARCHIVO LOCAL (Lumberjack)
	if cfg.LogPath != "" {
		dir := filepath.Dir(cfg.LogPath)
		_ = os.MkdirAll(dir, 0744)

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.LogPath,
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		}
		handlers = append(handlers, slog.NewJSONHandler(fileWriter, opts))
	}

	// 2. LOKI (HTTP directo, sin loki-client-go)
	if cfg.LokiURL != "" {
		lokiHandler := NewLokiHandler(LokiHandlerOptions{
			Level: level,
			URL:   cfg.LokiURL,
			Labels: map[string]string{
				"app":     "veredarii",
				"node_id": cfg.NodeID,
			},
		})
		handlers = append(handlers, lokiHandler)
	}

	// 3. STDOUT fallback
	if len(handlers) == 0 || cfg.Debug {
		handlers = append(handlers, slog.NewTextHandler(os.Stdout, opts))
	}

	combinedHandler := MultiHandler(handlers...)

	l = slog.New(combinedHandler).With(
		slog.String("node", cfg.NodeID),
	)
}

// ── Loki HTTP Handler ────────────────────────────────────────────────────────

type LokiHandlerOptions struct {
	Level  slog.Level
	URL    string
	Labels map[string]string
}

type lokiHandler struct {
	opts   LokiHandlerOptions
	attrs  []slog.Attr
	groups []string
	client *http.Client
}

func NewLokiHandler(opts LokiHandlerOptions) slog.Handler {
	return &lokiHandler{
		opts:   opts,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *lokiHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level
}

func (h *lokiHandler) Handle(_ context.Context, r slog.Record) error {
	// Construir línea de log como JSON
	fields := map[string]any{
		"msg":   r.Message,
		"level": r.Level.String(),
		"time":  r.Time.Format(time.RFC3339),
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	line, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	// Payload Loki push API v1
	ts := strconv.FormatInt(r.Time.UnixNano(), 10)
	payload := map[string]any{
		"streams": []map[string]any{
			{
				"stream": h.opts.Labels,
				"values": [][]string{{ts, string(line)}},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := h.client.Post(h.opts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("loki: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (h *lokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &lokiHandler{opts: h.opts, attrs: newAttrs, groups: h.groups, client: h.client}
}

func (h *lokiHandler) WithGroup(name string) slog.Handler {
	newGroups := append(append([]string{}, h.groups...), name)
	return &lokiHandler{opts: h.opts, attrs: h.attrs, groups: newGroups, client: h.client}
}

// ── MultiHandler ─────────────────────────────────────────────────────────────

func MultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

type multiHandler struct{ handlers []slog.Handler }

func (m *multiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		_ = h.Handle(ctx, r)
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

// ── WriteLog ──────────────────────────────────────────────────────────────────

func WriteLog(msg string, args ...any) {
	if l == nil {
		l = slog.Default()
	}
	l.Log(context.Background(), slog.LevelInfo, msg, args...)
}
