package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

func Base64encode(s []byte) string {
	return base64.StdEncoding.EncodeToString(s)
}

func Base64decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func MD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
