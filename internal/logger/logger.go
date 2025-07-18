package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	instance  *logrus.Logger
	once      sync.Once
	logChan   chan logMessage
	syncOnce  sync.Once
	waitGroup sync.WaitGroup
)

const (
	// Цвета для консоли
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorWhite  = "\033[97m"

	ColorDarkCyan = "\033[38;5;38m"

	// Фоны
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgPurple  = "\033[45m"
	BgCyan    = "\033[46m"
	BgDarkRed = "\033[101m"
	BgFatal   = "\033[41;1m"
)

// logMessage — единица логирования
type logMessage struct {
	entry *logrus.Entry
	level logrus.Level
	msg   string
}

// Init создает и настраивает логгер
func Init(logDir string) {
	once.Do(func() {
		instance = logrus.New()
		instance.SetOutput(io.Discard)

		instance.AddHook(&consoleHook{
			Formatter: &CustomFormatter{DisableColors: false},
		})

		if logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				logrus.Fatalf("Failed to create log directory: %v", err)
			}
			fileHook, err := newMultiFileHook(logDir)
			if err != nil {
				logrus.Fatalf("Failed to create file hook: %v", err)
			}
			instance.AddHook(fileHook)
		}

		logChan = make(chan logMessage, 1000)
		go processLogs()
	})
}

// CustomFormatter реализует кастомное форматирование логов
type CustomFormatter struct {
	DisableColors bool
}

// shortenLevel сокращает имя уровня логирования до 4 символов
func (f *CustomFormatter) shortenLevel(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel:
		return "TRAC"
	case logrus.DebugLevel:
		return "DEBU"
	case logrus.InfoLevel:
		return "INFO"
	case logrus.WarnLevel:
		return "WARN"
	case logrus.ErrorLevel:
		return "ERRO"
	case logrus.FatalLevel:
		return "FATA"
	case logrus.PanicLevel:
		return "PANI"
	default:
		return "UNKN"
	}
}

// Format форматирует запись лога
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	levelColor := f.levelColor(entry.Level)
	levelText := f.shortenLevel(entry.Level)
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	module := "DEFAULT"
	if m, ok := entry.Data["module"].(string); ok {
		module = strings.ToUpper(m)
	}

	// Компактный заголовок: [LEVEL TIMESTAMP MODULE]
	if !f.DisableColors {
		// Весь заголовок окрашивается в цвет уровня
		fmt.Fprintf(b, "%s[%s %s %s]%s ", levelColor, levelText, timestamp, module, ColorReset)
	} else {
		fmt.Fprintf(b, "[%s %s %s] ", levelText, timestamp, module)
	}

	// Дополнительная информация
	switch module {
	case "GIN":
		f.formatGin(b, entry)
	case "GORM":
		f.formatGorm(b, entry)
	default:
		f.formatDefault(b, entry)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// formatGin форматирует лог от Gin
func (f *CustomFormatter) formatGin(b *bytes.Buffer, entry *logrus.Entry) {
	method := entry.Data["method"].(string)
	status := entry.Data["status"].(int)
	path := entry.Data["path"].(string)
	latency := entry.Data["latency"].(time.Duration)
	ip := entry.Data["ip"].(string)

	_, methodBg := f.methodColors(method)
	_, statusBg := f.statusColors(status)

	// Формат: PATH | METHOD | STATUS | IP | MESSAGE/ERROR | LATENCY
	if !f.DisableColors {
		// Текст метода и статуса всегда белый, меняется только фон
		fmt.Fprintf(b, "%s%s%s | %s%s %s %s | %s%s %d %s | %s |",
			ColorPurple, path, ColorReset,
			methodBg, ColorWhite, method, ColorReset,
			statusBg, ColorWhite, status, ColorReset,
			ip,
		)
	} else {
		fmt.Fprintf(b, "%s | %s | %d | %s |",
			path, method, status, ip)
	}

	// Message или Error
	if msg := entry.Message; msg != "" && msg != "Request handled" {
		fmt.Fprintf(b, " %s |", msg)
	}
	if err, ok := entry.Data["error"]; ok {
		errMsg := err.(error).Error()
		if !f.DisableColors {
			fmt.Fprintf(b, " %s%s%s |", ColorRed, errMsg, ColorReset)
		} else {
			fmt.Fprintf(b, " %s |", errMsg)
		}
	}

	// Latency
	fmt.Fprintf(b, " %v", latency)
}

// formatGorm форматирует лог от GORM
func (f *CustomFormatter) formatGorm(b *bytes.Buffer, entry *logrus.Entry) {
	sql := entry.Data["sql"].(string)
	rows := entry.Data["rows"].(int64)
	latency := entry.Data["elapsed"].(time.Duration)

	// Формат: SQL QUERY | ROWS | MESSAGE/ERROR | LATENCY
	if !f.DisableColors {
		// SQL-запрос белым текстом
		fmt.Fprintf(b, "%s%s%s | %d rows |", ColorWhite, sql, ColorReset, rows)
	} else {
		fmt.Fprintf(b, "%s | %d rows |", sql, rows)
	}

	// Message или Error
	if err, ok := entry.Data["error"]; ok {
		errMsg := err.(error).Error()
		if !f.DisableColors {
			fmt.Fprintf(b, " %s%s%s |", ColorRed, errMsg, ColorReset)
		} else {
			fmt.Fprintf(b, " %s |", errMsg)
		}
	} else if msg := entry.Message; msg != "" && msg != "SQL Query" {
		fmt.Fprintf(b, " %s |", msg)
	}

	// Latency
	fmt.Fprintf(b, " %v", latency)
}

// formatDefault форматирует лог по умолчанию
func (f *CustomFormatter) formatDefault(b *bytes.Buffer, entry *logrus.Entry) {
	// Формат: CALLER | MESSAGE | ERROR
	if caller, ok := entry.Data["caller"]; ok {
		if !f.DisableColors {
			fmt.Fprintf(b, "%s%s%s | ", ColorPurple, caller, ColorReset)
		} else {
			fmt.Fprintf(b, "%s | ", caller)
		}
	}

	b.WriteString(entry.Message)

	if err, ok := entry.Data["error"]; ok {
		errMsg := err.(error).Error()
		if !f.DisableColors {
			fmt.Fprintf(b, " | %s%s%s", ColorRed, errMsg, ColorReset)
		} else {
			fmt.Fprintf(b, " | %s", errMsg)
		}
	}
}

func (f *CustomFormatter) levelColor(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel:
		return ColorGray
	case logrus.DebugLevel:
		return ColorGreen
	case logrus.InfoLevel:
		return ColorBlue
	case logrus.WarnLevel:
		return ColorYellow
	case logrus.ErrorLevel:
		return ColorRed
	case logrus.FatalLevel, logrus.PanicLevel:
		return BgFatal + ColorWhite
	default:
		return ColorReset
	}
}

