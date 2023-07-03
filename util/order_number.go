package util

import (
	"math/rand"
	"time"
)

const (
	orderNumberLength  = 10
	orderNumberCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateTrackNumber() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	orderNumber := make([]byte, orderNumberLength)
	charsetLength := len(orderNumberCharset)

	for i := 0; i < orderNumberLength; i++ {
		orderNumber[i] = orderNumberCharset[rand.Intn(charsetLength)]
	}

	return string(orderNumber)
}
