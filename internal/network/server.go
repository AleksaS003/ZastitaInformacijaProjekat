// internal/network/server.go
package network

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AleksaS003/zastitaprojekat/internal/algorithms/sha256"
	"github.com/AleksaS003/zastitaprojekat/internal/core"
	"github.com/AleksaS003/zastitaprojekat/internal/logger"
)

// TCPServer implementira server za prijem fajlova
type TCPServer struct {
	address   string
	listener  net.Listener
	clients   map[net.Conn]bool
	mu        sync.RWMutex
	active    bool
	stopChan  chan struct{}
	outputDir string
	key       []byte
}

func NewTCPServer(address, outputDir string, key []byte) *TCPServer {
	if err := logger.InitGlobal("./logs"); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
	}

	return &TCPServer{
		address:   address,
		clients:   make(map[net.Conn]bool),
		active:    false,
		stopChan:  make(chan struct{}),
		outputDir: outputDir,
		key:       key,
	}
}

// Start pokreće server
func (s *TCPServer) Start() error {
	if s.active {
		return fmt.Errorf("server is already running")
	}

	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		logger.Error(logger.SERVER_START, "Failed to create output directory", map[string]interface{}{
			"directory": s.outputDir,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		logger.Error(logger.SERVER_START, "Failed to start TCP listener", map[string]interface{}{
			"address": s.address,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to start listener: %w", err)
	}

	s.listener = listener
	s.active = true

	logger.LogNetwork(logger.SERVER_START, s.address,
		"TCP Server started successfully", true, map[string]interface{}{
			"output_dir": s.outputDir,
		})

	log.Printf("🚀 TCP Server started on %s", s.address)
	log.Printf("   Output directory: %s", s.outputDir)

	// Accept loop - OVAJ LOOP MORA DA TRAJE DOK SE SERVER NE ZAUSTAVI
	go s.acceptLoop()

	return nil
}

// Stop zaustavlja server
func (s *TCPServer) Stop() error {
	if !s.active {
		return fmt.Errorf("server is not running")
	}

	close(s.stopChan)
	s.active = false

	logger.LogNetwork(logger.SERVER_STOP, s.address,
		"TCP Server stopped", true, nil)

	s.mu.Lock()
	clientCount := len(s.clients)
	for conn := range s.clients {
		conn.Close()
		delete(s.clients, conn)
	}
	s.mu.Unlock()

	log.Printf("Server stopped. Disconnected %d clients.", clientCount)

	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

// acceptLoop prihvata nove konekcije
func (s *TCPServer) acceptLoop() {
	logger.Info(logger.SERVER_START, "Server accept loop started", true, map[string]interface{}{
		"address": s.address,
	})

	for s.active {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.active {
				logger.Error(logger.CLIENT_CONNECT, "Accept error", map[string]interface{}{
					"address": s.address,
					"error":   err.Error(),
				})
				log.Printf("Accept error: %v", err)
			}
			// NE IZLAZI IZ PETLJE - SAMO LOGUJ I NASTAVI
			continue
		}

		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()

		// Svaka konekcija dobija svoju gorutinu
		go s.handleConnection(conn)
	}
}

// handleConnection obrađuje konekciju sa klijentom
func (s *TCPServer) handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()

	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()

		logger.Info(logger.CLIENT_CONNECT, "Client disconnected", true, map[string]interface{}{
			"remote_addr": remoteAddr,
		})

		log.Printf("Client disconnected: %s", remoteAddr)
	}()

	logger.LogNetwork(logger.CLIENT_CONNECT, remoteAddr,
		"New client connected", true, nil)

	log.Printf("New client connected: %s", remoteAddr)

	// 1. Handshake
	if err := s.doHandshake(conn); err != nil {
		logger.Error(logger.CLIENT_CONNECT, "Handshake failed", map[string]interface{}{
			"remote_addr": remoteAddr,
			"error":       err.Error(),
		})
		log.Printf("Handshake failed with %s: %v", remoteAddr, err)
		return
	}

	logger.Info(logger.CLIENT_CONNECT, "Handshake successful", true, map[string]interface{}{
		"remote_addr": remoteAddr,
	})

	// 2. Prijem fajla
	filePath, metadata, err := s.receiveFile(conn)
	if err != nil {
		logger.Error(logger.RECEIVE_FILE, "File receive failed", map[string]interface{}{
			"remote_addr": remoteAddr,
			"error":       err.Error(),
		})
		log.Printf("File receive failed from %s: %v", remoteAddr, err)
		SendMessage(conn, ErrorCmd, []byte(err.Error()))
		return
	}

	logger.Info(logger.RECEIVE_FILE, "File received from client", true, map[string]interface{}{
		"remote_addr":    remoteAddr,
		"temp_file":      filePath,
		"original_file":  metadata.Filename,
		"algorithm":      metadata.EncryptionAlgorithm,
		"hash_algorithm": metadata.HashAlgorithm,
	})

	// 3. Verifikacija i dekripcija
	if err := s.verifyAndDecrypt(filePath, metadata); err != nil {
		logger.Error(logger.RECEIVE_FILE, "File verification/decryption failed", map[string]interface{}{
			"remote_addr": remoteAddr,
			"file":        metadata.Filename,
			"error":       err.Error(),
		})
		log.Printf("Verification failed for %s: %v", remoteAddr, err)
		SendMessage(conn, ErrorCmd, []byte(err.Error()))
		return
	}

	// 4. Success
	outputPath := filepath.Join(s.outputDir, metadata.Filename)
	fileInfo, _ := os.Stat(outputPath)

	logger.LogEncryption("decrypt", metadata.EncryptionAlgorithm, outputPath,
		fileInfo.Size(), true, map[string]interface{}{
			"remote_addr":    remoteAddr,
			"hash_verified":  true,
			"hash_algorithm": metadata.HashAlgorithm,
		})

	log.Printf("File successfully received from %s: %s", remoteAddr, metadata.Filename)
	SendMessage(conn, SuccessCmd, []byte("File received and verified"))

	// 5. NE GASI SERVER! Samo zatvori konekciju (defer će to uraditi)
	// Server nastavlja da radi i čeka nove konekcije
}

// doHandshake izvršava inicijalni handshake
func (s *TCPServer) doHandshake(conn net.Conn) error {
	msg, err := ReceiveMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to receive HELLO: %w", err)
	}

	if msg.Command != HelloCmd {
		return fmt.Errorf("expected HELLO, got %s", msg.Command)
	}

	logger.Info(logger.CLIENT_CONNECT, "Client algorithms", true, map[string]interface{}{
		"remote_addr": conn.RemoteAddr().String(),
		"algorithms":  string(msg.Payload),
	})

	log.Printf("Client algorithms: %s", string(msg.Payload))

	return SendMessage(conn, ReadyCmd, []byte("LEA,PCBC,SHA256"))
}

