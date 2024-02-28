package logger

import (
	"context"
	"io"
	"os"
	"proxy/pkg/config"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
}

type writeHook struct {
	Writers   []io.Writer
	LogLevels []logrus.Level
}

func (hook *writeHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	for _, w := range hook.Writers {
		_, err := w.Write([]byte(line))
		if err != nil {
			return err
		}
	}
	return nil
}

func (hook *writeHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func NewLogger(ctx context.Context, cfg config.Logger) Logger {
	l := logrus.New()

	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	lumber := &lumberjack.Logger{
		Filename: "." + "/server.log",
		MaxSize:  30,
		MaxAge:   2,
		Compress: true,
	}

	switch cfg.Level {
	case "Debug":
		l.SetLevel(logrus.DebugLevel)
	case "Warn":
		l.SetLevel(logrus.WarnLevel)
	case "Info":
		l.SetLevel(logrus.InfoLevel)
	case "Trace":
		l.SetLevel(logrus.TraceLevel)
	default:
		l.SetLevel(logrus.DebugLevel)
	}

	l.SetOutput(io.Discard)
	l.AddHook(&writeHook{
		Writers:   []io.Writer{lumber, os.Stdout},
		LogLevels: logrus.AllLevels,
	})

	return l
}
