package lea

import (
	"encoding/binary"
	"errors"
)

type LEA struct {
	roundKeys [][6]uint32
	keySize   int
	rounds    int
}

func NewLEA(key []byte) (*LEA, error) {
	keySize := len(key) * 8

	if keySize != 128 && keySize != 192 && keySize != 256 {
		return nil, errors.New("LEA key must be 128, 192, or 256 bits")
	}

	lea := &LEA{
		keySize: keySize,
	}

	switch keySize {
	case 128:
		lea.rounds = 24
	case 192:
		lea.rounds = 28
	case 256:
		lea.rounds = 32
	}

	if err := lea.keySchedule(key); err != nil {
		return nil, err
	}

	return lea, nil
}

func (l *LEA) EncryptBlock(block []byte) ([]byte, error) {
	if len(block) != 16 {
		return nil, errors.New("block must be 16 bytes (128 bits)")
	}

	var x [4]uint32
	for i := 0; i < 4; i++ {
		x[i] = binary.LittleEndian.Uint32(block[i*4 : (i+1)*4])
	}

	for r := 0; r < l.rounds; r++ {
		x[0] = l.rotateLeft(x[0]+(l.roundKeys[r][0]^x[1]^(l.roundKeys[r][1]&x[2])), 9)
		x[1] = l.rotateRight(x[1]+(l.roundKeys[r][2]^x[2]^(l.roundKeys[r][3]&x[3])), 5)
		x[2] = l.rotateRight(x[2]+(l.roundKeys[r][4]^x[3]^(l.roundKeys[r][5]&x[0])), 3)

		x[0], x[1], x[2], x[3] = x[1], x[2], x[3], x[0]
	}

	result := make([]byte, 16)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint32(result[i*4:(i+1)*4], x[i])
	}

	return result, nil
}

func (l *LEA) DecryptBlock(block []byte) ([]byte, error) {
	if len(block) != 16 {
		return nil, errors.New("block must be 16 bytes (128 bits)")
	}

	var x [4]uint32
	for i := 0; i < 4; i++ {
		x[i] = binary.LittleEndian.Uint32(block[i*4 : (i+1)*4])
	}

	for r := l.rounds - 1; r >= 0; r-- {

		x[0], x[1], x[2], x[3] = x[3], x[0], x[1], x[2]

		x[2] = l.rotateLeft(x[2], 3) - (l.roundKeys[r][4] ^ x[3] ^ (l.roundKeys[r][5] & x[0]))
		x[1] = l.rotateLeft(x[1], 5) - (l.roundKeys[r][2] ^ x[2] ^ (l.roundKeys[r][3] & x[3]))
		x[0] = l.rotateRight(x[0], 9) - (l.roundKeys[r][0] ^ x[1] ^ (l.roundKeys[r][1] & x[2]))
	}

	result := make([]byte, 16)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint32(result[i*4:(i+1)*4], x[i])
	}

	return result, nil
}

func (l *LEA) keySchedule(key []byte) error {

	delta := [8]uint32{
		0xc3efe9db, 0x44626b02, 0x79e27c8a, 0x78df30ec,
		0x715ea49e, 0xc785da0a, 0xe04ef22a, 0xe5c40957,
	}

	var K [8]uint32
	for i := 0; i < len(key)/4; i++ {
		K[i] = binary.LittleEndian.Uint32(key[i*4 : (i+1)*4])
	}

	l.roundKeys = make([][6]uint32, l.rounds)

	var T [8]uint32
	copy(T[:], K[:])

	for i := 0; i < l.rounds; i++ {

		for j := 0; j < 8; j++ {
			T[j] = l.rotateLeft(T[j]+delta[j&3], uint((j%4)+1))
		}

		switch l.keySize {
		case 128:
			l.roundKeys[i] = [6]uint32{T[0], T[1], T[2], T[3], T[0], T[1]}
		case 192:
			l.roundKeys[i] = [6]uint32{T[0], T[1], T[2], T[3], T[4], T[5]}
		case 256:
			l.roundKeys[i] = [6]uint32{T[(6*i)%8], T[(6*i+1)%8], T[(6*i+2)%8],
				T[(6*i+3)%8], T[(6*i+4)%8], T[(6*i+5)%8]}
		}
	}

	return nil
}

func (l *LEA) rotateLeft(x uint32, k uint) uint32 {
	return (x << k) | (x >> (32 - k))
}

func (l *LEA) rotateRight(x uint32, k uint) uint32 {
	return (x >> k) | (x << (32 - k))
}
