package internal

import (
	"crypto/md5"
	"encoding/hex"
)

// GetBytesMD5 returns the MD5 hash of the given data as a hex string.
func GetBytesMD5(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// GetStringMD5 returns the MD5 hash of the given string as a hex string.
func GetStringMD5(data string) string {
	return GetBytesMD5([]byte(data))
}