func (f *CustomFormatter) statusColors(status int) (string, string) {
	var bg string
	switch {
	case status >= 100 && status < 200:
		bg = BgBlue
	case status >= 200 && status < 300:
		bg = BgGreen
	case status >= 300 && status < 400:
		bg = BgYellow
	case status >= 400 && status < 500:
		bg = BgRed
	default:
		bg = BgDarkRed
	}
	return ColorWhite, bg
}

func (f *CustomFormatter) methodColors(method string) (string, string) {
	var bg string
	switch method {
	case "GET":
		bg = BgBlue
	case "POST":
		bg = BgGreen
	case "PUT":
		bg = BgYellow
	case "PATCH":
		bg = BgCyan
	case "DELETE":
		bg = BgRed
	default:
		bg = BgPurple
	}
	return ColorWhite, bg
}

// AsyncLogger — асинхронная обёртка над logrus.Entry
type AsyncLogger struct {
	entry *logrus.Entry
}

func (l *AsyncLogger) Trace(msg string) {
	l.sendLog(logrus.TraceLevel, msg)
}
func (l *AsyncLogger) Debug(msg string) {
	l.sendLog(logrus.DebugLevel, msg)
}
func (l *AsyncLogger) Info(msg string) {
	l.sendLog(logrus.InfoLevel, msg)
}
func (l *AsyncLogger) Warn(msg string) {
	l.sendLog(logrus.WarnLevel, msg)
}
func (l *AsyncLogger) Error(msg string) {
	l.sendLog(logrus.ErrorLevel, msg)
}
func (l *AsyncLogger) Fatal(msg string) {
	l.sendLog(logrus.FatalLevel, msg)
	Sync()
	os.Exit(1)
}
func (l *AsyncLogger) Panic(msg string) {
	l.sendLog(logrus.PanicLevel, msg)
	Sync()
	panic(msg)
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

func (l *AsyncLogger) sendLog(level logrus.Level, msg string) {
	entry := l.entry
	module, _ := entry.Data["module"].(string)

	// Добавляем информацию о вызывающем коде только для логов по умолчанию
	if module == "" || strings.ToUpper(module) == "DEFAULT" {
		if pc, file, line, ok := runtime.Caller(2); ok {
			funcName := runtime.FuncForPC(pc).Name()
			fileName := strings.TrimSuffix(filepath.Base(file), ".go") // Получаем имя файла без расширения
			shortFuncName := funcName[strings.LastIndex(funcName, ".")+1:]
			caller := fmt.Sprintf("%s/%s:%d", fileName, shortFuncName, line)
			entry = entry.WithField("caller", caller)
		}
	}

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

// consoleHook для вывода в консоль
type consoleHook struct {
	Formatter logrus.Formatter
}

func (h *consoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *consoleHook) Fire(entry *logrus.Entry) error {
	line, err := h.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(line)
	return err
}

// multiFileHook для записи в несколько файлов
type multiFileHook struct {
	formatter logrus.Formatter
	writers   map[logrus.Level]io.Writer
	allFile   io.Writer
	mu        sync.Mutex
}

func newMultiFileHook(logDir string) (*multiFileHook, error) {
	writers := make(map[logrus.Level]io.Writer)

	allFilePath := filepath.Join(logDir, "all.log")
	allFile, err := os.OpenFile(allFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open all.log: %w", err)
	}

	levels := []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
	for _, level := range levels {
		path := filepath.Join(logDir, fmt.Sprintf("%s.log", level.String()))
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file for level %s: %w", level, err)
		}
		writers[level] = file
	}

	runtime.SetFinalizer(instance, func(_ interface{}) {
		allFile.Close()
		for _, w := range writers {
			if f, ok := w.(*os.File); ok {
				f.Close()
			}
		}
	})

	return &multiFileHook{
		formatter: &CustomFormatter{DisableColors: true},
		writers:   writers,
		allFile:   allFile,
	}, nil
}

func (h *multiFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *multiFileHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	line, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	if _, err := h.allFile.Write(line); err != nil {
		return err
	}

	if writer, ok := h.writers[entry.Level]; ok {
		if _, err := writer.Write(line); err != nil {
			return err
		}
	}

	return nil
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
