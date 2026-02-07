package handlers

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AleksaS003/zastitaprojekat/cmd/cli/utils"
	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/lea"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func HandleLEA(args []string) {
	if len(args) < 1 {
		fmt.Println("Expected 'encrypt', 'decrypt', 'genkey', or 'genkey-file' subcommand")
		fmt.Println("Usage: crypto-cli lea <encrypt|decrypt|genkey|genkey-file> [options]")
		os.Exit(1)
	}

	switch args[0] {
	case "encrypt":
		logger.Info("LEA_ENCRYPT", "Starting LEA encryption", true, nil)
		handleLEAEncrypt(args[1:])
	case "decrypt":
		logger.Info("LEA_DECRYPT", "Starting LEA decryption", true, nil)
		handleLEADecrypt(args[1:])
	case "genkey":
		logger.Info("LEA_GENKEY", "Generating LEA key", true, nil)
		handleLEAGenKey(args[1:])
	case "genkey-file":
		logger.Info("LEA_GENKEY_FILE", "Generating LEA key file", true, nil)
		handleLEAGenKeyFile(args[1:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleLEAEncrypt(args []string) {
	cmd := flag.NewFlagSet("lea encrypt", flag.ExitOnError)
	file := cmd.String("file", "", "File to encrypt (required)")
	key := cmd.String("key", "", "Encryption key (hex string)")
	keyFile := cmd.String("keyfile", "", "File containing encryption key")
	output := cmd.String("output", "", "Output file (required)")

	cmd.Parse(args)

	if *file == "" || *output == "" {
		logger.Error("LEA_ENCRYPT", "Missing required arguments", map[string]interface{}{
			"file_provided":   *file != "",
			"output_provided": *output != "",
		})
		log.Fatal("Both --file and --output are required")
	}

	keyBytes, err := utils.LoadKey(*key, *keyFile)
	if err != nil {
		logger.Error("LEA_ENCRYPT", "Failed to load key", map[string]interface{}{
			"key_provided":     *key != "",
			"keyfile_provided": *keyFile != "",
			"error":            err.Error(),
		})
		log.Fatal("Failed to load key:", err)
	}

	cipher, err := lea.NewLEA(keyBytes)
	if err != nil {
		logger.Error("LEA_ENCRYPT", "Failed to create LEA cipher", map[string]interface{}{
			"key_size": len(keyBytes) * 8,
			"error":    err.Error(),
		})
		log.Fatal("Failed to create LEA cipher:", err)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		logger.Error("LEA_ENCRYPT", "Failed to read input file", map[string]interface{}{
			"file":  *file,
			"error": err.Error(),
		})
		log.Fatal("Failed to read input file:", err)
	}

	logger.Info("LEA_ENCRYPT", "Starting encryption process", true, map[string]interface{}{
		"input_file":  *file,
		"output_file": *output,
		"key_size":    len(keyBytes) * 8,
		"input_size":  len(data),
	})

	encrypted, err := cipher.Encrypt(data)
	if err != nil {
		logger.Error("LEA_ENCRYPT", "Encryption failed", map[string]interface{}{
			"input_file": *file,
			"input_size": len(data),
			"error":      err.Error(),
		})
		log.Fatal("Encryption failed:", err)
	}

	err = os.WriteFile(*output, encrypted, 0644)
	if err != nil {
		logger.Error("LEA_ENCRYPT", "Failed to write output file", map[string]interface{}{
			"output_file": *output,
			"error":       err.Error(),
		})
		log.Fatal("Failed to write output file:", err)
	}

	logger.LogEncryption("encrypt", "LEA", *file, int64(len(data)), true, map[string]interface{}{
		"output_file": *output,
		"output_size": len(encrypted),
		"key_size":    len(keyBytes) * 8,
		"overhead":    len(encrypted) - len(data),
	})

	fmt.Printf("✓ File encrypted successfully: %s -> %s\n", *file, *output)
	fmt.Printf("  Original size: %d bytes\n", len(data))
	fmt.Printf("  Encrypted size: %d bytes\n", len(encrypted))
}

func handleLEADecrypt(args []string) {
	cmd := flag.NewFlagSet("lea decrypt", flag.ExitOnError)
	file := cmd.String("file", "", "File to decrypt (required)")
	key := cmd.String("key", "", "Decryption key (hex string)")
	keyFile := cmd.String("keyfile", "", "File containing decryption key")
	output := cmd.String("output", "", "Output file (required)")

	cmd.Parse(args)

	if *file == "" || *output == "" {
		logger.Error("LEA_DECRYPT", "Missing required arguments", map[string]interface{}{
			"file_provided":   *file != "",
			"output_provided": *output != "",
		})
		log.Fatal("Both --file and --output are required")
	}

	keyBytes, err := utils.LoadKey(*key, *keyFile)
	if err != nil {
		logger.Error("LEA_DECRYPT", "Failed to load key", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatal("Failed to load key:", err)
	}

	cipher, err := lea.NewLEA(keyBytes)
	if err != nil {
		logger.Error("LEA_DECRYPT", "Failed to create LEA cipher", map[string]interface{}{
			"key_size": len(keyBytes) * 8,
			"error":    err.Error(),
		})
		log.Fatal("Failed to create LEA cipher:", err)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		logger.Error("LEA_DECRYPT", "Failed to read input file", map[string]interface{}{
			"file":  *file,
			"error": err.Error(),
		})
		log.Fatal("Failed to read input file:", err)
	}

	logger.Info("LEA_DECRYPT", "Starting decryption process", true, map[string]interface{}{
		"input_file":  *file,
		"output_file": *output,
		"key_size":    len(keyBytes) * 8,
		"input_size":  len(data),
	})

	decrypted, err := cipher.Decrypt(data)
	if err != nil {
		logger.Error("LEA_DECRYPT", "Decryption failed", map[string]interface{}{
			"input_file": *file,
			"input_size": len(data),
			"error":      err.Error(),
		})
		log.Fatal("Decryption failed:", err)
	}

	err = os.WriteFile(*output, decrypted, 0644)
	if err != nil {
		logger.Error("LEA_DECRYPT", "Failed to write output file", map[string]interface{}{
			"output_file": *output,
			"error":       err.Error(),
		})
		log.Fatal("Failed to write output file:", err)
	}

	logger.LogEncryption("decrypt", "LEA", *output, int64(len(decrypted)), true, map[string]interface{}{
		"input_file": *file,
		"input_size": len(data),
		"key_size":   len(keyBytes) * 8,
	})

	fmt.Printf("✓ File decrypted successfully: %s -> %s\n", *file, *output)
	fmt.Printf("  Encrypted size: %d bytes\n", len(data))
	fmt.Printf("  Decrypted size: %d bytes\n", len(decrypted))
}

func handleLEAGenKeyFile(args []string) {
	genKeyCmd := flag.NewFlagSet("genkey-file", flag.ExitOnError)
	size := genKeyCmd.Int("size", 256, "Key size: 128, 192, or 256")
	output := genKeyCmd.String("output", "lea.key", "Output file name")

	if len(args) < 1 {
		genKeyCmd.Usage()
		os.Exit(1)
	}

	genKeyCmd.Parse(args)

	var keySize int
	switch *size {
	case 128:
		keySize = 128
	case 192:
		keySize = 192
	case 256:
		keySize = 256
	default:
		logger.Error("LEA_GENKEY_FILE", "Invalid key size", map[string]interface{}{
			"size":        *size,
			"valid_sizes": []int{128, 192, 256},
		})
		log.Fatal("Size must be 128, 192, or 256")
	}

	key, err := lea.GenerateKey(keySize)
	if err != nil {
		logger.Error("LEA_GENKEY_FILE", "Failed to generate key", map[string]interface{}{
			"size":  keySize,
			"error": err.Error(),
		})
		log.Fatal("Failed to generate key:", err)
	}

	err = os.WriteFile(*output, key, 0600)
	if err != nil {
		logger.Error("LEA_GENKEY_FILE", "Failed to write key file", map[string]interface{}{
			"output_file": *output,
			"error":       err.Error(),
		})
		log.Fatal("Failed to write key file:", err)
	}

	logger.Info("LEA_GENKEY_FILE", "LEA key generated and saved", true, map[string]interface{}{
		"key_size":       keySize,
		"output_file":    *output,
		"key_size_bytes": len(key),
	})

	fmt.Printf("✓ Generated LEA %d-bit key and saved to: %s\n", keySize, *output)
	fmt.Printf("  Key size: %d bytes\n", len(key))
}

func handleLEAGenKey(args []string) {
	size := 256
	if len(args) > 0 {
		switch args[0] {
		case "128":
			size = 128
		case "192":
			size = 192
		case "256":
			size = 256
		}
	}

	key, err := lea.GenerateKey(size)
	if err != nil {
		logger.Error("LEA_GENKEY", "Failed to generate key", map[string]interface{}{
			"size":  size,
			"error": err.Error(),
		})
		log.Fatal("Failed to generate key:", err)
	}

	logger.Info("LEA_GENKEY", "LEA key generated", true, map[string]interface{}{
		"key_size":       size,
		"key_size_bytes": len(key),
	})

	fmt.Printf("Generated LEA %d-bit key:\n", size)

	fmt.Print("Hex: ")
	for _, b := range key {
		fmt.Printf("%02x", b)
	}
	fmt.Println()

	fmt.Println("\nRaw bytes (save with: ./crypto-cli lea genkey 128 | tail -c 16 > key.bin):")
	os.Stdout.Write(key)
	fmt.Println()
}
