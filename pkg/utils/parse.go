package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func FormatBytes(bytes uint64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
		TB = 1 << 40
	)

	var result string
	switch {
	case bytes >= TB:
		result = fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		result = fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		result = fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		result = fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		result = fmt.Sprintf("%d Byte", bytes)
	}

	return result
}

// ToJSON returns a json string
func ToJSON(v any) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		panic(err)
	}
	return buf.String()
}

func DefaultValue(defaultValue string, newValue string) string {
	if newValue == "" {
		return defaultValue
	}
	return newValue
}

func TrimRight(content string, suffix string) string {
	return strings.TrimRight(content, suffix)
}

func ParseEncrypt(name string, encrypt bool) string {
	if encrypt {
		return MD5(name)
	}
	return name
}
