package logger

type noopLogger struct{}

// Noop returns a Logger that discards all output.
// Use in unit tests to avoid nil panics without polluting test output.
// Fatal does NOT call os.Exit — safe to use in tests.
func Noop() Logger { return &noopLogger{} }

func (n *noopLogger) Debug(_ string, _ ...Field) {}
func (n *noopLogger) Info(_ string, _ ...Field)  {}
func (n *noopLogger) Warn(_ string, _ ...Field)  {}
func (n *noopLogger) Error(_ string, _ ...Field) {}
func (n *noopLogger) Fatal(_ string, _ ...Field) {}
func (n *noopLogger) With(_ ...Field) Logger     { return n }
