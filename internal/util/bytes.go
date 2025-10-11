package util

import "unsafe"

// StringToBytes converts string to []byte without allocation (zero-copy)
// WARNING: The returned slice must NOT be modified and shares memory with the string
// Use this only when you need read-only access to the bytes
func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts []byte to string without allocation (zero-copy)
// The resulting string shares the underlying byte array
// Safe for read-only operations
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
