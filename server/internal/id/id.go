package id

import (
	"crypto/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

var counter atomic.Uint32

func New() string {
	now := time.Now().UnixMilli()

	var b strings.Builder
	b.WriteByte('c')
	b.WriteString(strconv.FormatUint(uint64(now), 36))
	b.WriteString(fixedBase36(uint64(counter.Add(1)%1679616), 4))
	b.WriteString(randomString(8))
	return b.String()
}

func fixedBase36(value uint64, width int) string {
	encoded := strconv.FormatUint(value, 36)
	if len(encoded) >= width {
		return encoded[len(encoded)-width:]
	}
	return strings.Repeat("0", width-len(encoded)) + encoded
}

func randomString(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("0", length)
	}

	var b strings.Builder
	b.Grow(length)
	for _, value := range buf {
		b.WriteByte(alphabet[int(value)%len(alphabet)])
	}
	return b.String()
}
