package handlers

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func HandleLogs(args []string) {
	if len(args) < 1 {
		printLogsHelp()
		return
	}

	switch args[0] {
	case "show":
		handleShowLogs(args[1:])
	case "clear":
		handleClearLogs(args[1:])
	case "stats":
		handleLogStats(args[1:])
	default:
		printLogsHelp()
	}
}

func handleShowLogs(args []string) {
	cmd := flag.NewFlagSet("logs show", flag.ExitOnError)
	lines := cmd.Int("n", 50, "Number of lines to show")
	follow := cmd.Bool("f", false, "Follow log output")
	filter := cmd.String("filter", "", "Filter by activity type")

	if len(args) > 0 {
		cmd.Parse(args)
	}

	logFile := "./logs/crypto-app.log"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Println("No log file found.")
		return
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		return
	}

	linesArr := strings.Split(string(content), "\n")
	start := len(linesArr) - *lines - 1
	if start < 0 {
		start = 0
	}

	fmt.Printf("=== Showing last %d lines of log ===\n", *lines)
	for i := start; i < len(linesArr); i++ {
		line := linesArr[i]
		if *filter != "" && !strings.Contains(line, *filter) {
			continue
		}
		fmt.Println(line)
	}

	if *follow {
		fmt.Println("\n=== Following log (Ctrl+C to stop) ===")
		fmt.Println("Follow mode requires fsnotify implementation")
	}
}

func handleClearLogs(args []string) {
	cmd := flag.NewFlagSet("logs clear", flag.ExitOnError)
	confirm := cmd.Bool("yes", false, "Confirm deletion")

	if len(args) > 0 {
		cmd.Parse(args)
	}

	if !*confirm {
		fmt.Println("WARNING: This will delete all log files!")
		fmt.Println("Run with --yes flag to confirm")
		return
	}

	files := []string{
		"./logs/crypto-app.log",
		"./logs/activity-log.json",
		"./fsw.log",
	}

	deleted := 0
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			err := os.Remove(file)
			if err != nil {
				fmt.Printf("Error deleting %s: %v\n", file, err)
			} else {
				fmt.Printf("Deleted: %s\n", file)
				deleted++
			}
		}
	}

	fmt.Printf("\nDeleted %d log files.\n", deleted)

	os.MkdirAll("./logs", 0755)
	os.WriteFile("./logs/crypto-app.log", []byte(""), 0644)
	os.WriteFile("./logs/activity-log.json", []byte(""), 0644)

	logger.Info("LOGS_CLEARED", "Log files cleared by user", true, map[string]interface{}{
		"deleted_count": deleted,
	})
}

func handleLogStats(args []string) {
	logFile := "./logs/activity-log.json"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Println("No JSON log file found.")
		return
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	activityCount := make(map[string]int)
	successCount := 0
	errorCount := 0
	var firstTimestamp, lastTimestamp string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if i == 0 {
			if ts, ok := entry["timestamp"].(string); ok {
				firstTimestamp = ts
			}
		}
		if ts, ok := entry["timestamp"].(string); ok {
			lastTimestamp = ts
		}

		if activity, ok := entry["activity"].(string); ok {
			activityCount[activity]++
		}

		if success, ok := entry["success"].(bool); ok {
			if success {
				successCount++
			} else {
				errorCount++
			}
		}
	}

	fmt.Println("=== Log Statistics ===")
	fmt.Printf("Total entries: %d\n", len(lines)-1)
	fmt.Printf("Successful operations: %d\n", successCount)
	fmt.Printf("Failed operations: %d\n", errorCount)

	totalOps := successCount + errorCount
	if totalOps > 0 {
		successRate := float64(successCount) / float64(totalOps) * 100
		fmt.Printf("Success rate: %.1f%%\n", successRate)
	}

	if firstTimestamp != "" && lastTimestamp != "" {
		first, _ := time.Parse(time.RFC3339, firstTimestamp)
		last, _ := time.Parse(time.RFC3339, lastTimestamp)
		duration := last.Sub(first)
		fmt.Printf("Time span: %s to %s (%.1f hours)\n",
			first.Format("2006-01-02 15:04"),
			last.Format("2006-01-02 15:04"),
			duration.Hours())
	}

	fmt.Println("\n=== Activity Breakdown ===")
	for activity, count := range activityCount {
		fmt.Printf("  %-25s: %d\n", activity, count)
	}
}

func printLogsHelp() {
	fmt.Println(`Log Management Commands:
  logs show        - Show recent logs
    -n <lines>     - Number of lines to show (default: 50)
    -f             - Follow log output
    --filter <str> - Filter by activity type
  
  logs clear       - Clear all log files
    --yes          - Confirm deletion
  
  logs stats       - Show log statistics`)
}
