// Package logging provides caller-annotated logrus logging.
package logging

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"runtime"
)

func identifyCaller(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "<unknown>"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "<unknown>"
	}

	return fn.Name()
}

func Debug(args ...interface{}) {
	log.Debug(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprint(args...)))
}

func Debugf(format string, args ...interface{}) {
	log.Debug(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprintf(format, args...)))
}

func Info(args ...interface{}) {
	log.Info(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprint(args...)))
}

func Infof(format string, args ...interface{}) {
	log.Info(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprintf(format, args...)))
}

func Warn(args ...interface{}) {
	log.Warn(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprint(args...)))
}

func Warnf(format string, args ...interface{}) {
	log.Warn(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprintf(format, args...)))
}

func Error(args ...interface{}) {
	log.Error(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprint(args...)))
}

func Errorf(format string, args ...interface{}) {
	log.Error(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprintf(format, args...)))
}

func Fatal(args ...interface{}) {
	log.Fatal(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprint(args...)))
}

func Fatalf(format string, args ...interface{}) {
	log.Fatal(fmt.Sprintf("[%s] %s", identifyCaller(2), fmt.Sprintf(format, args...)))
}
