package handlers

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/AleksaS003/zastitaprojekat/internal/logger"
	"github.com/AleksaS003/zastitaprojekat/internal/network"
)

func HandleTCPServer(args []string) {
	cmd := flag.NewFlagSet("server", flag.ExitOnError)
	address := cmd.String("address", ":8080", "Server address (host:port)")
	outputDir := cmd.String("output", "./received", "Output directory for received files")
	keyfile := cmd.String("keyfile", "", "Decryption key file (required)")

	cmd.Parse(args)

	if *keyfile == "" {
		logger.Error("TCP_SERVER", "Keyfile not specified", nil)
		log.Fatal("--keyfile is required")
	}

	keyBytes, err := os.ReadFile(*keyfile)
	if err != nil {
		logger.Error("TCP_SERVER", "Failed to read key file", map[string]interface{}{
			"keyfile": *keyfile,
			"error":   err.Error(),
		})
		log.Fatal("Failed to read key file:", err)
	}
	keyBytes = bytes.TrimSpace(keyBytes)

	server := network.NewTCPServer(*address, *outputDir, keyBytes)

	logger.LogNetwork(logger.SERVER_START, *address,
		"TCP Server started via CLI", true, map[string]interface{}{
			"output_dir": *outputDir,
			"keyfile":    *keyfile,
			"key_size":   len(keyBytes) * 8,
		})

	fmt.Printf("   Starting TCP Server\n")
	fmt.Printf("   Address: %s\n", *address)
	fmt.Printf("   Output:  %s\n", *outputDir)
	fmt.Printf("   Key:     %s (%d bits)\n", *keyfile, len(keyBytes)*8)
	fmt.Printf("   Logs:    logs/crypto-app.log\n")
	fmt.Printf("   Press Ctrl+C to stop\n")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		if err := server.Start(); err != nil {
			logger.Error("TCP_SERVER", "Failed to start server", map[string]interface{}{
				"address": *address,
				"error":   err.Error(),
			})
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-c
	fmt.Println("\nShutting down server...")

	logger.LogNetwork(logger.SERVER_STOP, *address,
		"TCP Server stopped by user", true, nil)

	server.Stop()
	time.Sleep(1 * time.Second)
}

func HandleTCPClient(args []string) {
	cmd := flag.NewFlagSet("client", flag.ExitOnError)
	address := cmd.String("address", "localhost:8080", "Server address (host:port)")
	file := cmd.String("file", "", "File to send (required)")
	keyfile := cmd.String("keyfile", "", "Encryption key file (required)")
	algorithm := cmd.String("algo", "LEA-PCBC", "Algorithm: LEA, LEA-PCBC")

	cmd.Parse(args)

	if *file == "" || *keyfile == "" {
		logger.Error("TCP_CLIENT", "Missing required arguments", nil)
		log.Fatal("Both --file and --keyfile are required")
	}

	keyBytes, err := os.ReadFile(*keyfile)
	if err != nil {
		logger.Error("TCP_CLIENT", "Failed to read key file", map[string]interface{}{
			"keyfile": *keyfile,
			"error":   err.Error(),
		})
		log.Fatal("Failed to read key file:", err)
	}
	keyBytes = bytes.TrimSpace(keyBytes)

	fileInfo, err := os.Stat(*file)
	if err != nil {
		logger.Error("TCP_CLIENT", "Failed to get file info", map[string]interface{}{
			"file":  *file,
			"error": err.Error(),
		})
		log.Fatal("Failed to get file info:", err)
	}

	logger.LogNetwork(logger.CLIENT_CONNECT, *address,
		"TCP Client started via CLI", true, map[string]interface{}{
			"file":      *file,
			"algorithm": *algorithm,
			"keyfile":   *keyfile,
			"file_size": fileInfo.Size(),
			"key_size":  len(keyBytes) * 8,
		})

	client := network.NewTCPClient(*address, 10*time.Second)

	fmt.Printf("Connecting to server: %s\n", *address)
	if err := client.Connect(); err != nil {
		logger.Error("TCP_CLIENT", "Failed to connect to server", map[string]interface{}{
			"address": *address,
			"error":   err.Error(),
		})
		log.Fatal("Failed to connect:", err)
	}
	defer client.Disconnect()

	fmt.Printf("Sending file: %s (%d bytes)\n", *file, fileInfo.Size())
	fmt.Printf("Algorithm: %s\n", *algorithm)
	fmt.Printf("Key: %s (%d bits)\n", *keyfile, len(keyBytes)*8)

	if err := client.SendFile(*file, *algorithm, keyBytes); err != nil {
		logger.Error("TCP_CLIENT", "Failed to send file", map[string]interface{}{
			"file":      *file,
			"address":   *address,
			"algorithm": *algorithm,
			"error":     err.Error(),
		})
		log.Fatal("Failed to send file:", err)
	}

	logger.Info("TCP_CLIENT", "File sent successfully", true, map[string]interface{}{
		"file":      *file,
		"address":   *address,
		"algorithm": *algorithm,
		"file_size": fileInfo.Size(),
	})

	fmt.Println("File sent successfully!")
}
