package tcp

import (
	"encoding/binary"
	"errors"
	"key-value-store/internal/util"
)

const (
	HeaderSize     = 13
	MaxPayloadSize = 16 * 1024 * 1024
)

const (
	CmdSet      byte = 0x01
	CmdGet      byte = 0x02
	CmdDelete   byte = 0x03
	CmdAuth     byte = 0x20
	CmdResponse byte = 0xF0
	CmdError    byte = 0xFF
)

const (
	StatusOK            byte = 0x00
	StatusCreated       byte = 0x01
	StatusNoContent     byte = 0x02
	StatusBadRequest    byte = 0x10
	StatusUnauthorized  byte = 0x11
	StatusNotFound      byte = 0x12
	StatusConflict      byte = 0x13
	StatusInternalError byte = 0x20
	StatusInvalidTTL    byte = 0x21
	StatusKeyExpired    byte = 0x22
)

var (
	ErrInvalidFrame    = errors.New("invalid frame")
	ErrPayloadTooLarge = errors.New("payload too large")
	ErrIncompleteFrame = errors.New("incomplete frame")
)

// Format: [Length(4)][Command(1)][RequestID(8)][Payload(variable)]
type Frame struct {
	Length    uint32
	Command   byte
	RequestID uint64
	Payload   []byte
}

func (f *Frame) Encode() []byte {
	totalLen := HeaderSize + len(f.Payload)
	buf := make([]byte, totalLen)

	binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen))

	buf[4] = f.Command

	binary.BigEndian.PutUint64(buf[5:13], f.RequestID)

	if len(f.Payload) > 0 {
		copy(buf[13:], f.Payload)
	}

	return buf
}

func DecodeFrame(data []byte) (*Frame, error) {
	if len(data) < HeaderSize {
		return nil, ErrIncompleteFrame
	}

	length := binary.BigEndian.Uint32(data[0:4])

	if length > MaxPayloadSize+HeaderSize {
		return nil, ErrPayloadTooLarge
	}

	if uint32(len(data)) < length {
		return nil, ErrIncompleteFrame
	}

	command := data[4]

	requestID := binary.BigEndian.Uint64(data[5:13])

	var payload []byte
	if length > HeaderSize {
		payload = make([]byte, length-HeaderSize)
		copy(payload, data[13:length])
	}

	return &Frame{
		Length:    length,
		Command:   command,
		RequestID: requestID,
		Payload:   payload,
	}, nil
}

func NewFrame(command byte, requestID uint64, payload []byte) *Frame {
	return &Frame{
		Length:    uint32(HeaderSize + len(payload)),
		Command:   command,
		RequestID: requestID,
		Payload:   payload,
	}
}

func NewResponseFrame(requestID uint64, status byte, data []byte) *Frame {
	payload := make([]byte, 1+len(data))
	payload[0] = status
	if len(data) > 0 {
		copy(payload[1:], data)
	}

	return NewFrame(CmdResponse, requestID, payload)
}

func NewErrorFrame(requestID uint64, status byte, message string) *Frame {
	msgBytes := util.StringToBytes(message)
	payload := make([]byte, 1+len(msgBytes))
	payload[0] = status
	copy(payload[1:], msgBytes)

	return NewFrame(CmdError, requestID, payload)
}

func ParseResponsePayload(payload []byte) (status byte, data []byte, err error) {
	if len(payload) < 1 {
		return 0, nil, ErrInvalidFrame
	}

	status = payload[0]
	if len(payload) > 1 {
		data = payload[1:]
	}

	return status, data, nil
}
