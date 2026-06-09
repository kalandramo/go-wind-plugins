package temporal

import (
	"fmt"
	"log/slog"
)

const (
	logKey = "[Temporal]"
)

func LogDebug(args ...any) {
	slog.Debug(fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}

func LogInfo(args ...any) {
	slog.Info(fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}

func LogWarn(args ...any) {
	slog.Warn(fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}

func LogError(args ...any) {
	slog.Error(fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}

func LogFatal(args ...any) {
	slog.Error(fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}

func LogDebugf(format string, args ...any) {
	slog.Debug(fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}

func LogInfof(format string, args ...any) {
	slog.Info(fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}

func LogWarnf(format string, args ...any) {
	slog.Warn(fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}

func LogErrorf(format string, args ...any) {
	slog.Error(fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}

func LogFatalf(format string, args ...any) {
	slog.Error(fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}
