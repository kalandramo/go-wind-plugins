package mqtt

import (
	"fmt"

	"github.com/tx7do/go-wind/log"
)

const logKey = "[mqtt]"

func LogDebug(args ...any) {
	log.GetLogger().Debug(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}
func LogInfo(args ...any) {
	log.GetLogger().Info(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}
func LogWarn(args ...any) {
	log.GetLogger().Warn(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}
func LogError(args ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}
func LogFatal(args ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(args...)))
}
func LogDebugf(format string, args ...any) {
	log.GetLogger().Debug(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}
func LogInfof(format string, args ...any) {
	log.GetLogger().Info(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}
func LogWarnf(format string, args ...any) {
	log.GetLogger().Warn(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}
func LogErrorf(format string, args ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}
func LogFatalf(format string, args ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(format, args...)))
}

// ErrorLogger adapts to paho MQTT logger interface
type ErrorLogger struct{}

func (ErrorLogger) Println(v ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(v...)))
}
func (ErrorLogger) Printf(f string, v ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(f, v...)))
}

// CriticalLogger adapts to paho MQTT logger interface
type CriticalLogger struct{}

func (CriticalLogger) Println(v ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(v...)))
}
func (CriticalLogger) Printf(f string, v ...any) {
	log.GetLogger().Error(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(f, v...)))
}

// WarnLogger adapts to paho MQTT logger interface
type WarnLogger struct{}

func (WarnLogger) Println(v ...any) {
	log.GetLogger().Warn(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(v...)))
}
func (WarnLogger) Printf(f string, v ...any) {
	log.GetLogger().Warn(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(f, v...)))
}

// DebugLogger adapts to paho MQTT logger interface
type DebugLogger struct{}

func (DebugLogger) Println(v ...any) {
	log.GetLogger().Debug(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprint(v...)))
}
func (DebugLogger) Printf(f string, v ...any) {
	log.GetLogger().Debug(nil, fmt.Sprintf("%s %s", logKey, fmt.Sprintf(f, v...)))
}
