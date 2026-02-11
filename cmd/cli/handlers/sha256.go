package handlers

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AleksaS003/zastitaprojekat/cmd/cli/utils"
	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/sha256"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func HandleSHA256(args []string) {
	if len(args) < 1 {
		fmt.Println("Expected 'hash' or 'verify' subcommand")
		fmt.Println("Usage: crypto-cli sha256 <hash|verify> [options]")
		os.Exit(1)
	}

	switch args[0] {
	case "hash":
		logger.Info("SHA256_HASH", "Starting SHA256 hash", true, nil)
		handleSHA256Hash(args[1:])
	case "verify":
		logger.Info("SHA256_VERIFY", "Starting SHA256 verification", true, nil)
		handleSHA256Verify(args[1:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleSHA256Hash(args []string) {
	cmd := flag.NewFlagSet("sha256 hash", flag.ExitOnError)
	text := cmd.String("text", "", "Text to hash")
	file := cmd.String("file", "", "File to hash")
	output := cmd.String("output", "", "Output file for hash (optional)")

	cmd.Parse(args)

	hasText := false
	hasFile := false

	for i := 0; i < len(args); i++ {
		if args[i] == "--text" {
			hasText = true
			break
		}
		if strings.HasPrefix(args[i], "--text=") {
			hasText = true
			break
		}
	}

	for i := 0; i < len(args); i++ {
		if args[i] == "--file" {
			hasFile = true
			break
		}
		if strings.HasPrefix(args[i], "--file=") {
			hasFile = true
			break
		}
	}

	if !hasText && !hasFile {
		logger.Error("SHA256_HASH", "No input specified", nil)
		log.Fatal("Either --text or --file must be specified")
	}

	var hash [32]byte
	var err error
	var source string
	var sourceDetails map[string]interface{}

	if hasText {
		hash = sha256.HashString(*text)
		source = "text"
		sourceDetails = map[string]interface{}{
			"text_length":  len(*text),
			"text_preview": utils.GetTextPreview(*text, 50),
		}
		logger.Info("SHA256_HASH", "Hashing text input", true, sourceDetails)
		fmt.Printf("Text: \"%s\"\n", utils.GetTextPreview(*text, 100))
	} else {
		hash, err = sha256.HashFile(*file)
		if err != nil {
			logger.Error("SHA256_HASH", "Failed to hash file", map[string]interface{}{
				"file":  *file,
				"error": err.Error(),
			})
			log.Fatal("Failed to hash file:", err)
		}
		source = *file
		fileInfo, _ := os.Stat(*file)
		sourceDetails = map[string]interface{}{
			"file":      *file,
			"file_size": fileInfo.Size(),
		}
		logger.Info("SHA256_HASH", "Hashing file input", true, sourceDetails)
		fmt.Printf("File: %s\n", *file)
	}

	hashStr := sha256.HashToString(hash)

	if *output != "" {
		err := os.WriteFile(*output, []byte(hashStr), 0644)
		if err != nil {
			logger.Error("SHA256_HASH", "Failed to write hash file", map[string]interface{}{
				"output_file": *output,
				"error":       err.Error(),
			})
			log.Fatal("Failed to write hash file:", err)
		}

		logger.LogHashVerification(source, "", hashStr, true)

		fmt.Printf("  Hash saved to: %s\n", *output)
	} else {
		logger.LogHashVerification(source, "", hashStr, true)
		fmt.Printf("SHA-256 hash: %s\n", hashStr)
	}
}

func handleSHA256Verify(args []string) {
	cmd := flag.NewFlagSet("sha256 verify", flag.ExitOnError)
	file := cmd.String("file", "", "File to verify")
	hash := cmd.String("hash", "", "Expected hash (hex string)")
	hashFile := cmd.String("hashfile", "", "File containing expected hash")

	cmd.Parse(args)

	if *file == "" {
		logger.Error("SHA256_VERIFY", "No file specified", nil)
		log.Fatal("--file is required")
	}

	if *hash == "" && *hashFile == "" {
		logger.Error("SHA256_VERIFY", "No hash specified", nil)
		log.Fatal("Either --hash or --hashfile must be specified")
	}

	fileHash, err := sha256.HashFile(*file)
	if err != nil {
		logger.Error("SHA256_VERIFY", "Failed to hash file", map[string]interface{}{
			"file":  *file,
			"error": err.Error(),
		})
		log.Fatal("Failed to hash file:", err)
	}

	var expectedHashStr string
	var hashSource string
	if *hash != "" {
		expectedHashStr = *hash
		hashSource = "command_line"
	} else {
		data, err := os.ReadFile(*hashFile)
		if err != nil {
			logger.Error("SHA256_VERIFY", "Failed to read hash file", map[string]interface{}{
				"hash_file": *hashFile,
				"error":     err.Error(),
			})
			log.Fatal("Failed to read hash file:", err)
		}
		expectedHashStr = string(bytes.TrimSpace(data))
		hashSource = *hashFile
	}

	if len(expectedHashStr) != 64 {
		logger.Error("SHA256_VERIFY", "Invalid hash length", map[string]interface{}{
			"expected_length": 64,
			"actual_length":   len(expectedHashStr),
			"hash_preview":    expectedHashStr[:utils.Min(16, len(expectedHashStr))] + "...",
		})
		log.Fatal("Hash must be 64 hex characters (256 bits)")
	}

	fileHashStr := sha256.HashToString(fileHash)

	logger.Info("SHA256_VERIFY", "Starting hash verification", true, map[string]interface{}{
		"file":                  *file,
		"hash_source":           hashSource,
		"expected_hash_preview": expectedHashStr[:16] + "...",
		"actual_hash_preview":   fileHashStr[:16] + "...",
	})

	fmt.Printf("File: %s\n", *file)
	fmt.Printf("Calculated hash: %s\n", fileHashStr)
	fmt.Printf("Expected hash:   %s\n", expectedHashStr)

	match := fileHashStr == expectedHashStr

	logger.LogHashVerification(*file, expectedHashStr, fileHashStr, match)

	if match {
		fmt.Println("Hashes match - file integrity verified")
	} else {
		fmt.Println("Hashes DO NOT match - file may be corrupted")
	}
}
