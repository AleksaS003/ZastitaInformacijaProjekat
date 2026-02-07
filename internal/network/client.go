package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/AleksaS003/zastitaprojekat/internal/core"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

type TCPClient struct {
	address string
	conn    net.Conn
	timeout time.Duration
}

func NewTCPClient(address string, timeout time.Duration) *TCPClient {

	if err := logger.InitGlobal("./logs"); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
	}

	return &TCPClient{
		address: address,
		timeout: timeout,
	}
}

func (c *TCPClient) Connect() error {
	logger.LogNetwork(logger.CLIENT_CONNECT, c.address,
		"Connecting to server", true, map[string]interface{}{
			"timeout": c.timeout.String(),
		})

	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		logger.Error(logger.CLIENT_CONNECT, "Failed to connect to server", map[string]interface{}{
			"address": c.address,
			"timeout": c.timeout.String(),
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to connect to %s: %w", c.address, err)
	}

	c.conn = conn
	logger.LogNetwork(logger.CLIENT_CONNECT, c.address,
		"Successfully connected to server", true, nil)

	log.Printf("Connected to server: %s", c.address)
	return nil
}

func (c *TCPClient) Disconnect() error {
	if c.conn != nil {
		logger.Info(logger.CLIENT_CONNECT, "Disconnecting from server", true, map[string]interface{}{
			"address": c.address,
		})
		return c.conn.Close()
	}
	return nil
}

func (c *TCPClient) SendFile(filePath string, algorithm string, key []byte) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	originalFileInfo, err := os.Stat(filePath)
	if err != nil {
		logger.Error(logger.SEND_FILE, "Failed to get file info", map[string]interface{}{
			"file_path": filePath,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to get file info: %w", err)
	}

	logger.LogNetwork(logger.SEND_FILE, c.address,
		"Starting file transfer", true, map[string]interface{}{
			"file":          filePath,
			"algorithm":     algorithm,
			"file_size":     originalFileInfo.Size(),
			"original_name": filepath.Base(filePath),
		})

	log.Printf("Starting file transfer: %s (%d bytes)", filePath, originalFileInfo.Size())

	if err := c.doHandshake(algorithm); err != nil {
		logger.Error(logger.SEND_FILE, "Handshake failed", map[string]interface{}{
			"address":   c.address,
			"algorithm": algorithm,
			"error":     err.Error(),
		})
		return fmt.Errorf("handshake failed: %w", err)
	}

	logger.Info(logger.SEND_FILE, "Handshake successful", true, map[string]interface{}{
		"address":   c.address,
		"algorithm": algorithm,
	})

	encryptedPath, metadata, err := c.prepareFileForSending(filePath, algorithm, key)
	if err != nil {
		logger.Error(logger.SEND_FILE, "Failed to prepare file for sending", map[string]interface{}{
			"file_path": filePath,
			"algorithm": algorithm,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to prepare file: %w", err)
	}
	defer os.Remove(encryptedPath)

	encryptedFileInfo, _ := os.Stat(encryptedPath)

	logger.LogEncryption("encrypt", algorithm, filePath,
		originalFileInfo.Size(), true, map[string]interface{}{
			"encrypted_size": encryptedFileInfo.Size(),
			"temp_file":      encryptedPath,
			"hash_algorithm": metadata.HashAlgorithm,
			"hash":           metadata.Hash,
		})

	metadataJSON, err := metadata.ToJSON()
	if err != nil {
		logger.Error(logger.SEND_FILE, "Failed to serialize metadata", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	startPayload := fmt.Sprintf("%s|%d|%d",
		filepath.Base(filePath),
		encryptedFileInfo.Size(),
		len(metadataJSON))

	if err := SendMessage(c.conn, FileStartCmd, []byte(startPayload)); err != nil {
		logger.Error(logger.SEND_FILE, "Failed to send FILE_START", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to send FILE_START: %w", err)
	}

	if err := SendMessage(c.conn, "METADATA", metadataJSON); err != nil {
		logger.Error(logger.SEND_FILE, "Failed to send metadata", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	logger.Info(logger.SEND_FILE, "Metadata sent", true, map[string]interface{}{
		"metadata_size": len(metadataJSON),
		"algorithm":     metadata.EncryptionAlgorithm,
		"hash":          metadata.Hash[:16] + "...",
	})

	totalSent, chunkCount, err := c.sendFileInChunks(encryptedPath)
	if err != nil {
		logger.Error(logger.SEND_FILE, "Failed to send file data", map[string]interface{}{
			"temp_file": encryptedPath,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to send file: %w", err)
	}

	if err := SendMessage(c.conn, FileEndCmd, nil); err != nil {
		logger.Error(logger.SEND_FILE, "Failed to send FILE_END", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to send FILE_END: %w", err)
	}

	logger.Info(logger.SEND_FILE, "File data sent", true, map[string]interface{}{
		"total_bytes": totalSent,
		"chunk_size":  "32KB",
		"chunk_count": chunkCount,
	})

	return c.waitForVerification()
}

func (c *TCPClient) doHandshake(algorithm string) error {

	helloPayload := fmt.Sprintf("%s,SHA256", algorithm)
	if err := SendMessage(c.conn, HelloCmd, []byte(helloPayload)); err != nil {
		return fmt.Errorf("failed to send HELLO: %w", err)
	}

	msg, err := ReceiveMessage(c.conn)
	if err != nil {
		return fmt.Errorf("failed to receive READY: %w", err)
	}

	if msg.Command != ReadyCmd {
		return fmt.Errorf("expected READY, got %s", msg.Command)
	}

	logger.Info(logger.SEND_FILE, "Server ready response", true, map[string]interface{}{
		"server_algorithms": string(msg.Payload),
	})

	log.Printf("Server ready: %s", string(msg.Payload))
	return nil
}

func (c *TCPClient) prepareFileForSending(filePath, algorithm string, key []byte) (string, *core.Metadata, error) {
	logger.Info(logger.ENCRYPT, "Preparing file for sending", true, map[string]interface{}{
		"file":      filePath,
		"algorithm": algorithm,
	})

	tempFile := fmt.Sprintf("%s_send_%d.enc", filepath.Base(filePath), time.Now().UnixNano())

	fileProcessor := core.NewFileProcessor()
	err := fileProcessor.EncryptFileWithMetadata(filePath, tempFile, algorithm, key)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to encrypt file", map[string]interface{}{
			"file":      filePath,
			"algorithm": algorithm,
			"error":     err.Error(),
		})
		return "", nil, fmt.Errorf("failed to encrypt file: %w", err)
	}

	data, err := os.ReadFile(tempFile)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to read encrypted file", map[string]interface{}{
			"temp_file": tempFile,
			"error":     err.Error(),
		})
		return "", nil, fmt.Errorf("failed to read encrypted file: %w", err)
	}

	metadata, _, err := core.ExtractFromEncryptedFile(data)
	if err != nil {
		logger.Error(logger.ENCRYPT, "Failed to extract metadata", map[string]interface{}{
			"error": err.Error(),
		})
		return "", nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	logger.Info(logger.ENCRYPT, "File prepared successfully", true, map[string]interface{}{
		"original_file":  metadata.Filename,
		"encrypted_size": len(data),
		"hash_algorithm": metadata.HashAlgorithm,
		"hash":           metadata.Hash[:16] + "...",
	})

	return tempFile, metadata, nil
}

func (c *TCPClient) sendFileInChunks(filePath string) (int64, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error(logger.SEND_FILE, "Failed to open file", map[string]interface{}{
			"file_path": filePath,
			"error":     err.Error(),
		})
		return 0, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		logger.Error(logger.SEND_FILE, "Failed to get file info", map[string]interface{}{
			"file_path": filePath,
			"error":     err.Error(),
		})
		return 0, 0, fmt.Errorf("failed to get file info: %w", err)
	}

	buffer := make([]byte, 32*1024)
	var totalSent int64
	var chunkCount int

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(logger.SEND_FILE, "Failed to read file chunk", map[string]interface{}{
				"file_path":   filePath,
				"chunk_index": chunkCount,
				"error":       err.Error(),
			})
			return totalSent, chunkCount, fmt.Errorf("failed to read file: %w", err)
		}

		if err := SendMessage(c.conn, FileDataCmd, buffer[:n]); err != nil {
			logger.Error(logger.SEND_FILE, "Failed to send file chunk", map[string]interface{}{
				"chunk_index": chunkCount,
				"chunk_size":  n,
				"error":       err.Error(),
			})
			return totalSent, chunkCount, fmt.Errorf("failed to send file data: %w", err)
		}

		totalSent += int64(n)
		chunkCount++

		if chunkCount%10 == 0 {
			progress := float64(totalSent) * 100 / float64(fileInfo.Size())
			logger.Info(logger.SEND_FILE, "File transfer progress", true, map[string]interface{}{
				"chunks_sent": chunkCount,
				"bytes_sent":  totalSent,
				"progress":    fmt.Sprintf("%.1f%%", progress),
			})
		}
	}

	logger.Info(logger.SEND_FILE, "File transfer completed", true, map[string]interface{}{
		"total_chunks":       chunkCount,
		"total_bytes":        totalSent,
		"average_chunk_size": totalSent / int64(chunkCount),
	})

	log.Printf("Total sent: %d bytes in %d chunks", totalSent, chunkCount)
	return totalSent, chunkCount, nil
}

func (c *TCPClient) waitForVerification() error {
	logger.Info(logger.SEND_FILE, "Waiting for server verification", true, nil)

	msg, err := ReceiveMessage(c.conn)
	if err != nil {
		logger.Error(logger.SEND_FILE, "Failed to receive verification", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to receive verification: %w", err)
	}

	switch msg.Command {
	case SuccessCmd:
		logger.LogNetwork(logger.SEND_FILE, c.address,
			"File successfully received and verified by server", true, map[string]interface{}{
				"server_response": string(msg.Payload),
			})

		log.Printf(" File successfully received and verified by server")
		return nil

	case ErrorCmd:
		errorMsg := string(msg.Payload)
		logger.Error(logger.SEND_FILE, "Server reported error", map[string]interface{}{
			"server_error": errorMsg,
		})
		return fmt.Errorf("server error: %s", errorMsg)

	default:
		logger.Error(logger.SEND_FILE, "Unexpected server response", map[string]interface{}{
			"command": msg.Command,
			"payload": string(msg.Payload),
		})
		return fmt.Errorf("unexpected response: %s", msg.Command)
	}
}
