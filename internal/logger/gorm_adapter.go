package logger

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

// GORMLogger реализует интерфейс gorm.Logger
type GORMLogger struct {
	*logrus.Entry
	SlowThreshold time.Duration
	LogLevel      logger.LogLevel
}

func NewGORMLogger() *GORMLogger {
	return &GORMLogger{
		Entry:         instance.WithField("module", "gorm"),
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
	}
}

func (l *GORMLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Entry.Infof(msg, data...)
	}
}

func (l *GORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Entry.Warnf(msg, data...)
	}
}

func (l *GORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Entry.Errorf(msg, data...)
	}
}

func (l *GORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel == logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := logrus.Fields{
		"elapsed": elapsed.String(),
		"rows":    rows,
	}
	if err != nil {
		l.Entry.WithError(err).WithFields(fields).Errorf("SQL error: %s", sql)
	} else if elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn {
		l.Entry.WithFields(fields).Warnf("SLOW SQL: %s", sql)
	} else if l.LogLevel >= logger.Info {
		l.Entry.WithFields(fields).Infof("SQL: %s", sql)
	}
}
