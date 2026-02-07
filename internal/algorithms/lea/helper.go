package lea

import (
	"crypto/rand"
	"errors"
	"io"
)

func (l *LEA) Encrypt(data []byte) ([]byte, error) {

	padded := l.addPadding(data)

	result := make([]byte, 0, len(padded))

	for i := 0; i < len(padded); i += 16 {
		block := padded[i : i+16]
		encrypted, err := l.EncryptBlock(block)
		if err != nil {
			return nil, err
		}
		result = append(result, encrypted...)
	}

	return result, nil
}

func (l *LEA) Decrypt(data []byte) ([]byte, error) {
	if len(data)%16 != 0 {
		return nil, errors.New("ciphertext length must be multiple of 16")
	}

	result := make([]byte, 0, len(data))

	for i := 0; i < len(data); i += 16 {
		block := data[i : i+16]
		decrypted, err := l.DecryptBlock(block)
		if err != nil {
			return nil, err
		}
		result = append(result, decrypted...)
	}

	return l.removePadding(result), nil
}

func (l *LEA) addPadding(data []byte) []byte {
	padding := 16 - (len(data) % 16)
	if padding == 0 {
		padding = 16
	}

	padded := make([]byte, len(data)+padding)
	copy(padded, data)

	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	return padded
}

func (l *LEA) removePadding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	padding := int(data[len(data)-1])
	if padding > 16 || padding == 0 {
		return data
	}

	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return data
		}
	}

	return data[:len(data)-padding]
}

func GenerateKey(size int) ([]byte, error) {

	if size != 128 && size != 192 && size != 256 {
		return nil, errors.New("key size must be 128, 192, or 256 bits")
	}

	bytes := size / 8
	key := make([]byte, bytes)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return key, nil
}
