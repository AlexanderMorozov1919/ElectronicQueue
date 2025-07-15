package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GORMLogger реализует интерфейс gorm.Logger
type GORMLogger struct {
	SlowThreshold time.Duration
	LogLevel      gormlogger.LogLevel
}

func NewGORMLogger() *GORMLogger {
	return &GORMLogger{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      gormlogger.Info,
	}
}

func (l *GORMLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		Default().WithField("module", "GORM").Info(fmt.Sprintf(msg, data...))
	}
}

func (l *GORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		Default().WithField("module", "GORM").Warn(fmt.Sprintf(msg, data...))
	}
}

func (l *GORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		Default().WithField("module", "GORM").Error(fmt.Sprintf(msg, data...))
	}
}

func (l *GORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	log := Default().WithFields(map[string]interface{}{
		"module":  "GORM",
		"sql":     sql,
		"rows":    rows,
		"elapsed": elapsed,
	})

	// Ошибки, не связанные с 'record not found', считаются серьезными
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("GORM Error")
		return
	}

	// Медленные запросы
	if l.SlowThreshold > 0 && elapsed > l.SlowThreshold {
		log.Warn("Slow SQL")
		return
	}

	// Обычные запросы
	if l.LogLevel >= gormlogger.Info {
		log.Info("SQL Query")
	}
}
