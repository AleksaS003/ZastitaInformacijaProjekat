package fsw

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/AleksaS003/zastitaprojekat/internal/core"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
	"github.com/fsnotify/fsnotify"
)

type FileEvent struct {
	Type      string
	Path      string
	Timestamp time.Time
	Success   bool
	Message   string
}

type FileSystemWatcher struct {
	watcher       *fsnotify.Watcher
	watchDir      string
	outputDir     string
	algorithm     string
	key           []byte
	active        bool
	eventChan     chan FileEvent
	stopChan      chan struct{}
	logger        *log.Logger
	fileProcessor *core.FileProcessor
	logFile       *os.File
}

func NewFileSystemWatcher(watchDir, outputDir, algorithm string, key []byte) (*FileSystemWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(logger.FSW_START, "Failed to create watcher", err)
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := os.MkdirAll(watchDir, 0755); err != nil {
		logger.Error(logger.FSW_START, "Failed to create watch directory", err)
		return nil, fmt.Errorf("failed to create watch directory: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error(logger.FSW_START, "Failed to create output directory", err)
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	logFile, err := os.OpenFile("fsw.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error(logger.FSW_START, "Failed to create log file", err)
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	fileLogger := log.New(logFile, "FSW: ", log.Ldate|log.Ltime|log.Lshortfile)

	fsw := &FileSystemWatcher{
		watcher:       watcher,
		watchDir:      watchDir,
		outputDir:     outputDir,
		algorithm:     algorithm,
		key:           key,
		active:        false,
		eventChan:     make(chan FileEvent, 100),
		stopChan:      make(chan struct{}),
		logger:        fileLogger,
		fileProcessor: core.NewFileProcessor(),
		logFile:       logFile,
	}

	return fsw, nil
}

func (f *FileSystemWatcher) Start() error {
	if f.active {
		return fmt.Errorf("watcher is already active")
	}

	err := f.watcher.Add(f.watchDir)
	if err != nil {
		logger.Error(logger.FSW_START, "Failed to add watch directory", err)
		return fmt.Errorf("failed to add watch directory: %w", err)
	}

	f.active = true

	logger.Info(logger.FSW_START, "File System Watcher started", true, map[string]interface{}{
		"watch_dir":  f.watchDir,
		"output_dir": f.outputDir,
		"algorithm":  f.algorithm,
	})

	f.logger.Printf("FSW started watching: %s", f.watchDir)
	f.logger.Printf("Output directory: %s", f.outputDir)
	f.logger.Printf("Algorithm: %s", f.algorithm)

	go f.eventLoop()

	return nil
}

func (f *FileSystemWatcher) Stop() error {
	if !f.active {
		return fmt.Errorf("watcher is not active")
	}

	close(f.stopChan)
	err := f.watcher.Close()
	f.active = false

	logger.Info(logger.FSW_STOP, "File System Watcher stopped", true, map[string]interface{}{
		"watch_dir": f.watchDir,
	})

	f.logger.Println("FSW stopped")

	if f.logFile != nil {
		f.logFile.Close()
	}

	return err
}

func (f *FileSystemWatcher) eventLoop() {
	logger.Info(logger.FSW_START, "FSW event loop started", true, nil)

	for {
		select {
		case event, ok := <-f.watcher.Events:
			if !ok {
				return
			}
			f.handleEvent(event)

		case err, ok := <-f.watcher.Errors:
			if !ok {
				return
			}
			logger.Error(logger.FSW_NEW_FILE, "FSW error occurred", map[string]interface{}{
				"error": err.Error(),
				"dir":   f.watchDir,
			})
			f.logger.Printf("FSW error: %v", err)
			f.eventChan <- FileEvent{
				Type:      "ERROR",
				Path:      f.watchDir,
				Timestamp: time.Now(),
				Success:   false,
				Message:   fmt.Sprintf("FSW error: %v", err),
			}

		case <-f.stopChan:
			logger.Info(logger.FSW_STOP, "FSW event loop stopped", true, nil)
			return
		}
	}
}

func (f *FileSystemWatcher) handleEvent(event fsnotify.Event) {
	fileEvent := FileEvent{
		Path:      event.Name,
		Timestamp: time.Now(),
	}

	info, err := os.Stat(event.Name)
	if err != nil || info.IsDir() {
		return
	}

	if filepath.Ext(event.Name) == ".enc" {
		return
	}

	fileName := filepath.Base(event.Name)
	fileSize := info.Size()

	switch event.Op {
	case fsnotify.Create:
		fileEvent.Type = "CREATE"
		fileEvent.Message = "New file detected"

		logger.LogFSWEvent("CREATE", event.Name,
			fmt.Sprintf("New file detected: %s (%d bytes)", fileName, fileSize), true)

		f.logger.Printf("New file detected: %s (%d bytes)", event.Name, fileSize)

		time.Sleep(100 * time.Millisecond)

		success, msg := f.autoEncrypt(event.Name)
		fileEvent.Success = success
		fileEvent.Message = msg

	case fsnotify.Write:
		fileEvent.Type = "MODIFY"
		fileEvent.Message = "File modified"

		logger.LogFSWEvent("MODIFY", event.Name,
			fmt.Sprintf("File modified: %s", fileName), true)

		f.logger.Printf("File modified: %s", event.Name)

	case fsnotify.Remove:
		fileEvent.Type = "DELETE"
		fileEvent.Message = "File deleted"

		logger.LogFSWEvent("DELETE", event.Name,
			fmt.Sprintf("File deleted: %s", fileName), true)

		f.logger.Printf("File deleted: %s", event.Name)

	case fsnotify.Rename:
		fileEvent.Type = "RENAME"
		fileEvent.Message = "File renamed"

		logger.LogFSWEvent("RENAME", event.Name,
			fmt.Sprintf("File renamed: %s", fileName), true)

		f.logger.Printf("File renamed: %s", event.Name)
	}

	select {
	case f.eventChan <- fileEvent:

	default:

	}
}

func (f *FileSystemWatcher) autoEncrypt(filePath string) (bool, string) {

	fileName := filepath.Base(filePath)
	outputPath := filepath.Join(f.outputDir, fileName+".enc")

	if _, err := os.Stat(outputPath); err == nil {
		logger.Warning(logger.ENCRYPT, "File already encrypted", false, map[string]interface{}{
			"file_path":   filePath,
			"output_path": outputPath,
		})
		return false, "File already encrypted"
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to get file info", map[string]interface{}{
			"file_path": filePath,
			"error":     err.Error(),
		})
		return false, fmt.Sprintf("Failed to read file: %v", err)
	}

	err = f.fileProcessor.EncryptFileWithMetadata(filePath, outputPath, f.algorithm, f.key)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Auto-encryption failed", map[string]interface{}{
			"file_path":   filePath,
			"output_path": outputPath,
			"algorithm":   f.algorithm,
			"error":       err.Error(),
		})
		f.logger.Printf("Auto-encryption failed for %s: %v", filePath, err)
		return false, fmt.Sprintf("Encryption failed: %v", err)
	}

	logger.LogEncryption("encrypt", f.algorithm, filePath, fileInfo.Size(), true, map[string]interface{}{
		"output_path": outputPath,
		"watch_dir":   f.watchDir,
	})

	f.logger.Printf("Auto-encryption successful: %s -> %s", filePath, outputPath)
	return true, fmt.Sprintf("File encrypted: %s", outputPath)
}

func (f *FileSystemWatcher) GetEventChannel() <-chan FileEvent {
	return f.eventChan
}

func (f *FileSystemWatcher) IsActive() bool {
	return f.active
}

func (f *FileSystemWatcher) GetWatchDir() string {
	return f.watchDir
}

func (f *FileSystemWatcher) GetOutputDir() string {
	return f.outputDir
}

func (f *FileSystemWatcher) EncryptExistingFiles() ([]string, error) {
	logger.Info(logger.ENCRYPT, "Encrypting existing files in directory", true, map[string]interface{}{
		"directory": f.watchDir,
		"algorithm": f.algorithm,
	})

	files, err := f.fileProcessor.ProcessDirectory(f.watchDir, f.outputDir, f.algorithm, f.key, "encrypt")
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to encrypt existing files", map[string]interface{}{
			"directory": f.watchDir,
			"error":     err.Error(),
		})
		return files, err
	}

	logger.Info(logger.ENCRYPT, "Existing files encrypted successfully", true, map[string]interface{}{
		"directory":  f.watchDir,
		"file_count": len(files),
		"output_dir": f.outputDir,
	})

	return files, nil
}
