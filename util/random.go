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

// Random generate a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// Random generate a random decimal between min and max
func RandomDecimalString(min, max float64) string {
	rF64 := min + rand.Float64()*(max-min)
	return decimal.NewFromFloat(rF64).Abs().StringFixed(2)
}

func RandomDecimal(min, max float64) decimal.Decimal {
	rF64 := min + rand.Float64()*(max-min)
	return decimal.NewFromFloat(rF64).Abs().Round(2)
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
	return RandomInt(1, 1000)
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}

// func RandomURL() string {
// 	return fmt.Sprintf("https://%s.com/%s", RandomString(6), RandomString(5))
// }

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

func RandomStartDate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2010, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0).Local().UTC()
}

func RandomEndDate() time.Time {
	min := time.Date(2010, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0).Local().UTC()
}

func RandomColor() string {
	colors := []string{"red", "black", "blue", "brown", "yellow", "green", "white"}
	n := len(colors)
	return colors[rand.Intn(n)]
}

func RandomSize() string {
	sizes := []string{"S", "M", "L", "XL", "XXL", "XXXL"}
	n := len(sizes)
	return sizes[rand.Intn(n)]
}

func RandomPromotionURL() string {
	urls := []string{
		"https://i.pinimg.com/originals/a7/ce/56/a7ce56e472d4c16f4ba47ef6ba4bea18.png",
		"https://genxfinance.com/genesis/wp-content/uploads/2010/12/clothing-sale.jpg",
		"https://i.pinimg.com/originals/19/0c/d9/190cd9d43f4477db4b422dc9e3b6e347.jpg",
		"https://cdn.dribbble.com/users/2548571/screenshots/5408903/sale_offer-01-01.jpg",
		"https://bestofyou.gr/sites/default/files/styles/article_inner/public/2022-11/iStock-1085149318.jpg",
		"https://1.bp.blogspot.com/_hXYbJNRnLrI/Swo7KmZyxlI/AAAAAAAAACo/nVtdePZYJAg/s1600/hp_us.jpg",
		"https://img.freepik.com/free-vector/abstract-sale-promotion-banner-template_23-2148217955.jpg",
	}
	n := len(urls)
	return urls[rand.Intn(n)]
}

