package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"key-value-store/internal/util"
	"time"
)

type TokenManager struct {
	secretKey []byte
}

func NewTokenManager(secretKey []byte) *TokenManager {
	return &TokenManager{
		secretKey: secretKey,
	}
}

// GenerateToken creates a token with expiration
// Token format: base64url(bucket_name + "." + expiry_unix + "." + signature)
// ttlSeconds: token validity duration in seconds (0 = no expiration)
func (tm *TokenManager) GenerateToken(bucketName string, ttlSeconds int64) string {
	nameBytes := util.StringToBytes(bucketName)

	var expiryUnix int64
	if ttlSeconds > 0 {
		expiryUnix = time.Now().Unix() + ttlSeconds
	} else {
		expiryUnix = 0 // No expiration
	}

	// Convert expiry to 8-byte representation
	expiryBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expiryBytes, uint64(expiryUnix))

	// Build data to sign: bucket_name.expiry
	dataToSign := make([]byte, len(nameBytes)+1+8)
	pos := 0
	pos += copy(dataToSign[pos:], nameBytes)
	dataToSign[pos] = '.'
	pos++
	copy(dataToSign[pos:], expiryBytes)

	signature := tm.sign(dataToSign)

	// Build final token: bucket_name.expiry.signature
	tokenLen := len(nameBytes) + 1 + 8 + 1 + len(signature)
	token := make([]byte, tokenLen)

	pos = 0
	pos += copy(token[pos:], nameBytes)
	token[pos] = '.'
	pos++
	pos += copy(token[pos:], expiryBytes)
	token[pos] = '.'
	pos++
	copy(token[pos:], signature)

	encoded := base64.RawURLEncoding.EncodeToString(token)
	return encoded
}

// ValidateToken verifies token signature, expiration, and bucket name match
// Returns true if token is valid, not expired, and matches expected bucket
func (tm *TokenManager) ValidateToken(tokenStr, expected string) bool {
	if tokenStr == "" || expected == "" {
		return false
	}

	tokenBytes, err := base64.RawURLEncoding.DecodeString(tokenStr)
	if err != nil {
		return false
	}

	// Find first separator (after bucket name)
	firstSep := -1
	for i := 0; i < len(tokenBytes); i++ {
		if tokenBytes[i] == '.' {
			firstSep = i
			break
		}
	}

	if firstSep == -1 || firstSep == 0 {
		return false
	}

	// Find second separator (after expiry)
	secondSep := -1
	for i := firstSep + 1; i < len(tokenBytes); i++ {
		if tokenBytes[i] == '.' {
			secondSep = i
			break
		}
	}

	if secondSep == -1 || secondSep-firstSep != 9 { // 1 (dot) + 8 (expiry bytes)
		return false
	}

	nameBytes := tokenBytes[:firstSep]
	expiryBytes := tokenBytes[firstSep+1 : secondSep]
	providedSig := tokenBytes[secondSep+1:]

	// Zero-copy string comparison for bucket name
	bucketName := util.BytesToString(nameBytes)
	if bucketName != expected {
		return false
	}

	// Check expiration (fast integer comparison)
	expiryUnix := int64(binary.BigEndian.Uint64(expiryBytes))
	if expiryUnix > 0 && time.Now().Unix() > expiryUnix {
		return false // Token expired
	}

	// Build data that was signed
	dataToSign := make([]byte, len(nameBytes)+1+8)
	pos := 0
	pos += copy(dataToSign[pos:], nameBytes)
	dataToSign[pos] = '.'
	pos++
	copy(dataToSign[pos:], expiryBytes)

	// Constant-time signature verification
	expectedSig := tm.sign(dataToSign)
	return hmac.Equal(providedSig, expectedSig)
}

func (tm *TokenManager) sign(data []byte) []byte {
	mac := hmac.New(sha256.New, tm.secretKey)
	mac.Write(data)
	return mac.Sum(nil)
}
