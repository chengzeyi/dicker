package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenRandStrBytes(n int) string {
	const letterBytes = "1234567890"
	bSlice := make([]byte, n)
	for i := range bSlice {
		bSlice[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(bSlice)
}
