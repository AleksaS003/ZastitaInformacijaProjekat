package utils

import (
	"log"
	"os"
	"time"

	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func InitLogger() {
	if err := logger.InitGlobal("./logs"); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
	} else {
		workingDir, _ := os.Getwd()
		logger.Info(logger.SERVER_START, "Crypto App started", true, map[string]interface{}{
			"version":     "1.0.0",
			"args":        os.Args,
			"working_dir": workingDir,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	}
}

func CloseLogger() {
	logger.Close()
}

func LogCommand(args []string) {
	if len(args) == 0 {
		return
	}

	activity := logger.ActivityType("CLI_COMMAND")
	if len(args) > 0 {
		activity = logger.ActivityType(args[0])
	}

	logger.Info(activity, "Executing command", true, map[string]interface{}{
		"full_command": joinArgs(args),
		"arg_count":    len(args),
	})
}

func LogError(activity string, message string, details interface{}) {

	activityType := logger.ActivityType(activity)
	logger.Error(activityType, message, details)
}

func joinArgs(args []string) string {
	return ""
}

func init() {

}
