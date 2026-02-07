package network

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	HelloCmd     = "HELLO"
	ReadyCmd     = "READY"
	FileStartCmd = "FILE_START"
	FileDataCmd  = "FILE_DATA"
	FileEndCmd   = "FILE_END"
	VerifyCmd    = "VERIFY"
	SuccessCmd   = "SUCCESS"
	FailCmd      = "FAIL"
	ErrorCmd     = "ERROR"

	MaxPacketSize = 64 * 1024
)

type Message struct {
	Command string
	Payload []byte
}

func EncodeMessage(msg Message) []byte {

	cmdBytes := []byte(msg.Command)
	payloadLen := len(msg.Payload)

	totalLen := 2 + len(cmdBytes) + 4 + payloadLen
	data := make([]byte, totalLen)

	binary.BigEndian.PutUint16(data[0:2], uint16(len(cmdBytes)))

	copy(data[2:2+len(cmdBytes)], cmdBytes)

	payloadStart := 2 + len(cmdBytes)
	binary.BigEndian.PutUint32(data[payloadStart:payloadStart+4], uint32(payloadLen))

	if payloadLen > 0 {
		copy(data[payloadStart+4:], msg.Payload)
	}

	return data
}

func DecodeMessage(reader io.Reader) (Message, error) {
	var msg Message

	var cmdLen uint16
	if err := binary.Read(reader, binary.BigEndian, &cmdLen); err != nil {
		return msg, fmt.Errorf("failed to read command length: %w", err)
	}

	cmdBytes := make([]byte, cmdLen)
	if _, err := io.ReadFull(reader, cmdBytes); err != nil {
		return msg, fmt.Errorf("failed to read command: %w", err)
	}
	msg.Command = string(cmdBytes)

	var payloadLen uint32
	if err := binary.Read(reader, binary.BigEndian, &payloadLen); err != nil {
		return msg, fmt.Errorf("failed to read payload length: %w", err)
	}

	if payloadLen > 0 {
		msg.Payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(reader, msg.Payload); err != nil {
			return msg, fmt.Errorf("failed to read payload: %w", err)
		}
	}

	return msg, nil
}

func SendMessage(writer io.Writer, command string, payload []byte) error {
	msg := Message{Command: command, Payload: payload}
	data := EncodeMessage(msg)

	_, err := writer.Write(data)
	return err
}

func ReceiveMessage(reader io.Reader) (Message, error) {
	return DecodeMessage(reader)
}
