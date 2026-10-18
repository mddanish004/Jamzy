package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const timestampLayout = "2006-01-02T15:04:05.000Z"

type Logger struct {
	Out io.Writer
	Err io.Writer
}

func New() Logger {
	return Logger{Out: os.Stdout, Err: os.Stderr}
}

func (l Logger) Info(message string, meta ...any) {
	l.write(l.stdout(), "INFO", message, meta)
}

func (l Logger) Warn(message string, meta ...any) {
	l.write(l.stdout(), "WARN", message, meta)
}

func (l Logger) Error(message string, meta ...any) {
	l.write(l.stderr(), "ERROR", message, meta)
}

func (l Logger) stdout() io.Writer {
	if l.Out == nil {
		return os.Stdout
	}
	return l.Out
}

func (l Logger) stderr() io.Writer {
	if l.Err == nil {
		return os.Stderr
	}
	return l.Err
}

func (l Logger) write(w io.Writer, level string, message string, meta []any) {
	line := fmt.Sprintf("[%s] %s: %s", time.Now().UTC().Format(timestampLayout), level, message)

	if len(meta) > 0 {
		parts := make([]string, 0, len(meta))
		for _, entry := range meta {
			parts = append(parts, format(entry))
		}
		line = line + " " + strings.Join(parts, " ")
	}

	fmt.Fprintln(w, line)
}

func format(value any) string {
	if err, ok := value.(error); ok && err != nil {
		return err.Error()
	}
	return fmt.Sprint(value)
}
