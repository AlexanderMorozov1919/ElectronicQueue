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

// Init создает логгер
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
				Formatter: &logrus.JSONFormatter{},
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

// AsyncLogger — асинхронная обёртка над logrus.Entry
type AsyncLogger struct {
	entry *logrus.Entry
}

func (l *AsyncLogger) Trace(msg string) {
	sendLog(logrus.TraceLevel, l.entry, msg)
}
func (l *AsyncLogger) Debug(msg string) {
	sendLog(logrus.DebugLevel, l.entry, msg)
}
func (l *AsyncLogger) Info(msg string) {
	sendLog(logrus.InfoLevel, l.entry, msg)
}
func (l *AsyncLogger) Warn(msg string) {
	sendLog(logrus.WarnLevel, l.entry, msg)
}
func (l *AsyncLogger) Error(msg string) {
	sendLog(logrus.ErrorLevel, l.entry, msg)
}
func (l *AsyncLogger) Fatal(msg string) {
	sendLog(logrus.FatalLevel, l.entry, msg)
}
func (l *AsyncLogger) Panic(msg string) {
	sendLog(logrus.PanicLevel, l.entry, msg)
}
func (l *AsyncLogger) WithError(err error) *AsyncLogger {
	return &AsyncLogger{entry: l.entry.WithError(err)}
}
func (l *AsyncLogger) WithField(k string, v interface{}) *AsyncLogger {
	return &AsyncLogger{entry: l.entry.WithField(k, v)}
}
func (l *AsyncLogger) WithFields(fields logrus.Fields) *AsyncLogger {
	return &AsyncLogger{entry: l.entry.WithFields(fields)}
}

func sendLog(level logrus.Level, entry *logrus.Entry, msg string) {
	waitGroup.Add(1)
	logChan <- logMessage{level: level, entry: entry, msg: msg}
}

// processLogs читает из канала и пишет лог
func processLogs() {
	for msg := range logChan {
		switch msg.level {
		case logrus.TraceLevel:
			msg.entry.Trace(msg.msg)
		case logrus.DebugLevel:
			msg.entry.Debug(msg.msg)
		case logrus.InfoLevel:
			msg.entry.Info(msg.msg)
		case logrus.WarnLevel:
			msg.entry.Warn(msg.msg)
		case logrus.ErrorLevel:
			msg.entry.Error(msg.msg)
		case logrus.FatalLevel:
			msg.entry.Fatal(msg.msg)
		case logrus.PanicLevel:
			msg.entry.Panic(msg.msg)
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

// Default возвращает логгер по умолчанию
func Default() *AsyncLogger {
	return &AsyncLogger{entry: instance.WithField("module", "default")}
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
