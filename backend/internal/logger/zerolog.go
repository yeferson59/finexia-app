package logger

import (
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

func applyFields(e *zerolog.Event, fields []Field) *zerolog.Event {
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

func (l *zeroLogger) Debug(msg string, fields ...Field) {
	applyFields(l.zl.Debug(), fields).Msg(msg)
}

func (l *zeroLogger) Info(msg string, fields ...Field) {
	applyFields(l.zl.Info(), fields).Msg(msg)
}

func (l *zeroLogger) Warn(msg string, fields ...Field) {
	applyFields(l.zl.Warn(), fields).Msg(msg)
}

func (l *zeroLogger) Error(msg string, fields ...Field) {
	applyFields(l.zl.Error(), fields).Msg(msg)
}

func (l *zeroLogger) Fatal(msg string, fields ...Field) {
	applyFields(l.zl.Fatal(), fields).Msg(msg)
}

// With returns a child logger with the given fields pre-attached.
// The parent logger is not modified.
func (l *zeroLogger) With(fields ...Field) Logger {
	ctx := l.zl.With()
	for _, f := range fields {
		switch f.kind {
		case fieldStr:
			ctx = ctx.Str(f.key, f.strVal)
		case fieldErr:
			if f.anyVal != nil {
				ctx = ctx.Err(f.anyVal.(error))
			}
		case fieldInt:
			ctx = ctx.Int(f.key, int(f.intVal))
		case fieldInt64:
			ctx = ctx.Int64(f.key, f.intVal)
		case fieldBool:
			ctx = ctx.Bool(f.key, f.boolVal)
		case fieldFloat64:
			ctx = ctx.Float64(f.key, f.floatVal)
		case fieldTime:
			ctx = ctx.Time(f.key, f.timeVal)
		case fieldDur:
			ctx = ctx.Dur(f.key, f.durVal)
		case fieldAny:
			ctx = ctx.Interface(f.key, f.anyVal)
		}
	}
	return &zeroLogger{zl: ctx.Logger()}
}
