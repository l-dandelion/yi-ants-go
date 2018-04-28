package log

import (
	"github.com/l-dandelion/yi-ants-go/lib/library/log/base"
	"os"
)

var defaultLogger = Logger(
	base.TYPE_LOGRUS,
	base.LEVEL_INFO,
	base.FORMAT_TEXT,
	os.Stdout,
)

func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

func Infoln(v ...interface{}) {
	defaultLogger.Infoln(v...)
}

func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

func Warnln(v ...interface{}) {
	defaultLogger.Warnln(v...)
}

func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

func Errorln(v ...interface{}) {
	defaultLogger.Errorln(v...)
}

func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

func Fatalln(v ...interface{}) {
	defaultLogger.Fatalln(v...)
}

func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

func Panic(v ...interface{}) {
	defaultLogger.Panic(v...)
}

func Panicln(v ...interface{}) {
	defaultLogger.Panicln(v...)
}

func Panicf(format string, v ...interface{}) {
	defaultLogger.Panicf(format, v...)
}
