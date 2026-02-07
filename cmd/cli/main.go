package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AleksaS003/zastitaprojekat/cmd/cli/handlers"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func main() {

	if err := logger.InitGlobal("./logs"); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
	} else {
		workingDir, _ := os.Getwd()
		logger.Info(logger.SERVER_START, "Crypto App started", true, map[string]interface{}{
			"version":     "1.0.0",
			"args":        os.Args,
			"working_dir": workingDir,
		})
		defer logger.Close()
	}

	if len(os.Args) < 2 {
		handlers.PrintHelp()
		os.Exit(1)
	}

	logger.Info(logger.ActivityType("CLI_COMMAND"),
		fmt.Sprintf("Executing command: %s", os.Args[1]),
		true,
		map[string]interface{}{
			"full_command": strings.Join(os.Args, " "),
			"arg_count":    len(os.Args),
		})

	switch os.Args[1] {

	case "foursquare":
		handlers.HandleFoursquare(os.Args[2:])
	case "lea":
		handlers.HandleLEA(os.Args[2:])
	case "pcbc":
		handlers.HandlePCBC(os.Args[2:])
	case "sha256":
		handlers.HandleSHA256(os.Args[2:])

	case "encrypt-file":
		handlers.HandleEncryptFile(os.Args[2:])
	case "decrypt-file":
		handlers.HandleDecryptFile(os.Args[2:])

	case "fsw":
		handlers.HandleFSW(os.Args[2:])

	case "server":
		handlers.HandleTCPServer(os.Args[2:])
	case "client":
		handlers.HandleTCPClient(os.Args[2:])

	case "logs":
		handlers.HandleLogs(os.Args[2:])

	case "help":
		handlers.PrintHelp()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		logger.Error(logger.ActivityType("CLI_COMMAND"), "Unknown command", map[string]interface{}{
			"command": os.Args[1],
			"valid_commands": []string{
				"foursquare", "lea", "pcbc", "sha256",
				"encrypt-file", "decrypt-file", "help",
				"fsw", "server", "client", "logs",
			},
		})
		handlers.PrintHelp()
		os.Exit(1)
	}
}
