package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Random generate a random interger between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// Random generate a random interger between min and max
func RandomDecimal(min, max float64) string {
	rF64 := min + rand.Float64()*(max-min)
	return decimal.NewFromFloat(rF64).StringFixedBank(2)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomOwner generates a random user name
func RandomUser() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}

/*
RandBool
    This function returns a random boolean value based on the current time
*/
// func RandomBool() bool {
// 	rand.Seed(time.Now().UnixNano())
// 	return rand.Intn(2) == 1
// }

// RandomBool generates a random boolean
func RandomBool() bool {
	bool := []bool{true, false}
	n := len(bool)
	return bool[rand.Intn(n)]
}
