package handlers

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AleksaS003/zastitaprojekat/internal/core"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func HandleEncryptFile(args []string) {
	cmd := flag.NewFlagSet("encrypt-file", flag.ExitOnError)
	file := cmd.String("file", "", "File to encrypt (required)")
	keyfile := cmd.String("keyfile", "", "Encryption key file (required)")
	algorithm := cmd.String("algo", "LEA-PCBC", "Algorithm: LEA, LEA-PCBC")
	output := cmd.String("output", "", "Output file (optional)")

	cmd.Parse(args)

	if *file == "" || *keyfile == "" {
		logger.Error(logger.ActivityType("ENCRYPT_FILE"), "Missing required arguments", nil)
		log.Fatal("Both --file and --keyfile are required")
	}

	keyBytes, err := os.ReadFile(*keyfile)
	if err != nil {
		logger.Error(logger.ActivityType("ENCRYPT_FILE"), "Failed to read key file", map[string]interface{}{
			"keyfile": *keyfile,
			"error":   err.Error(),
		})
		log.Fatal("Failed to read key file:", err)
	}
	keyBytes = bytes.TrimSpace(keyBytes)

	outputFile := *output
	if outputFile == "" {
		outputFile = *file + ".enc"
	}

	logger.Info(logger.ActivityType("ENCRYPT_FILE"), "Starting file encryption with metadata", true, map[string]interface{}{
		"input_file":  *file,
		"output_file": outputFile,
		"algorithm":   *algorithm,
		"key_size":    len(keyBytes) * 8,
		"keyfile":     *keyfile,
	})

	processor := core.NewFileProcessor()

	err = processor.EncryptFileWithMetadata(*file, outputFile, *algorithm, keyBytes)
	if err != nil {
		logger.Error(logger.ActivityType("ENCRYPT_FILE"), "Encryption failed", map[string]interface{}{
			"input_file": *file,
			"algorithm":  *algorithm,
			"error":      err.Error(),
		})
		log.Fatal("Encryption failed:", err)
	}

	fmt.Printf("✓ File encrypted with metadata header\n")
	fmt.Printf("  Input:  %s\n", *file)
	fmt.Printf("  Output: %s\n", outputFile)
	fmt.Printf("  Algorithm: %s\n", *algorithm)

	data, _ := os.ReadFile(outputFile)
	if len(data) >= 4 {
		metadataLen := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		if len(data) >= int(4+metadataLen) {
			var metadata map[string]interface{}
			json.Unmarshal(data[4:4+metadataLen], &metadata)
			fmt.Printf("  Metadata size: %d bytes\n", metadataLen)
			if filename, ok := metadata["filename"].(string); ok {
				fmt.Printf("  Original filename: %s\n", filename)
			}
		}
	}
}

func HandleDecryptFile(args []string) {
	cmd := flag.NewFlagSet("decrypt-file", flag.ExitOnError)
	file := cmd.String("file", "", "File to decrypt (required)")
	keyfile := cmd.String("keyfile", "", "Decryption key file (required)")
	output := cmd.String("output", "", "Output file (optional)")

	cmd.Parse(args)

	if *file == "" || *keyfile == "" {
		logger.Error(logger.ActivityType("DECRYPT_FILE"), "Missing required arguments", nil)
		log.Fatal("Both --file and --keyfile are required")
	}

	keyBytes, err := os.ReadFile(*keyfile)
	if err != nil {
		logger.Error(logger.ActivityType("DECRYPT_FILE"), "Failed to read key file", map[string]interface{}{
			"keyfile": *keyfile,
			"error":   err.Error(),
		})
		log.Fatal("Failed to read key file:", err)
	}
	keyBytes = bytes.TrimSpace(keyBytes)

	outputFile := *output
	if outputFile == "" {
		if strings.HasSuffix(*file, ".enc") {
			outputFile = (*file)[:len(*file)-4]
		} else {
			outputFile = *file + ".dec"
		}
	}

	logger.Info(logger.ActivityType("DECRYPT_FILE"), "Starting file decryption with metadata", true, map[string]interface{}{
		"input_file":  *file,
		"output_file": outputFile,
		"key_size":    len(keyBytes) * 8,
		"keyfile":     *keyfile,
	})

	processor := core.NewFileProcessor()

	metadata, err := processor.DecryptFileWithMetadata(*file, outputFile, keyBytes)
	if err != nil {
		logger.Error(logger.ActivityType("DECRYPT_FILE"), "Decryption failed", map[string]interface{}{
			"input_file": *file,
			"error":      err.Error(),
		})
		log.Fatal("Decryption failed:", err)
	}

	logger.Info(logger.ActivityType("DECRYPT_FILE"), "File decrypted successfully", true, map[string]interface{}{
		"algorithm":         metadata.EncryptionAlgorithm,
		"original_filename": metadata.Filename,
		"hash_algorithm":    metadata.HashAlgorithm,
		"hash_verified":     metadata.Hash != "",
		"iv_used":           metadata.IV != "",
	})

	fmt.Printf("  File decrypted successfully\n")
	fmt.Printf("  Input:  %s\n", *file)
	fmt.Printf("  Output: %s\n", outputFile)
	fmt.Printf("  Algorithm: %s\n", metadata.EncryptionAlgorithm)
	fmt.Printf("  Original filename: %s\n", metadata.Filename)
	fmt.Printf("  File size: %d bytes\n", metadata.Size)
	fmt.Printf("  Hash verified: %s\n", metadata.HashAlgorithm)

	if metadata.IV != "" {
		fmt.Printf("  IV used: %s...\n", metadata.IV[:16])
	}
}
