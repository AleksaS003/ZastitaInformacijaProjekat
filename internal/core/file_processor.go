package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/lea"
	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/pcbc"
	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/sha256"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

type FileProcessor struct {
}

func NewFileProcessor() *FileProcessor {
	return &FileProcessor{}
}

func (fp *FileProcessor) EncryptFileWithMetadata(
	inputPath string,
	outputPath string,
	algorithm string,
	key []byte,
) error {

	originalFileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	logger.Info(logger.ENCRYPT, "Starting file encryption", true, map[string]interface{}{
		"input_file":  inputPath,
		"output_file": outputPath,
		"algorithm":   algorithm,
		"file_size":   originalFileInfo.Size(),
		"key_size":    len(key) * 8,
	})

	data, err := os.ReadFile(inputPath)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to read input file", map[string]interface{}{
			"file_path": inputPath,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to read input file: %w", err)
	}

	fileHash := sha256.HashBytes(data)
	hashStr := sha256.HashToString(fileHash)

	logger.Info(logger.VERIFY_HASH, "File hash calculated", true, map[string]interface{}{
		"file":       inputPath,
		"hash":       hashStr,
		"hash_bytes": len(fileHash),
	})

	var encryptedData []byte
	var iv []byte

	switch algorithm {
	case "LEA":
		logger.Info(logger.ENCRYPT, "Using LEA algorithm", true, map[string]interface{}{
			"key_size": len(key) * 8,
		})

		cipher, err := lea.NewLEA(key)
		if err != nil {
			logger.Error(logger.ENCRYPT, "Failed to create LEA cipher", map[string]interface{}{
				"key_size": len(key) * 8,
				"error":    err.Error(),
			})
			return fmt.Errorf("failed to create LEA cipher: %w", err)
		}
		encryptedData, err = cipher.Encrypt(data)
		if err != nil {
			logger.Error(logger.ENCRYPT, "LEA encryption failed", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("encryption failed: %w", err)
		}

	case "LEA-PCBC":
		logger.Info(logger.ENCRYPT, "Using LEA-PCBC algorithm", true, map[string]interface{}{
			"key_size": len(key) * 8,
			"mode":     "PCBC",
		})

		pcbcCipher, err := pcbc.NewLEAPCBC(key)
		if err != nil {
			logger.Error(logger.ENCRYPT, "Failed to create PCBC cipher", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("failed to create PCBC cipher: %w", err)
		}
		encryptedData, err = pcbcCipher.Encrypt(data)
		if err != nil {
			logger.Error(logger.ENCRYPT, "PCBC encryption failed", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("PCBC encryption failed: %w", err)
		}
		iv = pcbcCipher.GetIV()

		logger.Info(logger.ENCRYPT, "PCBC IV generated", true, map[string]interface{}{
			"iv_size": len(iv),
			"iv_hex":  fmt.Sprintf("%x...", iv[:8]),
		})

	default:
		logger.Error(logger.ENCRYPT, "Unsupported algorithm", map[string]interface{}{
			"algorithm": algorithm,
			"supported": []string{"LEA", "LEA-PCBC"},
		})
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	metadata, err := NewMetadata(
		inputPath,
		algorithm,
		"SHA-256",
		hashStr,
		iv,
	)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to create metadata", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	finalData, err := metadata.AddToEncryptedFile(nil, encryptedData)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to add metadata header", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to add metadata header: %w", err)
	}

	err = os.WriteFile(outputPath, finalData, 0644)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to write output file", map[string]interface{}{
			"output_path": outputPath,
			"error":       err.Error(),
		})
		return fmt.Errorf("failed to write output file: %w", err)
	}

	outputFileInfo, _ := os.Stat(outputPath)

	logger.LogEncryption("encrypt", algorithm, inputPath, originalFileInfo.Size(), true, map[string]interface{}{
		"output_file":    outputPath,
		"original_size":  originalFileInfo.Size(),
		"encrypted_size": outputFileInfo.Size(),
		"overhead":       outputFileInfo.Size() - originalFileInfo.Size(),
		"metadata_size":  len(finalData) - len(encryptedData),
		"hash":           hashStr[:16] + "...",
		"hash_algorithm": "SHA-256",
		"iv_used":        iv != nil,
	})

	return nil
}

