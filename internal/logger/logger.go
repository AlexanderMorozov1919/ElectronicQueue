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
	instance *logrus.Logger
	once     sync.Once
)

// Init создает логгер (вызывать в main.go)
func Init(logFile string) {
	once.Do(func() {
		instance = logrus.New()
		instance.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})

		// Настройка вывода в файл (без цветов)
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

			// Добавляем закрытие файла при завершении
			runtime.SetFinalizer(instance, func(_ interface{}) {
				file.Close()
			})
		}
	})
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
func Default() *logrus.Entry {
	return instance.WithField("module", "default")
}

func Sync() error {
	if instance == nil {
		return nil
	}

	// Безопасная проверка хуков
	if infoHooks, exist := instance.Hooks[logrus.InfoLevel]; exist && len(infoHooks) > 0 {
		if hook, ok := infoHooks[0].(*fileHook); ok {
			if file, ok := hook.Writer.(*os.File); ok {
				return file.Sync()
			}
		}
	}
	return nil
}
