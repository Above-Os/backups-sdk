package util

import (
	"fmt"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	var a = "2025-02-28 16:00:00"
	var t1, _ = time.Parse("2006-01-02 15:04:05", a)
	// 1740758400000
	fmt.Println(t1.UnixMilli())
}