func (fp *FileProcessor) DecryptFileWithMetadata(
	inputPath string,
	outputPath string,
	key []byte,
) (*Metadata, error) {

	inputFileInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	logger.Info(logger.DECRYPT, "Starting file decryption", true, map[string]interface{}{
		"input_file":  inputPath,
		"output_file": outputPath,
		"file_size":   inputFileInfo.Size(),
		"key_size":    len(key) * 8,
	})

	data, err := os.ReadFile(inputPath)
	if err != nil {
		logger.Error(logger.DECRYPT, "Failed to read input file", map[string]interface{}{
			"file_path": inputPath,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	metadata, encryptedData, err := ExtractFromEncryptedFile(data)
	if err != nil {
		logger.Error(logger.DECRYPT, "Failed to extract metadata", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	logger.Info(logger.DECRYPT, "Metadata extracted", true, map[string]interface{}{
		"algorithm":      metadata.EncryptionAlgorithm,
		"original_file":  metadata.Filename,
		"hash_algorithm": metadata.HashAlgorithm,
		"hash_present":   metadata.Hash != "",
		"iv_present":     metadata.IV != "",
		"encrypted_size": len(encryptedData),
		"metadata_size":  len(data) - len(encryptedData),
	})

	var decryptedData []byte

	switch metadata.EncryptionAlgorithm {
	case "LEA":
		logger.Info(logger.DECRYPT, "Using LEA algorithm for decryption", true, nil)

		cipher, err := lea.NewLEA(key)
		if err != nil {
			logger.Error(logger.DECRYPT, "Failed to create LEA cipher", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to create LEA cipher: %w", err)
		}
		decryptedData, err = cipher.Decrypt(encryptedData)
		if err != nil {
			logger.Error(logger.DECRYPT, "LEA decryption failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("decryption failed: %w", err)
		}

	case "LEA-PCBC":
		logger.Info(logger.DECRYPT, "Using LEA-PCBC algorithm for decryption", true, map[string]interface{}{
			"iv_present": metadata.IV != "",
		})

		var iv []byte
		if metadata.IV != "" {

			iv = make([]byte, len(metadata.IV)/2)
			for i := 0; i < len(iv); i++ {
				fmt.Sscanf(metadata.IV[i*2:i*2+2], "%02x", &iv[i])
			}

			logger.Info(logger.DECRYPT, "IV extracted from metadata", true, map[string]interface{}{
				"iv_size": len(iv),
				"iv_hex":  fmt.Sprintf("%x...", iv[:8]),
			})
		}

		pcbcCipher, err := pcbc.NewLEAPCBCWithIV(key, iv)
		if err != nil {
			logger.Error(logger.DECRYPT, "Failed to create PCBC cipher", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to create PCBC cipher: %w", err)
		}
		decryptedData, err = pcbcCipher.Decrypt(encryptedData)
		if err != nil {
			logger.Error(logger.DECRYPT, "PCBC decryption failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("PCBC decryption failed: %w", err)
		}

	default:
		logger.Error(logger.DECRYPT, "Unsupported algorithm in metadata", map[string]interface{}{
			"algorithm": metadata.EncryptionAlgorithm,
		})
		return nil, fmt.Errorf("unsupported algorithm: %s", metadata.EncryptionAlgorithm)
	}

	if metadata.Hash != "" {
		logger.Info(logger.VERIFY_HASH, "Starting hash verification", true, map[string]interface{}{
			"expected_hash": metadata.Hash[:16] + "...",
			"algorithm":     metadata.HashAlgorithm,
		})

		decryptedHash := sha256.HashBytes(decryptedData)
		decryptedHashStr := sha256.HashToString(decryptedHash)

		logger.LogHashVerification(outputPath, metadata.Hash, decryptedHashStr, decryptedHashStr == metadata.Hash)

		if decryptedHashStr != metadata.Hash {
			logger.Error(logger.VERIFY_HASH, "Hash verification failed", map[string]interface{}{
				"file":          outputPath,
				"expected_hash": metadata.Hash,
				"actual_hash":   decryptedHashStr,
			})
			return metadata, fmt.Errorf("hash verification failed: file may be corrupted")
		}

		logger.Info(logger.VERIFY_HASH, "Hash verification successful", true, map[string]interface{}{
			"file": outputPath,
		})
	} else {
		logger.Warning(logger.VERIFY_HASH, "No hash in metadata for verification", true, map[string]interface{}{
			"file": outputPath,
		})
	}

	err = os.WriteFile(outputPath, decryptedData, 0644)
	if err != nil {
		logger.Error(logger.DECRYPT, "Failed to write output file", map[string]interface{}{
			"output_path": outputPath,
			"error":       err.Error(),
		})
		return metadata, fmt.Errorf("failed to write output file: %w", err)
	}

	outputFileInfo, _ := os.Stat(outputPath)

	logger.LogEncryption("decrypt", metadata.EncryptionAlgorithm, outputPath,
		outputFileInfo.Size(), true, map[string]interface{}{
			"input_file":     inputPath,
			"original_file":  metadata.Filename,
			"hash_verified":  metadata.Hash != "",
			"hash_algorithm": metadata.HashAlgorithm,
			"iv_used":        metadata.IV != "",
		})

	return metadata, nil
}

func (fp *FileProcessor) ProcessDirectory(
	dirPath string,
	outputDir string,
	algorithm string,
	key []byte,
	action string,
) ([]string, error) {
	logger.Info(logger.ENCRYPT, "Processing directory", true, map[string]interface{}{
		"directory":  dirPath,
		"output_dir": outputDir,
		"algorithm":  algorithm,
		"action":     action,
	})

	var processedFiles []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to read directory", map[string]interface{}{
			"directory": dirPath,
			"error":     err.Error(),
		})
		return nil, err
	}

	logger.Info(logger.ENCRYPT, "Found files in directory", true, map[string]interface{}{
		"directory":  dirPath,
		"file_count": len(entries),
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		inputPath := filepath.Join(dirPath, entry.Name())
		outputPath := filepath.Join(outputDir, entry.Name())

		if action == "encrypt" {
			outputPath = outputPath + ".enc"
			err = fp.EncryptFileWithMetadata(inputPath, outputPath, algorithm, key)
		} else {

			if filepath.Ext(outputPath) == ".enc" {
				outputPath = outputPath[:len(outputPath)-4]
			}
			_, err = fp.DecryptFileWithMetadata(inputPath, outputPath, key)
		}

		if err != nil {
			logger.Error(logger.ENCRYPT, "Failed to process file", map[string]interface{}{
				"file":   inputPath,
				"action": action,
				"error":  err.Error(),
			})
			return processedFiles, fmt.Errorf("failed to process %s: %w", inputPath, err)
		}

		processedFiles = append(processedFiles, inputPath)
		logger.Info(logger.ENCRYPT, "File processed successfully", true, map[string]interface{}{
			"file":   inputPath,
			"action": action,
		})
	}

	logger.Info(logger.ENCRYPT, "Directory processing completed", true, map[string]interface{}{
		"directory":       dirPath,
		"action":          action,
		"files_processed": len(processedFiles),
		"output_dir":      outputDir,
	})

	return processedFiles, nil
}
