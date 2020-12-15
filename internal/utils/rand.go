package utils

import (
	"math/rand"
	"time"
)

// RandomString generate random string with lower case alphabets and digits
func RandomString(length int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz1234567890"
	str := make([]byte, length)

	for i := range str {
		str[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(str)
}
