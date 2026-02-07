package sha256

import (
	"fmt"
	"io"
	"os"
)

func HashBytes(data []byte) [32]byte {
	s := NewSHA256()
	s.Write(data)
	return s.Sum256()
}

func HashString(data string) [32]byte {
	return HashBytes([]byte(data))
}

func HashFile(filename string) ([32]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	s := NewSHA256()

	buf := make([]byte, 32*1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			s.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return [32]byte{}, err
		}
	}

	return s.Sum256(), nil
}

func HashToString(hash [32]byte) string {
	return fmt.Sprintf("%x", hash)
}

func VerifyHash(hash1, hash2 [32]byte) bool {
	for i := 0; i < 32; i++ {
		if hash1[i] != hash2[i] {
			return false
		}
	}
	return true
}
