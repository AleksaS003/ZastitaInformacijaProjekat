package handlers

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/AleksaS003/zastitaprojekat/internal/fsw"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func HandleFSW(args []string) {
	if len(args) < 1 {
		fmt.Println("Expected 'start', 'stop', or 'encrypt-existing' subcommand")
		fmt.Println("Usage: crypto-cli fsw <start|stop|encrypt-existing> [options]")
		os.Exit(1)
	}

	switch args[0] {
	case "start":
		logger.Info(logger.FSW_START, "Starting File System Watcher via CLI", true, nil)
		handleFSWStart(args[1:])
	case "stop":
		logger.Info(logger.FSW_STOP, "Stopping File System Watcher via CLI", true, nil)
		handleFSWStop(args[1:])
	case "encrypt-existing":

		logger.Info(logger.ActivityType("FSW_ENCRYPT_EXISTING"), "Encrypting existing files via CLI", true, nil)
		handleFSWEncryptExisting(args[1:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleFSWStart(args []string) {
	cmd := flag.NewFlagSet("fsw start", flag.ExitOnError)
	watchDir := cmd.String("watch", "./watch", "Directory to watch")
	outputDir := cmd.String("output", "./encrypted", "Output directory for encrypted files")
	keyfile := cmd.String("keyfile", "", "Encryption key file (required)")
	algorithm := cmd.String("algo", "LEA-PCBC", "Algorithm: LEA, LEA-PCBC")

	cmd.Parse(args)

	if *keyfile == "" {
		logger.Error(logger.FSW_START, "Keyfile not specified", nil)
		log.Fatal("--keyfile is required")
	}

	keyBytes, err := os.ReadFile(*keyfile)
	if err != nil {
		logger.Error(logger.FSW_START, "Failed to read key file", map[string]interface{}{
			"keyfile": *keyfile,
			"error":   err.Error(),
		})
		log.Fatal("Failed to read key file:", err)
	}
	keyBytes = bytes.TrimSpace(keyBytes)

	watcher, err := fsw.NewFileSystemWatcher(*watchDir, *outputDir, *algorithm, keyBytes)
	if err != nil {
		logger.Error(logger.FSW_START, "Failed to create FSW", map[string]interface{}{
			"watch_dir":  *watchDir,
			"output_dir": *outputDir,
			"algorithm":  *algorithm,
			"error":      err.Error(),
		})
		log.Fatal("Failed to create FSW:", err)
	}

	err = watcher.Start()
	if err != nil {
		logger.Error(logger.FSW_START, "Failed to start FSW", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatal("Failed to start FSW:", err)
	}

	logger.Info(logger.FSW_START, "File System Watcher started via CLI", true, map[string]interface{}{
		"watch_dir":  *watchDir,
		"output_dir": *outputDir,
		"algorithm":  *algorithm,
		"keyfile":    *keyfile,
		"key_size":   len(keyBytes) * 8,
	})

	fmt.Printf("   File System Watcher STARTED\n")
	fmt.Printf("   Watching: %s\n", *watchDir)
	fmt.Printf("   Output:   %s\n", *outputDir)
	fmt.Printf("   Algorithm: %s\n", *algorithm)
	fmt.Printf("   Log file: fsw.log\n")
	fmt.Printf("   Activity log: logs/crypto-app.log\n")
	fmt.Println("\nPress Ctrl+C to stop...")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	eventChan := watcher.GetEventChannel()

	for {
		select {
		case event := <-eventChan:
			timestamp := event.Timestamp.Format("15:04:05")
			status := "good"
			if !event.Success {
				status = "error"
			}
			fmt.Printf("[%s] %s %s: %s\n", timestamp, status, event.Type, event.Message)

		case <-c:
			fmt.Println("\nReceived interrupt signal, stopping FSW...")
			watcher.Stop()
			logger.Info(logger.FSW_STOP, "FSW stopped by user interrupt", true, map[string]interface{}{
				"watch_dir": *watchDir,
			})
			return
		}
	}
}

func handleFSWStop(args []string) {
	logger.Info(logger.FSW_STOP, "FSW stop command invoked", true, nil)

	fmt.Println("Stopping FSW...")
	fmt.Println("Note: In CLI mode, press Ctrl+C to stop")
	fmt.Println("In GUI mode, use the stop button")
}

func handleFSWEncryptExisting(args []string) {
	cmd := flag.NewFlagSet("fsw encrypt-existing", flag.ExitOnError)
	watchDir := cmd.String("watch", "./watch", "Directory with existing files")
	outputDir := cmd.String("output", "./encrypted", "Output directory")
	keyfile := cmd.String("keyfile", "", "Encryption key file (required)")
	algorithm := cmd.String("algo", "LEA-PCBC", "Algorithm: LEA, LEA-PCBC")

	cmd.Parse(args)

	if *keyfile == "" {

		logger.Error(logger.ActivityType("FSW_ENCRYPT_EXISTING"), "Keyfile not specified", nil)
		log.Fatal("--keyfile is required")
	}

	keyBytes, err := os.ReadFile(*keyfile)
	if err != nil {

		logger.Error(logger.ActivityType("FSW_ENCRYPT_EXISTING"), "Failed to read key file", map[string]interface{}{
			"keyfile": *keyfile,
			"error":   err.Error(),
		})
		log.Fatal("Failed to read key file:", err)
	}
	keyBytes = bytes.TrimSpace(keyBytes)

	watcher, err := fsw.NewFileSystemWatcher(*watchDir, *outputDir, *algorithm, keyBytes)
	if err != nil {

		logger.Error(logger.ActivityType("FSW_ENCRYPT_EXISTING"), "Failed to create FSW", map[string]interface{}{
			"watch_dir":  *watchDir,
			"output_dir": *outputDir,
			"algorithm":  *algorithm,
			"error":      err.Error(),
		})
		log.Fatal("Failed to create FSW:", err)
	}

	files, err := watcher.EncryptExistingFiles()
	if err != nil {

		logger.Error(logger.ActivityType("FSW_ENCRYPT_EXISTING"), "Failed to encrypt existing files", map[string]interface{}{
			"watch_dir": *watchDir,
			"error":     err.Error(),
		})
		log.Fatal("Failed to encrypt existing files:", err)
	}

	logger.Info(logger.ActivityType("FSW_ENCRYPT_EXISTING"), "Existing files encrypted", true, map[string]interface{}{
		"directory":  *watchDir,
		"file_count": len(files),
		"output_dir": *outputDir,
		"algorithm":  *algorithm,
	})

	fmt.Printf("Encrypted %d existing files\n", len(files))
	for _, file := range files {
		fmt.Printf("   - %s\n", filepath.Base(file))
	}
}
