package handlers

import "fmt"

func PrintHelp() {
	fmt.Println(`Crypto App - CLI Tool
=========================
Version: 1.0.0 with Activity Logging

Usage:
  crypto-cli <command> [arguments]

Commands:
  server        - Start TCP server to receive files
  client        - Send file to TCP server
  
  fsw           - File System Watcher
    start       - Start watching directory
    stop        - Stop watching directory
    encrypt-existing - Encrypt existing files
  
  encrypt-file  - Encrypt file with metadata
  decrypt-file  - Decrypt file with metadata
  
  foursquare    - Use Foursquare cipher
    encrypt     - Encrypt text/file
    decrypt     - Decrypt text/file
  
  lea           - Use LEA encryption
    encrypt     - Encrypt file
    decrypt     - Decrypt file
    genkey      - Generate key (console output)
    genkey-file - Generate key file
  
  pcbc          - Use PCBC mode with LEA
    encrypt     - Encrypt file
    decrypt     - Decrypt file
  
  sha256        - Use SHA-256 hash function
    hash        - Hash text/file
    verify      - Verify file hash
  
  logs          - Manage activity logs
    show        - Show recent logs
    clear       - Clear log files
    stats       - Show log statistics
  
  help          - Show this help message

Examples:
  # Foursquare
  crypto-cli foursquare encrypt --text="HELLO" --key1=test --key2=key
  crypto-cli foursquare encrypt --file=message.txt --output=encrypted.txt
  
  # LEA
  crypto-cli lea genkey-file --size=128 --output=mykey.bin
  crypto-cli lea encrypt --file=secret.txt --keyfile=mykey.bin --output=encrypted.bin
  crypto-cli lea decrypt --file=encrypted.bin --keyfile=mykey.bin --output=decrypted.txt

  # LEA with hex key directly
  crypto-cli lea encrypt --file=secret.txt --key=00112233445566778899aabbccddeeff --output=encrypted.bin

  # PCBC Mode
  crypto-cli pcbc encrypt --file=data.txt --keyfile=key.bin --output=data.enc
  crypto-cli pcbc decrypt --file=data.enc --keyfile=key.bin --output=data.txt

  # SHA-256
  crypto-cli sha256 hash --file=document.pdf
  crypto-cli sha256 verify --file=document.pdf --hashfile=document.pdf.sha256

  # File System Watcher
  crypto-cli fsw start --watch=./watch --output=./encrypted --keyfile=key.bin
  crypto-cli fsw encrypt-existing --watch=./watch --keyfile=key.bin

  # TCP Server/Client
  crypto-cli server --address=:8080 --output=./received --keyfile=key.bin
  crypto-cli client --address=localhost:8080 --file=data.txt --keyfile=key.bin

  # Log Management
  crypto-cli logs show -n 100
  crypto-cli logs stats
  crypto-cli logs clear --yes

Log Files:
  - logs/crypto-app.log     - Text logs (human readable)
  - logs/activity-log.json  - JSON logs (for analysis)
  - fsw.log                 - FSW specific logs

All activities are logged for security auditing and monitoring.`)
}
