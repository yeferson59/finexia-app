package logger

import (
	"context"
	"time"
)

type fieldKind uint8

const (
	fieldStr fieldKind = iota
	fieldErr
	fieldInt
	fieldInt64
	fieldBool
	fieldFloat64
	fieldTime
	fieldDur
	fieldAny
)

// Field is an opaque log field. Construct via helper functions only.
type Field struct {
	kind     fieldKind
	key      string
	strVal   string
	intVal   int64
	floatVal float64
	boolVal  bool
	timeVal  time.Time
	durVal   time.Duration
	anyVal   any
}

func Str(key, val string) Field {
	return Field{kind: fieldStr, key: key, strVal: val}
}

func Err(err error) Field {
	f := Field{kind: fieldErr, key: "error"}
	if err != nil {
		f.strVal = err.Error()
		f.anyVal = err
	}
	return f
}

func Int(key string, val int) Field {
	return Field{kind: fieldInt, key: key, intVal: int64(val)}
}

func Int64(key string, val int64) Field {
	return Field{kind: fieldInt64, key: key, intVal: val}
}

func Bool(key string, val bool) Field {
	return Field{kind: fieldBool, key: key, boolVal: val}
}

func Float64(key string, val float64) Field {
	return Field{kind: fieldFloat64, key: key, floatVal: val}
}

func Time(key string, val time.Time) Field {
	return Field{kind: fieldTime, key: key, timeVal: val}
}

func Dur(key string, val time.Duration) Field {
	return Field{kind: fieldDur, key: key, durVal: val}
}

func Any(key string, val any) Field {
	return Field{kind: fieldAny, key: key, anyVal: val}
}

// Logger is the injectable interface used by all application components.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	// Fatal logs at fatal level then calls os.Exit(1) (in the zerolog implementation).
	Fatal(ctx context.Context, msg string, fields ...Field)
	// With returns a child Logger with the given fields attached to every subsequent entry.
	With(fields ...Field) Logger
}
