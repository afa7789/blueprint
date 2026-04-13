package logger

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
)

type Logger struct {
	repo   domain.AppLogRepository
	source string
}

func New(repo domain.AppLogRepository, source string) *Logger {
	return &Logger{repo: repo, source: source}
}

func (l *Logger) Info(ctx context.Context, msg string, meta ...map[string]interface{}) {
	l.write(ctx, "info", msg, meta)
}

func (l *Logger) Warn(ctx context.Context, msg string, meta ...map[string]interface{}) {
	l.write(ctx, "warn", msg, meta)
}

func (l *Logger) Error(ctx context.Context, msg string, meta ...map[string]interface{}) {
	l.write(ctx, "error", msg, meta)
}

func (l *Logger) Debug(ctx context.Context, msg string, meta ...map[string]interface{}) {
	l.write(ctx, "debug", msg, meta)
}

func (l *Logger) write(ctx context.Context, level, msg string, meta []map[string]interface{}) {
	_ = ctx
	// Always write to stdout
	log.Printf("[%s] [%s] %s", level, l.source, msg)

	// Write to DB if repo available
	if l.repo == nil {
		return
	}

	var metadata json.RawMessage
	if len(meta) > 0 {
		metadata, _ = json.Marshal(meta[0])
	}

	src := l.source
	entry := &domain.AppLog{
		Level:     level,
		Message:   msg,
		Source:    &src,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	// Non-blocking DB write
	go func() {
		_ = l.repo.Create(context.Background(), entry)
	}()
}
