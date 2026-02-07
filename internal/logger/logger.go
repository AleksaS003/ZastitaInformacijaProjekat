package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogLevel string

const (
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
)

type ActivityType string

const (
	FSW_START      ActivityType = "FSW_START"
	FSW_STOP       ActivityType = "FSW_STOP"
	FSW_NEW_FILE   ActivityType = "FSW_NEW_FILE"
	ENCRYPT        ActivityType = "ENCRYPT"
	DECRYPT        ActivityType = "DECRYPT"
	SEND_FILE      ActivityType = "SEND_FILE"
	RECEIVE_FILE   ActivityType = "RECEIVE_FILE"
	VERIFY_HASH    ActivityType = "VERIFY_HASH"
	CLIENT_CONNECT ActivityType = "CLIENT_CONNECT"
	SERVER_START   ActivityType = "SERVER_START"
	SERVER_STOP    ActivityType = "SERVER_STOP"
	FILE_MODIFY    ActivityType = "FILE_MODIFY"
	FILE_DELETE    ActivityType = "FILE_DELETE"
)

type LogEntry struct {
	Timestamp  time.Time    `json:"timestamp"`
	Level      LogLevel     `json:"level"`
	Activity   ActivityType `json:"activity"`
	Message    string       `json:"message"`
	Details    interface{}  `json:"details,omitempty"`
	Success    bool         `json:"success"`
	Source     string       `json:"source,omitempty"`
	RemoteAddr string       `json:"remote_addr,omitempty"`
}

type Logger struct {
	fileLogger   *log.Logger
	console      *log.Logger
	file         *os.File
	jsonFile     *os.File
	mu           sync.RWMutex
	logLevel     LogLevel
	enabled      bool
	logDirectory string
}

var (
	globalLogger *Logger
	once         sync.Once
)

func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFilePath := filepath.Join(logDir, "crypto-app.log")
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	jsonFilePath := filepath.Join(logDir, "activity-log.json")
	jsonFile, err := os.OpenFile(jsonFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to open JSON log file: %w", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)

	return &Logger{
		fileLogger:   log.New(file, "", log.Ldate|log.Ltime),
		console:      log.New(multiWriter, "", log.Ldate|log.Ltime),
		file:         file,
		jsonFile:     jsonFile,
		logLevel:     INFO,
		enabled:      true,
		logDirectory: logDir,
	}, nil
}

func InitGlobal(logDir string) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(logDir)
	})
	return err
}

func GetGlobal() *Logger {
	if globalLogger == nil {

		globalLogger = &Logger{
			console: log.New(os.Stdout, "", log.Ldate|log.Ltime),
			enabled: true,
		}
	}
	return globalLogger
}

func (l *Logger) Log(level LogLevel, activity ActivityType, message string, success bool, details interface{}) {
	if !l.enabled {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Activity:  activity,
		Message:   message,
		Details:   details,
		Success:   success,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	logLine := fmt.Sprintf("[%s] %s: %s - Success: %v",
		level, activity, message, success)
	if details != nil {
		logLine += fmt.Sprintf(" | Details: %v", details)
	}

	l.fileLogger.Println(logLine)
	l.console.Println(logLine)

	jsonData, err := json.Marshal(entry)
	if err == nil {
		l.jsonFile.Write(jsonData)
		l.jsonFile.WriteString("\n")
	}
}

func (l *Logger) Info(activity ActivityType, message string, success bool, details interface{}) {
	l.Log(INFO, activity, message, success, details)
}

func (l *Logger) Warning(activity ActivityType, message string, success bool, details interface{}) {
	l.Log(WARNING, activity, message, success, details)
}

func (l *Logger) Error(activity ActivityType, message string, details interface{}) {
	l.Log(ERROR, activity, message, false, details)
}

func (l *Logger) LogFSWEvent(eventType, filePath, message string, success bool) {
	details := map[string]interface{}{
		"file_path": filePath,
		"event":     eventType,
	}

	var activity ActivityType
	switch eventType {
	case "CREATE":
		activity = FSW_NEW_FILE
	case "MODIFY":
		activity = FILE_MODIFY
	case "DELETE":
		activity = FILE_DELETE
	case "START":
		activity = FSW_START
	case "STOP":
		activity = FSW_STOP
	default:
		activity = FSW_NEW_FILE
	}

	l.Info(activity, message, success, details)
}

func (l *Logger) LogEncryption(operation, algorithm, filePath string, fileSize int64, success bool, details map[string]interface{}) {
	var activity ActivityType
	if operation == "encrypt" {
		activity = ENCRYPT
	} else {
		activity = DECRYPT
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["algorithm"] = algorithm
	details["file_path"] = filePath
	details["file_size"] = fileSize

	message := fmt.Sprintf("%s %s using %s",
		strings.Title(operation), filepath.Base(filePath), algorithm)

	l.Info(activity, message, success, details)
}

func (l *Logger) LogNetwork(activity ActivityType, address, message string, success bool, details interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	if m, ok := details.(map[string]interface{}); ok {
		m["address"] = address
	} else {
		details = map[string]interface{}{
			"address": address,
			"details": details,
		}
	}

	l.Info(activity, message, success, details)
}

func (l *Logger) LogHashVerification(filePath, expectedHash, actualHash string, match bool) {
	details := map[string]interface{}{
		"file_path":     filePath,
		"expected_hash": expectedHash,
		"actual_hash":   actualHash,
		"match":         match,
	}

	message := fmt.Sprintf("Hash verification for %s", filepath.Base(filePath))
	if match {
		message += " - SUCCESS"
	} else {
		message += " - FAILED"
	}

	l.Info(VERIFY_HASH, message, match, details)
}

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}
	if l.jsonFile != nil {
		l.jsonFile.Close()
	}
}

func Info(activity ActivityType, message string, success bool, details interface{}) {
	GetGlobal().Info(activity, message, success, details)
}

func Warning(activity ActivityType, message string, success bool, details interface{}) {
	GetGlobal().Warning(activity, message, success, details)
}

func Error(activity ActivityType, message string, details interface{}) {
	GetGlobal().Error(activity, message, details)
}

func LogFSWEvent(eventType, filePath, message string, success bool) {
	GetGlobal().LogFSWEvent(eventType, filePath, message, success)
}

func LogEncryption(operation, algorithm, filePath string, fileSize int64, success bool, details map[string]interface{}) {
	GetGlobal().LogEncryption(operation, algorithm, filePath, fileSize, success, details)
}

func LogNetwork(activity ActivityType, address, message string, success bool, details interface{}) {
	GetGlobal().LogNetwork(activity, address, message, success, details)
}

func LogHashVerification(filePath, expectedHash, actualHash string, match bool) {
	GetGlobal().LogHashVerification(filePath, expectedHash, actualHash, match)
}

func Close() {
	GetGlobal().Close()
}
