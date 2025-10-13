package tcp

import (
	"encoding/binary"
	"key-value-store/internal/util"
)

func writeString(buf []byte, offset int, s string) int {
	strBytes := util.StringToBytes(s)
	binary.BigEndian.PutUint16(buf[offset:], uint16(len(strBytes)))
	offset += 2
	copy(buf[offset:], strBytes)
	return offset + len(strBytes)
}

func readString(data []byte, offset int) (string, int, error) {
	if len(data) < offset+2 {
		return "", offset, ErrInvalidFrame
	}
	strLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2
	if len(data) < offset+strLen {
		return "", offset, ErrInvalidFrame
	}
	str := util.BytesToString(data[offset : offset+strLen])
	return str, offset + strLen, nil
}

// Format: [TokenLen(2)][Token][BucketLen(2)][Bucket][KeyLen(2)][Key][TTL(8)][SingleRead(1)][ValueLen(4)][Value]
func EncodeSetPayload(token, bucket, key string, ttl int64, singleRead bool, value []byte) []byte {
	size := 2 + len(token) + 2 + len(bucket) + 2 + len(key) + 8 + 1 + 4 + len(value)
	buf := make([]byte, size)

	offset := 0
	offset = writeString(buf, offset, token)
	offset = writeString(buf, offset, bucket)
	offset = writeString(buf, offset, key)

	binary.BigEndian.PutUint64(buf[offset:], uint64(ttl))
	offset += 8

	if singleRead {
		buf[offset] = 1
	}
	offset++

	binary.BigEndian.PutUint32(buf[offset:], uint32(len(value)))
	offset += 4
	copy(buf[offset:], value)

	return buf
}

func DecodeSetPayload(data []byte) (token, bucket, key string, ttl int64, singleRead bool, value []byte, err error) {
	offset := 0

	token, offset, err = readString(data, offset)
	if err != nil {
		return
	}

	bucket, offset, err = readString(data, offset)
	if err != nil {
		return
	}

	key, offset, err = readString(data, offset)
	if err != nil {
		return
	}

	if len(data) < offset+8 {
		err = ErrInvalidFrame
		return
	}
	ttl = int64(binary.BigEndian.Uint64(data[offset:]))
	offset += 8

	if len(data) < offset+1 {
		err = ErrInvalidFrame
		return
	}
	singleRead = data[offset] == 1
	offset++

	if len(data) < offset+4 {
		err = ErrInvalidFrame
		return
	}
	valueLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+valueLen {
		err = ErrInvalidFrame
		return
	}
	value = make([]byte, valueLen)
	copy(value, data[offset:offset+valueLen])

	return
}

// Format: [TokenLen(2)][Token][BucketLen(2)][Bucket][KeyLen(2)][Key]
func EncodeGetPayload(token, bucket, key string) []byte {
	size := 2 + len(token) + 2 + len(bucket) + 2 + len(key)
	buf := make([]byte, size)

	offset := 0
	offset = writeString(buf, offset, token)
	offset = writeString(buf, offset, bucket)
	writeString(buf, offset, key)

	return buf
}

func DecodeGetPayload(data []byte) (token, bucket, key string, err error) {
	offset := 0

	token, offset, err = readString(data, offset)
	if err != nil {
		return
	}

	bucket, offset, err = readString(data, offset)
	if err != nil {
		return
	}

	key, _, err = readString(data, offset)
	return
}

func EncodeDeletePayload(token, bucket, key string) []byte {
	return EncodeGetPayload(token, bucket, key)
}

func DecodeDeletePayload(data []byte) (token, bucket, key string, err error) {
	return DecodeGetPayload(data)
}

// Format: [KeyLen(2)][Key][TTL(8)][CreatedAt(8)][ExpiresAt(8)][SingleRead(1)][ValueLen(4)][Value]
func EncodeValueResponse(key string, ttl int64, createdAt, expiresAt int64, singleRead bool, value []byte) []byte {
	size := 2 + len(key) + 8 + 8 + 8 + 1 + 4 + len(value)
	buf := make([]byte, size)

	offset := 0
	offset = writeString(buf, offset, key)

	binary.BigEndian.PutUint64(buf[offset:], uint64(ttl))
	offset += 8

	binary.BigEndian.PutUint64(buf[offset:], uint64(createdAt))
	offset += 8

	binary.BigEndian.PutUint64(buf[offset:], uint64(expiresAt))
	offset += 8

	if singleRead {
		buf[offset] = 1
	}
	offset++

	binary.BigEndian.PutUint32(buf[offset:], uint32(len(value)))
	offset += 4
	copy(buf[offset:], value)

	return buf
}