// receiveFile prima fajl od klijenta
func (s *TCPServer) receiveFile(conn net.Conn) (string, *core.Metadata, error) {
	remoteAddr := conn.RemoteAddr().String()

	// 1. FILE_START poruka
	startMsg, err := ReceiveMessage(conn)
	if err != nil {
		return "", nil, fmt.Errorf("failed to receive FILE_START: %w", err)
	}

	if startMsg.Command != FileStartCmd {
		return "", nil, fmt.Errorf("expected FILE_START, got %s", startMsg.Command)
	}

	logger.Info(logger.RECEIVE_FILE, "File transfer started", true, map[string]interface{}{
		"remote_addr": remoteAddr,
		"start_info":  string(startMsg.Payload),
	})

	// 2. Prijem metadata
	metadataMsg, err := ReceiveMessage(conn)
	if err != nil {
		return "", nil, fmt.Errorf("failed to receive metadata: %w", err)
	}

	metadata, err := core.FromJSON(metadataMsg.Payload)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// 3. Generiši putanju za privremeni fajl
	tempFile := filepath.Join(s.outputDir, fmt.Sprintf("temp_%d.enc", time.Now().UnixNano()))
	file, err := os.Create(tempFile)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// 4. Prijem fajla (chunkovano)
	var totalReceived int64

	for {
		msg, err := ReceiveMessage(conn)
		if err != nil {
			return "", nil, fmt.Errorf("failed to receive file data: %w", err)
		}

		if msg.Command == FileEndCmd {
			break
		}

		if msg.Command != FileDataCmd {
			return "", nil, fmt.Errorf("expected FILE_DATA or FILE_END, got %s", msg.Command)
		}

		n, err := file.Write(msg.Payload)
		if err != nil {
			return "", nil, fmt.Errorf("failed to write to file: %w", err)
		}

		totalReceived += int64(n)
	}

	logger.Info(logger.RECEIVE_FILE, "File transfer completed", true, map[string]interface{}{
		"remote_addr":    remoteAddr,
		"temp_file":      tempFile,
		"bytes_received": totalReceived,
		"original_file":  metadata.Filename,
		"algorithm":      metadata.EncryptionAlgorithm,
	})

	log.Printf("File received: %s (%d bytes)", tempFile, totalReceived)
	return tempFile, metadata, nil
}

