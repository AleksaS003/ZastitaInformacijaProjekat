package handlers

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AleksaS003/zastitaprojekat/cmd/cli/utils"
	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/pcbc"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func HandlePCBC(args []string) {
	if len(args) < 1 {
		fmt.Println("Expected 'encrypt' or 'decrypt' subcommand")
		fmt.Println("Usage: crypto-cli pcbc <encrypt|decrypt> [options]")
		os.Exit(1)
	}

	switch args[0] {
	case "encrypt":
		logger.Info("PCBC_ENCRYPT", "Starting PCBC encryption", true, nil)
		handlePCBCEncrypt(args[1:])
	case "decrypt":
		logger.Info("PCBC_DECRYPT", "Starting PCBC decryption", true, nil)
		handlePCBCDecrypt(args[1:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handlePCBCEncrypt(args []string) {
	cmd := flag.NewFlagSet("pcbc encrypt", flag.ExitOnError)
	file := cmd.String("file", "", "File to encrypt (required)")
	key := cmd.String("key", "", "Encryption key (hex string)")
	keyFile := cmd.String("keyfile", "", "File containing encryption key")
	output := cmd.String("output", "", "Output file (required)")

	cmd.Parse(args)

	if *file == "" || *output == "" {
		logger.Error("PCBC_ENCRYPT", "Missing required arguments", nil)
		log.Fatal("Both --file and --output are required")
	}

	keyBytes, err := utils.LoadKey(*key, *keyFile)
	if err != nil {
		logger.Error("PCBC_ENCRYPT", "Failed to load key", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatal("Failed to load key:", err)
	}

	pcbcCipher, err := pcbc.NewLEAPCBC(keyBytes)
	if err != nil {
		logger.Error("PCBC_ENCRYPT", "Failed to create PCBC cipher", map[string]interface{}{
			"key_size": len(keyBytes) * 8,
			"error":    err.Error(),
		})
		log.Fatal("Failed to create PCBC cipher:", err)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		logger.Error("PCBC_ENCRYPT", "Failed to read input file", map[string]interface{}{
			"file":  *file,
			"error": err.Error(),
		})
		log.Fatal("Failed to read input file:", err)
	}

	logger.Info("PCBC_ENCRYPT", "Starting PCBC encryption", true, map[string]interface{}{
		"input_file":  *file,
		"output_file": *output,
		"key_size":    len(keyBytes) * 8,
		"input_size":  len(data),
		"mode":        "PCBC",
	})

	encrypted, err := pcbcCipher.Encrypt(data)
	if err != nil {
		logger.Error("PCBC_ENCRYPT", "Encryption failed", map[string]interface{}{
			"input_file": *file,
			"input_size": len(data),
			"error":      err.Error(),
		})
		log.Fatal("Encryption failed:", err)
	}

	err = os.WriteFile(*output, encrypted, 0644)
	if err != nil {
		logger.Error("PCBC_ENCRYPT", "Failed to write output file", map[string]interface{}{
			"output_file": *output,
			"error":       err.Error(),
		})
		log.Fatal("Failed to write output file:", err)
	}

	iv := pcbcCipher.GetIV()

	logger.LogEncryption("encrypt", "LEA-PCBC", *file, int64(len(data)), true, map[string]interface{}{
		"output_file": *output,
		"output_size": len(encrypted),
		"key_size":    len(keyBytes) * 8,
		"iv_size":     len(iv),
		"iv_hex":      fmt.Sprintf("%x...", iv[:8]),
		"overhead":    len(encrypted) - len(data),
	})

	fmt.Printf("  File encrypted with PCBC mode\n")
	fmt.Printf("  IV (hex): %x\n", iv)
	fmt.Printf("  Original size: %d bytes\n", len(data))
	fmt.Printf("  Encrypted size: %d bytes (IV + ciphertext)\n", len(encrypted))
	fmt.Printf("  Output saved to: %s\n", *output)
}

func handlePCBCDecrypt(args []string) {
	cmd := flag.NewFlagSet("pcbc decrypt", flag.ExitOnError)
	file := cmd.String("file", "", "File to decrypt (required)")
	key := cmd.String("key", "", "Decryption key (hex string)")
	keyFile := cmd.String("keyfile", "", "File containing decryption key")
	output := cmd.String("output", "", "Output file (required)")

	cmd.Parse(args)

	if *file == "" || *output == "" {
		logger.Error("PCBC_DECRYPT", "Missing required arguments", nil)
		log.Fatal("Both --file and --output are required")
	}

	keyBytes, err := utils.LoadKey(*key, *keyFile)
	if err != nil {
		logger.Error("PCBC_DECRYPT", "Failed to load key", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatal("Failed to load key:", err)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		logger.Error("PCBC_DECRYPT", "Failed to read input file", map[string]interface{}{
			"file":  *file,
			"error": err.Error(),
		})
		log.Fatal("Failed to read input file:", err)
	}

	logger.Info("PCBC_DECRYPT", "Starting PCBC decryption", true, map[string]interface{}{
		"input_file":  *file,
		"output_file": *output,
		"key_size":    len(keyBytes) * 8,
		"input_size":  len(data),
		"mode":        "PCBC",
	})

	pcbcCipher, err := pcbc.NewLEAPCBC(keyBytes)
	if err != nil {
		logger.Error("PCBC_DECRYPT", "Failed to create PCBC cipher", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatal("Failed to create PCBC cipher:", err)
	}

	decrypted, err := pcbcCipher.Decrypt(data)
	if err != nil {
		logger.Error("PCBC_DECRYPT", "Decryption failed", map[string]interface{}{
			"input_file": *file,
			"input_size": len(data),
			"error":      err.Error(),
		})
		log.Fatal("Decryption failed:", err)
	}

	err = os.WriteFile(*output, decrypted, 0644)
	if err != nil {
		logger.Error("PCBC_DECRYPT", "Failed to write output file", map[string]interface{}{
			"output_file": *output,
			"error":       err.Error(),
		})
		log.Fatal("Failed to write output file:", err)
	}

	logger.LogEncryption("decrypt", "LEA-PCBC", *output, int64(len(decrypted)), true, map[string]interface{}{
		"input_file": *file,
		"input_size": len(data),
		"key_size":   len(keyBytes) * 8,
	})

	fmt.Printf("  File decrypted with PCBC mode\n")
	fmt.Printf("  Encrypted size: %d bytes\n", len(data))
	fmt.Printf("  Decrypted size: %d bytes\n", len(decrypted))
	fmt.Printf("  Output saved to: %s\n", *output)
}
