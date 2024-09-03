package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"runtime"
	"strings"
)

type ColorHandler struct {
	slog.Handler
	logger *log.Logger
}

func (handler *ColorHandler) Handle(_ context.Context, record slog.Record) error {
	var level string
	if record.Level == slog.LevelWarn || record.Level == slog.LevelError {
		level = "\033[1;31m" + record.Level.String() + "\033[0m"
	} else {
		level = "\033[1;33m" + record.Level.String() + "\033[0m"
	}
	message := "\033[1m" + record.Message + "\033[0m"

	// Taken from sources of log/slog/record.go, since the method to obtain the frame source is private.
	frames := runtime.CallersFrames([]uintptr{record.PC})
	frame, _ := frames.Next()

	source := "\n  \033[36msource\033[0m \033[2m" + fmt.Sprintf("%s:%d", frame.Function, frame.Line) + "\033[0m"

	var fields []string
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, fmt.Sprintf("\n  \033[32m%s\033[0m \033[2m%s\033[0m", attr.Key, strings.TrimSpace(attr.Value.String())))
		return true
	})
	formattedFields := strings.Join(fields, "")

	handler.logger.Println(fmt.Sprintf("%s %s %s%s", level, message, source, formattedFields))
	return nil
}

func NewColorHandler(out io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	return &ColorHandler{Handler: slog.NewTextHandler(out, opts), logger: log.New(out, "", 0)}
}

type FileHandler struct {
	slog.Handler
	logger *log.Logger
}

func (handler *FileHandler) Handle(_ context.Context, record slog.Record) error {
	timestamp := record.Time.Format("2006-01-02T15:04:05")
	frames := runtime.CallersFrames([]uintptr{record.PC})
	frame, _ := frames.Next()
	var fields []string
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, fmt.Sprintf("%s=%s", strings.ReplaceAll(attr.Key, " ", "-"), strings.TrimSpace(attr.Value.String())))
		return true
	})
	handler.logger.Println(fmt.Sprintf("%s %s %s:%d %s %s", timestamp, record.Level.String(), frame.Function, frame.Line, record.Message, strings.Join(fields, " ")))
	return nil
}

func NewFileHandler(out io.Writer, opts *slog.HandlerOptions) *FileHandler {
	return &FileHandler{Handler: slog.NewTextHandler(out, opts), logger: log.New(out, "", 0)}
}
