package pcbc

import (
	"fmt"

	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/lea"
)

type LEAPCBC struct {
	leaCipher *lea.LEA
	pcbc      *PCBC
	iv        []byte
}

func NewLEAPCBC(key []byte) (*LEAPCBC, error) {

	leaCipher, err := lea.NewLEA(key)
	if err != nil {
		return nil, err
	}

	block := &leaBlockAdapter{leaCipher}

	iv, err := GenerateIV(block.BlockSize())
	if err != nil {
		return nil, err
	}

	pcbc, err := NewPCBC(block, iv)
	if err != nil {
		return nil, err
	}

	return &LEAPCBC{
		leaCipher: leaCipher,
		pcbc:      pcbc,
		iv:        iv,
	}, nil
}

func NewLEAPCBCWithIV(key, iv []byte) (*LEAPCBC, error) {
	leaCipher, err := lea.NewLEA(key)
	if err != nil {
		return nil, err
	}

	block := &leaBlockAdapter{leaCipher}
	pcbc, err := NewPCBC(block, iv)
	if err != nil {
		return nil, err
	}

	return &LEAPCBC{
		leaCipher: leaCipher,
		pcbc:      pcbc,
		iv:        iv,
	}, nil
}

func (l *LEAPCBC) Encrypt(plaintext []byte) ([]byte, error) {

	padded := addPadding(plaintext, l.pcbc.blockSize)

	ciphertext := make([]byte, len(l.iv)+len(padded))
	copy(ciphertext, l.iv)

	err := l.pcbc.Encrypt(ciphertext[len(l.iv):], padded)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func (l *LEAPCBC) Decrypt(ciphertext []byte) ([]byte, error) {

	if len(ciphertext) < len(l.iv) {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:len(l.iv)]
	data := ciphertext[len(l.iv):]

	block := &leaBlockAdapter{l.leaCipher}
	pcbc, err := NewPCBC(block, iv)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(data))

	err = pcbc.Decrypt(plaintext, data)
	if err != nil {
		return nil, err
	}

	return removePadding(plaintext), nil
}

func (l *LEAPCBC) GetIV() []byte {
	return l.iv
}

type leaBlockAdapter struct {
	leaCipher *lea.LEA
}

func (a *leaBlockAdapter) BlockSize() int {
	return 16
}

func (a *leaBlockAdapter) Encrypt(dst, src []byte) {
	encrypted, err := a.leaCipher.EncryptBlock(src)
	if err != nil {
		panic(err)
	}
	copy(dst, encrypted)
}

func (a *leaBlockAdapter) Decrypt(dst, src []byte) {
	decrypted, err := a.leaCipher.DecryptBlock(src)
	if err != nil {
		panic(err)
	}
	copy(dst, decrypted)
}

func addPadding(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	if padding == 0 {
		padding = blockSize
	}

	padded := make([]byte, len(data)+padding)
	copy(padded, data)

	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	return padded
}

func removePadding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return data
	}

	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return data
		}
	}

	return data[:len(data)-padding]
}
