package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

func LoadKey(keyStr, keyFile string) ([]byte, error) {
	var keyBytes []byte
	var err error

	if keyStr != "" {
		keyStr = strings.TrimSpace(keyStr)
		keyBytes, err = hex.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hex key: %v", err)
		}
		logger.Info("KEY_LOAD", "Loaded key from hex string", true, map[string]interface{}{
			"key_length":    len(keyStr),
			"key_size_bits": len(keyBytes) * 8,
		})
	} else if keyFile != "" {
		keyBytes, err = os.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %v", err)
		}

		keyBytes = bytes.TrimSpace(keyBytes)

		if IsHexString(keyBytes) {
			keyBytes, err = hex.DecodeString(string(keyBytes))
			if err != nil {
				return nil, fmt.Errorf("invalid hex in key file: %v", err)
			}
		}

		logger.Info("KEY_LOAD", "Loaded key from file", true, map[string]interface{}{
			"key_file":      keyFile,
			"key_size_bits": len(keyBytes) * 8,
		})
	} else {
		return nil, fmt.Errorf("either --key or --keyfile must be specified")
	}

	keySize := len(keyBytes) * 8
	if keySize != 128 && keySize != 192 && keySize != 256 {
		return nil, fmt.Errorf("key must be 128, 192, or 256 bits (got %d bits)", keySize)
	}

	return keyBytes, nil
}

func IsHexString(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	if len(data) < 32 || len(data) > 64 {
		return false
	}

	hexChars := "0123456789abcdefABCDEF"
	for _, b := range data {
		if !strings.ContainsRune(hexChars, rune(b)) {
			return false
		}
	}
	return true
}