// RandomURL generates a random URL for products
func RandomURL() string {
	urls := []string{
		"https://plus.unsplash.com/premium_photo-1670430782104-e4dca78d746b?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE0fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://plus.unsplash.com/premium_photo-1670787505396-31741dac157c?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE1fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670948516733-0220701ea0de?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE2fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670513756456-6d51163ff25c?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDEzfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670718221502-42bee4ee96c7?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDExfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670616440058-927ed245e905?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDIwfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670718089430-d75ba6c1a194?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDEwfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://plus.unsplash.com/premium_photo-1670264592766-0ae99e33aacd?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDF8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1652087069456-06fb70235fa2?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDl8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1671094362390-02b76f5e0556?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDV8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://plus.unsplash.com/premium_photo-1670930887547-518452a967fb?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDR8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1667655866927-c334f8db27ae?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDZ8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1670896329450-066bc7c029f5?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE3fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670911170630-c9acad2e9da7?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDJ8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1669383488518-3f367058d9db?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDEyfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670531910262-5ddb5ad666ea?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE5fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://images.unsplash.com/photo-1670928591025-7d7304e77e0f?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDh8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1670718089430-d75ba6c1a194?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxwaG90by1vZi10aGUtZGF5fHx8fGVufDB8fHx8&dpr=1&auto=format%2Ccompress&fit=crop&w=1999&h=594",
		"https://images.unsplash.com/photo-1670718089430-d75ba6c1a194?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxwaG90by1vZi10aGUtZGF5fHx8fGVufDB8fHx8&auto=format%2Ccompress&fit=crop&w=1000&h=1000",
		"https://images.unsplash.com/photo-1670938258821-2956d4ce9c9b?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDd8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1671099484139-b4674a9bf986?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDN8UzRNS0xBc0JCNzR8fGVufDB8fHx8&w=1000&q=80",
		"https://images.unsplash.com/photo-1644483878407-f05ff59c8a38?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE4fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&w=1000&q=80",
		"https://plus.unsplash.com/premium_photo-1670430782104-e4dca78d746b?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE0fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670948516733-0220701ea0de?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE2fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://plus.unsplash.com/premium_photo-1670787505396-31741dac157c?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE1fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670513756456-6d51163ff25c?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDEzfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670616440058-927ed245e905?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDIwfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670718089430-d75ba6c1a194?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDEwfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670718221502-42bee4ee96c7?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDExfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1652087069456-06fb70235fa2?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDl8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://plus.unsplash.com/premium_photo-1670264592766-0ae99e33aacd?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDF8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1671094362390-02b76f5e0556?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDV8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://plus.unsplash.com/premium_photo-1670930887547-518452a967fb?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDR8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1667655866927-c334f8db27ae?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDZ8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670896329450-066bc7c029f5?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE3fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670911170630-c9acad2e9da7?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDJ8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1669383488518-3f367058d9db?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDEyfFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670531910262-5ddb5ad666ea?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE5fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670928591025-7d7304e77e0f?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDh8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1670938258821-2956d4ce9c9b?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDd8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1644483878407-f05ff59c8a38?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDE4fFM0TUtMQXNCQjc0fHxlbnwwfHx8fA%3D%3D&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/photo-1671099484139-b4674a9bf986?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHx0b3BpYy1mZWVkfDN8UzRNS0xBc0JCNzR8fGVufDB8fHx8&auto=format&fit=crop&w=500&q=60",
		"https://images.unsplash.com/profile-1651045251984-265ddaefccbfimage?dpr=1&auto=format&fit=crop&w=32&h=32&q=60&crop=faces&bg=fff",
		"https://n.nordstrommedia.com/id/sr3/0c67d797-45bd-4110-bf56-790ca959246a.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/f7db7fc1-384d-4c59-b9ea-811bb81d462f.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/25521384-f631-496f-993b-877a391091d4.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/089dffc5-9f33-40c5-b804-6d2454bb356d.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/2df3be13-5d17-40da-8f03-dea56c75b285.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/d6cbb0e1-be28-470b-88ed-f1c2d924c357.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/cad73f73-670d-4942-8679-1c409151d05f.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/5c713ef3-579a-4e12-8ac0-e30b74fecd73.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/95753e60-816e-4404-9294-56b1e218db1c.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/f35dc5ee-69af-4aff-a109-cd30c3c92aaa.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/807c0748-4960-4ab2-a42e-1994d44bd990.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/fb0055d6-f7ed-4028-b321-e1380c467871.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/e306b441-06a6-415c-88e2-f9b172575452.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/09770a38-1249-4cba-bdad-cc65019b94fb.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/1058f0a0-6d0c-4d07-bed1-820308135934.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/2ff59b4a-0ef0-4088-b7ad-0d646d113c44.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/e24fbedd-c28b-4ece-b6f4-9fe8b755deb2.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/ea5da33a-0422-47f5-87f1-5c91f933909b.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/3bbdf111-f968-42c7-ae17-9b73ee40f3a1.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/5e6d4106-ece9-4d51-babe-7b232ecf67f1.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/944b92cd-630f-4bb2-8e4b-1dbcb49f5856.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/89cfb3e2-2ab4-47f1-b601-30105336ff39.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/cc4e0c8d-ba75-44f9-879c-c5adb86a96d4.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/28c6dce9-58da-4762-9fdb-c1bab4ce37be.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/1fd2d60e-5d5c-4d75-ac53-caebf65e091f.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/d2c08fe3-cbfa-4eb4-9c7f-053159f45540.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/badb06cf-dc31-4e56-b303-228751d922d5.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/71c10b6f-0970-4fff-abac-6e24aa1f0a6b.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/10e6b6ab-0858-47ba-be77-5d089b39ca3f.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/17087eb1-4591-4f4f-9a63-e2bae833a405.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/600e4868-15ed-4776-9a47-95674925a221.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/11193de6-5ee2-4ec8-a078-4e15930a070c.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/d556e3cb-36ff-4bd2-bd0e-032929f1aca8.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/88432917-31f7-4969-99e5-0a80c51d033f.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/2755385d-c63c-47ec-a4d6-f1538e97587a.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/d072806d-7177-4f77-b419-b947a410b751.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/eb2cdfa9-a25d-46b6-a2e6-3bf979f10b82.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/2e0ef053-ec33-461a-9077-1d1c997baa0e.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/66ed332b-d2da-49f7-a7d0-6cefa7b23f4b.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/687f3cc0-9b88-4985-beb2-21d101aab77b.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/377c2687-f48e-4fb7-abed-c7c52bad0aad.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/5c51a7e6-0d24-4df7-9d4b-9a8cd9f4b0c1.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/a04f5b18-2d2d-45fc-af6f-7f3c63ae86e1.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/d4a29c1c-0506-4fdc-9c10-d5b7faa9d87f.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/51c31109-82ef-4d71-9720-5462c660c159.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/b6d82427-1edb-4386-ba33-e2190a64b472.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/0329cf64-f18b-4927-984e-4d0572d8e0bc.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/476f80eb-bd62-4816-a1a1-4ab4e629f4d7.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/5e31930d-7e1d-4fc1-976b-af5e495a417a.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/5c0961ed-3c96-4a6d-864a-832889414c89.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/162d9b3b-e9bf-49c0-b370-ee74b1951268.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/20fc3408-a828-4f19-808a-267532826dfa.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/e84ef2eb-4b7f-4b21-8d20-421c3fc0d7e6.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/a85c4ff0-a127-48f9-ae1a-a89c86efddc3.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/3a534f63-f7f0-462c-bb6f-043ab40c6774.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/4d2ac452-f2ea-4870-81f5-58fb7d4eae93.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/cd250314-616a-43e5-9e4e-f9bfbc5288e6.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/a18950ed-268c-4dc2-b671-975a5893d094.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/9045ccdf-c3de-4dac-9061-6e1fb8cceb90.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/cc8dec3e-b657-4970-b818-bee1b5a58959.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/fe6b6e67-25fe-4122-8bac-5a0ac90bf17c.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/c3be7eea-172e-418e-8423-72544ec747dc.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/dac9080b-d968-421b-9559-1e669877d6ec.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/bb6a3d8e-a401-44fb-bbab-5185d44b9124.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/ac070ed5-30db-4ce6-aa85-3e96dbe326a6.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/cbc6b3bb-3e6e-4c1b-ade4-747f2f518d5e.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/54e720e8-5344-4287-8881-b2cec4b06826.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/3a4f1350-f199-4bfe-9774-e26b24794cc8.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/892ad7b3-d789-42c4-b7df-797a4adc2a65.jpeg?h=365&w=240&dpr=2",
		"https://n.nordstrommedia.com/id/sr3/b73dac45-12e9-4fca-a876-d1fbe5c42be1.jpeg?h=365&w=240&dpr=2",
	}
	n := len(urls)
	return urls[rand.Intn(n)]
}
