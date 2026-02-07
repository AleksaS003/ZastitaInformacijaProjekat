package core

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Metadata struct {
	Filename            string    `json:"filename"`
	Size                int64     `json:"size"`
	Timestamp           time.Time `json:"timestamp"`
	EncryptionAlgorithm string    `json:"encryption_algorithm"`
	HashAlgorithm       string    `json:"hash_algorithm,omitempty"`
	Hash                string    `json:"hash,omitempty"`
	IV                  string    `json:"iv,omitempty"`
	KeyInfo             string    `json:"key_info,omitempty"`
}

func NewMetadata(filepath string, encAlgo string, hashAlgo string, hash string, iv []byte) (*Metadata, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{
		Filename:            filepath,
		Size:                fileInfo.Size(),
		Timestamp:           time.Now().UTC(),
		EncryptionAlgorithm: encAlgo,
		HashAlgorithm:       hashAlgo,
		Hash:                hash,
	}

	if iv != nil {
		metadata.IV = fmt.Sprintf("%x", iv)
	}

	return metadata, nil
}

func (m *Metadata) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

func FromJSON(data []byte) (*Metadata, error) {
	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (m *Metadata) WriteToFile(filepath string) error {
	data, err := m.ToJSON()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func (m *Metadata) AddToEncryptedFile(metadataPath []byte, encryptedData []byte) ([]byte, error) {

	metadataJSON, err := m.ToJSON()
	if err != nil {
		return nil, err
	}

	header := make([]byte, 4+len(metadataJSON))

	metadataLen := uint32(len(metadataJSON))
	header[0] = byte(metadataLen)
	header[1] = byte(metadataLen >> 8)
	header[2] = byte(metadataLen >> 16)
	header[3] = byte(metadataLen >> 24)

	copy(header[4:], metadataJSON)

	result := make([]byte, len(header)+len(encryptedData))
	copy(result, header)
	copy(result[len(header):], encryptedData)

	return result, nil
}

func ExtractFromEncryptedFile(data []byte) (*Metadata, []byte, error) {
	if len(data) < 4 {
		return nil, nil, fmt.Errorf("data too short for metadata header")
	}

	metadataLen := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24

	if len(data) < int(4+metadataLen) {
		return nil, nil, fmt.Errorf("invalid metadata length")
	}

	metadataJSON := data[4 : 4+metadataLen]

	metadata, err := FromJSON(metadataJSON)
	if err != nil {
		return nil, nil, err
	}

	encryptedData := data[4+metadataLen:]

	return metadata, encryptedData, nil
}