// verifyAndDecrypt verifikuje i dekriptuje primljeni fajl
func (s *TCPServer) verifyAndDecrypt(encryptedPath string, metadata *core.Metadata) error {
	filename := metadata.Filename
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}
	if idx := strings.LastIndex(filename, "\\"); idx != -1 {
		filename = filename[idx+1:]
	}

	outputPath := filepath.Join(s.outputDir, filename)

	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		logger.Error(logger.RECEIVE_FILE, "Failed to create output directory", map[string]interface{}{
			"directory": s.outputDir,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	logger.Info(logger.DECRYPT, "Starting file decryption", true, map[string]interface{}{
		"input_file":  encryptedPath,
		"output_file": outputPath,
		"algorithm":   metadata.EncryptionAlgorithm,
		"hash_check":  metadata.Hash != "",
	})

	fileProcessor := core.NewFileProcessor()
	_, err := fileProcessor.DecryptFileWithMetadata(encryptedPath, outputPath, s.key)
	if err != nil {
		logger.Error(logger.DECRYPT, "Decryption failed", map[string]interface{}{
			"input_file":  encryptedPath,
			"output_file": outputPath,
			"algorithm":   metadata.EncryptionAlgorithm,
			"error":       err.Error(),
		})
		return fmt.Errorf("decryption/verification failed: %w", err)
	}

	if metadata.Hash != "" {
		decryptedData, err := os.ReadFile(outputPath)
		if err == nil {
			decryptedHash := sha256.HashBytes(decryptedData)
			decryptedHashStr := sha256.HashToString(decryptedHash)

			hashMatch := decryptedHashStr == metadata.Hash
			logger.LogHashVerification(outputPath, metadata.Hash, decryptedHashStr, hashMatch)

			if !hashMatch {
				logger.Error(logger.VERIFY_HASH, "Hash verification failed", map[string]interface{}{
					"file":          outputPath,
					"expected_hash": metadata.Hash,
					"actual_hash":   decryptedHashStr,
				})
				return fmt.Errorf("hash verification failed")
			}
		}
	}

	os.Remove(encryptedPath)

	fileInfo, _ := os.Stat(outputPath)
	logger.LogEncryption("decrypt", metadata.EncryptionAlgorithm, outputPath,
		fileInfo.Size(), true, map[string]interface{}{
			"hash_verified":  metadata.Hash != "",
			"hash_algorithm": metadata.HashAlgorithm,
			"iv_used":        metadata.IV != "",
		})

	log.Printf("✅ File successfully decrypted and verified: %s (%d bytes)",
		filename, fileInfo.Size())
	return nil
}
