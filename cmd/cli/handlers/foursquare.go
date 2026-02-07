package handlers

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/foursquare"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func HandleFoursquare(args []string) {
	if len(args) < 1 {
		fmt.Println("Expected 'encrypt' or 'decrypt' subcommand")
		os.Exit(1)
	}

	switch args[0] {
	case "encrypt":
		logger.Info("FOURSQUARE_ENCRYPT", "Starting Foursquare encryption", true, nil)
		handleFoursquareEncrypt(args[1:])
	case "decrypt":
		logger.Info("FOURSQUARE_DECRYPT", "Starting Foursquare decryption", true, nil)
		handleFoursquareDecrypt(args[1:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func handleFoursquareEncrypt(args []string) {
	cmd := flag.NewFlagSet("foursquare encrypt", flag.ExitOnError)
	text := cmd.String("text", "", "Text to encrypt")
	file := cmd.String("file", "", "File to encrypt")
	key1 := cmd.String("key1", "keyword", "First key")
	key2 := cmd.String("key2", "example", "Second key")
	output := cmd.String("output", "", "Output file (optional)")

	cmd.Parse(args)

	if *text == "" && *file == "" {
		logger.Error("FOURSQUARE_ENCRYPT", "No input specified", map[string]interface{}{
			"args": args,
		})
		log.Fatal("Either --text or --file must be specified")
	}

	cipher, err := foursquare.NewCipher(*key1, *key2)
	if err != nil {
		logger.Error("FOURSQUARE_ENCRYPT", "Failed to create cipher", map[string]interface{}{
			"key1":  *key1,
			"key2":  *key2,
			"error": err.Error(),
		})
		log.Fatal("Failed to create cipher:", err)
	}

	var inputText string
	var inputSource string
	if *text != "" {
		inputText = *text
		inputSource = "text"
		logger.Info("FOURSQUARE_ENCRYPT", "Using text input", true, map[string]interface{}{
			"text_length": len(inputText),
			"key1":        *key1,
			"key2":        *key2,
		})
	} else {
		content, err := os.ReadFile(*file)
		if err != nil {
			logger.Error("FOURSQUARE_ENCRYPT", "Failed to read file", map[string]interface{}{
				"file":  *file,
				"error": err.Error(),
			})
			log.Fatal("Failed to read file:", err)
		}
		inputText = string(content)
		inputSource = *file
		logger.Info("FOURSQUARE_ENCRYPT", "Using file input", true, map[string]interface{}{
			"file":      *file,
			"file_size": len(content),
			"key1":      *key1,
			"key2":      *key2,
		})
	}

	encrypted, err := cipher.Encrypt(inputText)
	if err != nil {
		logger.Error("FOURSQUARE_ENCRYPT", "Encryption failed", map[string]interface{}{
			"input_source": inputSource,
			"input_length": len(inputText),
			"error":        err.Error(),
		})
		log.Fatal("Encryption failed:", err)
	}

	if *output != "" {
		err := os.WriteFile(*output, []byte(encrypted), 0644)
		if err != nil {
			logger.Error("FOURSQUARE_ENCRYPT", "Failed to write output file", map[string]interface{}{
				"output_file": *output,
				"error":       err.Error(),
			})
			log.Fatal("Failed to write output file:", err)
		}

		logger.Info("FOURSQUARE_ENCRYPT", "Encrypted content written to file", true, map[string]interface{}{
			"input_source": inputSource,
			"output_file":  *output,
			"output_size":  len(encrypted),
			"algorithm":    "Foursquare",
		})

		fmt.Printf("Encrypted content written to: %s\n", *output)
	} else {
		logger.Info("FOURSQUARE_ENCRYPT", "Encryption completed", true, map[string]interface{}{
			"input_source": inputSource,
			"output_size":  len(encrypted),
			"algorithm":    "Foursquare",
		})

		fmt.Println("Encrypted text:")
		fmt.Println(encrypted)
	}
}

func handleFoursquareDecrypt(args []string) {
	cmd := flag.NewFlagSet("foursquare decrypt", flag.ExitOnError)
	text := cmd.String("text", "", "Text to decrypt")
	file := cmd.String("file", "", "File to decrypt")
	key1 := cmd.String("key1", "keyword", "First key")
	key2 := cmd.String("key2", "example", "Second key")
	output := cmd.String("output", "", "Output file (optional)")

	cmd.Parse(args)

	if *text == "" && *file == "" {
		logger.Error("FOURSQUARE_DECRYPT", "No input specified", nil)
		log.Fatal("Either --text or --file must be specified")
	}

	cipher, err := foursquare.NewCipher(*key1, *key2)
	if err != nil {
		logger.Error("FOURSQUARE_DECRYPT", "Failed to create cipher", map[string]interface{}{
			"key1":  *key1,
			"key2":  *key2,
			"error": err.Error(),
		})
		log.Fatal("Failed to create cipher:", err)
	}

	var inputText string
	var inputSource string
	if *text != "" {
		inputText = *text
		inputSource = "text"
		logger.Info("FOURSQUARE_DECRYPT", "Using text input", true, map[string]interface{}{
			"text_length": len(inputText),
			"key1":        *key1,
			"key2":        *key2,
		})
	} else {
		content, err := os.ReadFile(*file)
		if err != nil {
			logger.Error("FOURSQUARE_DECRYPT", "Failed to read file", map[string]interface{}{
				"file":  *file,
				"error": err.Error(),
			})
			log.Fatal("Failed to read file:", err)
		}
		inputText = string(content)
		inputSource = *file
		logger.Info("FOURSQUARE_DECRYPT", "Using file input", true, map[string]interface{}{
			"file":      *file,
			"file_size": len(content),
			"key1":      *key1,
			"key2":      *key2,
		})
	}

	decrypted, err := cipher.Decrypt(inputText)
	if err != nil {
		logger.Error("FOURSQUARE_DECRYPT", "Decryption failed", map[string]interface{}{
			"input_source": inputSource,
			"input_length": len(inputText),
			"error":        err.Error(),
		})
		log.Fatal("Decryption failed:", err)
	}

	if *output != "" {
		err := os.WriteFile(*output, []byte(decrypted), 0644)
		if err != nil {
			logger.Error("FOURSQUARE_DECRYPT", "Failed to write output file", map[string]interface{}{
				"output_file": *output,
				"error":       err.Error(),
			})
			log.Fatal("Failed to write output file:", err)
		}

		logger.Info("FOURSQUARE_DECRYPT", "Decrypted content written to file", true, map[string]interface{}{
			"input_source": inputSource,
			"output_file":  *output,
			"output_size":  len(decrypted),
			"algorithm":    "Foursquare",
		})

		fmt.Printf("Decrypted content written to: %s\n", *output)
	} else {
		logger.Info("FOURSQUARE_DECRYPT", "Decryption completed", true, map[string]interface{}{
			"input_source": inputSource,
			"output_size":  len(decrypted),
			"algorithm":    "Foursquare",
		})

		fmt.Println("Decrypted text:")
		fmt.Println(decrypted)
	}
}
