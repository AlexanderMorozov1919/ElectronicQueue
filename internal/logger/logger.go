package logger

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	instance  *logrus.Logger
	once      sync.Once
	logChan   chan logMessage
	syncOnce  sync.Once
	waitGroup sync.WaitGroup
)

// logMessage — единица логирования
type logMessage struct {
	entry *logrus.Entry
	level logrus.Level
	msg   string
}

// Init создает логгер (вызывать в main.go)
func Init(logFile string) {
	once.Do(func() {
		instance = logrus.New()
		instance.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})

		if logFile != "" {
			if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
				instance.Fatalf("Failed to create log directory: %v", err)
			}

			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				instance.Fatalf("Failed to open log file: %v", err)
			}

			instance.AddHook(&fileHook{
				Writer:    file,
				Formatter: &logrus.JSONFormatter{}, // или TextFormatter без цветов
			})

			runtime.SetFinalizer(instance, func(_ interface{}) {
				file.Close()
			})
		}

		// Канал для логирования
		logChan = make(chan logMessage, 1000)

		// Запускаем обработку логов в фоне
		go processLogs()
	})
}

// asyncLogger — асинхронная обёртка над logrus.Entry
type asyncLogger struct {
	entry *logrus.Entry
}

func (l *asyncLogger) Info(msg string) {
	sendLog(logrus.InfoLevel, l.entry, msg)
}
func (l *asyncLogger) Warn(msg string) {
	sendLog(logrus.WarnLevel, l.entry, msg)
}
func (l *asyncLogger) Error(msg string) {
	sendLog(logrus.ErrorLevel, l.entry, msg)
}
func (l *asyncLogger) WithError(err error) *asyncLogger {
	return &asyncLogger{entry: l.entry.WithError(err)}
}
func (l *asyncLogger) Fatal(msg string) {
	sendLog(logrus.FatalLevel, l.entry, msg)
}
func (l *asyncLogger) WithField(k string, v interface{}) *asyncLogger {
	return &asyncLogger{entry: l.entry.WithField(k, v)}
}
func (l *asyncLogger) WithFields(fields logrus.Fields) *asyncLogger {
	return &asyncLogger{entry: l.entry.WithFields(fields)}
}

func sendLog(level logrus.Level, entry *logrus.Entry, msg string) {
	waitGroup.Add(1)
	logChan <- logMessage{level: level, entry: entry, msg: msg}
}

// processLogs читает из канала и пишет лог
func processLogs() {
	for msg := range logChan {
		switch msg.level {
		case logrus.InfoLevel:
			msg.entry.Info(msg.msg)
		case logrus.WarnLevel:
			msg.entry.Warn(msg.msg)
		case logrus.ErrorLevel:
			msg.entry.Error(msg.msg)
		default:
			msg.entry.Print(msg.msg)
		}
		waitGroup.Done()
	}
}

// fileHook для записи в файл
type fileHook struct {
	Writer    io.Writer
	Formatter logrus.Formatter
	mu        sync.Mutex
}

func (h *fileHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	line, err := h.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.Writer.Write(line)
	return err
}

func (h *fileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Default возвращает логгер по умолчанию (асинхронный)
func Default() *asyncLogger {
	return &asyncLogger{entry: instance.WithField("module", "default")}
}

// Sync дожидается окончания логирования и закрывает канал
func Sync() error {
	syncOnce.Do(func() {
		if logChan != nil {
			waitGroup.Wait()
			close(logChan)
		}
	})
	return nil
}
