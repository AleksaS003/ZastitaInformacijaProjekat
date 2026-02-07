package pcbc

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type PCBC struct {
	block     cipher.Block
	blockSize int
	iv        []byte
}

func NewPCBC(block cipher.Block, iv []byte) (*PCBC, error) {
	blockSize := block.BlockSize()

	if len(iv) != blockSize {
		return nil, errors.New("IV length must equal block size")
	}

	return &PCBC{
		block:     block,
		blockSize: blockSize,
		iv:        iv,
	}, nil
}

func GenerateIV(blockSize int) ([]byte, error) {
	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

func (p *PCBC) Encrypt(dst, src []byte) error {
	if len(src)%p.blockSize != 0 {
		return errors.New("input not full blocks")
	}
	if len(dst) < len(src) {
		return errors.New("output smaller than input")
	}

	prevPlain := make([]byte, p.blockSize)
	prevCipher := make([]byte, p.blockSize)
	copy(prevPlain, p.iv)
	copy(prevCipher, p.iv)

	for i := 0; i < len(src); i += p.blockSize {

		xorBlock := make([]byte, p.blockSize)
		for j := 0; j < p.blockSize; j++ {
			xorBlock[j] = src[i+j] ^ prevPlain[j] ^ prevCipher[j]
		}

		p.block.Encrypt(dst[i:], xorBlock)

		copy(prevPlain, src[i:i+p.blockSize])
		copy(prevCipher, dst[i:i+p.blockSize])
	}

	return nil
}

func (p *PCBC) Decrypt(dst, src []byte) error {
	if len(src)%p.blockSize != 0 {
		return errors.New("input not full blocks")
	}
	if len(dst) < len(src) {
		return errors.New("output smaller than input")
	}

	prevPlain := make([]byte, p.blockSize)
	prevCipher := make([]byte, p.blockSize)
	copy(prevPlain, p.iv)
	copy(prevCipher, p.iv)

	for i := 0; i < len(src); i += p.blockSize {

		temp := make([]byte, p.blockSize)
		p.block.Decrypt(temp, src[i:i+p.blockSize])

		for j := 0; j < p.blockSize; j++ {
			dst[i+j] = temp[j] ^ prevPlain[j] ^ prevCipher[j]
		}

		copy(prevPlain, dst[i:i+p.blockSize])
		copy(prevCipher, src[i:i+p.blockSize])
	}

	return nil
}
