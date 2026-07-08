package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Level mirrors zerolog levels so callers don't import zerolog directly.
type Level int8

const (
	LevelDebug Level = Level(zerolog.DebugLevel)
	LevelInfo  Level = Level(zerolog.InfoLevel)
	LevelWarn  Level = Level(zerolog.WarnLevel)
	LevelError Level = Level(zerolog.ErrorLevel)
	LevelFatal Level = Level(zerolog.FatalLevel)
)

// Config controls how the zerolog-backed Logger is built.
type Config struct {
	Level       Level
	Output      io.Writer // defaults to os.Stderr if nil
	Environment string    // "production" → JSON; anything else → ConsoleWriter (pretty)
}

type zeroLogger struct {
	zl zerolog.Logger
}

// New creates a zerolog-backed Logger.
// In development (non-production), output is human-readable via ConsoleWriter.
func New(cfg Config) Logger {
	out := cfg.Output
	if out == nil {
		out = os.Stderr
	}

	var writer io.Writer
	if cfg.Environment == "production" {
		writer = out
	} else {
		writer = zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: time.RFC3339,
		}
	}

	zl := zerolog.New(writer).
		Level(zerolog.Level(cfg.Level)).
		With().
		Timestamp().
		Logger()

	return &zeroLogger{zl: zl}
}

func applyFields(ctx context.Context, e *zerolog.Event, fields []Field) *zerolog.Event {
	e = e.Ctx(ctx)

	for _, f := range fields {
		switch f.kind {
		case fieldStr:
			e = e.Str(f.key, f.strVal)
		case fieldErr:
			if f.anyVal != nil {
				e = e.Err(f.anyVal.(error))
			}
		case fieldInt:
			e = e.Int(f.key, int(f.intVal))
		case fieldInt64:
			e = e.Int64(f.key, f.intVal)
		case fieldBool:
			e = e.Bool(f.key, f.boolVal)
		case fieldFloat64:
			e = e.Float64(f.key, f.floatVal)
		case fieldTime:
			e = e.Time(f.key, f.timeVal)
		case fieldDur:
			e = e.Dur(f.key, f.durVal)
		case fieldAny:
			e = e.Interface(f.key, f.anyVal)
		}
	}

	return e
}

func (l *zeroLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	applyFields(ctx, l.zl.Debug(), fields).Msg(msg)
}

func (l *zeroLogger) Info(ctx context.Context, msg string, fields ...Field) {
	applyFields(ctx, l.zl.Info(), fields).Msg(msg)
}

func (l *zeroLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	applyFields(ctx, l.zl.Warn(), fields).Msg(msg)
}

func (l *zeroLogger) Error(ctx context.Context, msg string, fields ...Field) {
	applyFields(ctx, l.zl.Error(), fields).Msg(msg)
}

func (l *zeroLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	applyFields(ctx, l.zl.Fatal(), fields).Msg(msg)
}

// With returns a child logger with the given fields pre-attached.
// The parent logger is not modified.
func (l *zeroLogger) With(fields ...Field) Logger {
	c := l.zl.With()
	for _, f := range fields {
		switch f.kind {
		case fieldStr:
			c = c.Str(f.key, f.strVal)
		case fieldErr:
			if f.anyVal != nil {
				c = c.Err(f.anyVal.(error))
			}
		case fieldInt:
			c = c.Int(f.key, int(f.intVal))
		case fieldInt64:
			c = c.Int64(f.key, f.intVal)
		case fieldBool:
			c = c.Bool(f.key, f.boolVal)
		case fieldFloat64:
			c = c.Float64(f.key, f.floatVal)
		case fieldTime:
			c = c.Time(f.key, f.timeVal)
		case fieldDur:
			c = c.Dur(f.key, f.durVal)
		case fieldAny:
			c = c.Interface(f.key, f.anyVal)
		}
	}
	return &zeroLogger{zl: c.Logger()}
}
