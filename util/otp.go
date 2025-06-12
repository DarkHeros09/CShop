package util

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateOTP() string {

	otpChars := "0123456789"
	otpCharsLength := len(otpChars)
	var otp strings.Builder

	// Pre-allocate space for a 6-character OTP
	otp.Grow(6)

	// Generate OTP
	for i := 0; i < 6; i++ {
		otp.WriteByte(otpChars[rand.Intn(otpCharsLength)])
	}

	return otp.String()

}
