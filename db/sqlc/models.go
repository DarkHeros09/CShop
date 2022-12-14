// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null"
)

type Address struct {
	ID          int64     `json:"id"`
	AddressLine string    `json:"address_line"`
	Region      string    `json:"region"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Admin struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Active    bool      `json:"active"`
	TypeID    int64     `json:"type_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login"`
}

type AdminType struct {
	ID        int64     `json:"id"`
	AdminType string    `json:"admin_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryPromotion struct {
	CategoryID  int64 `json:"category_id"`
	PromotionID int64 `json:"promotion_id"`
	// default is false
	Active bool `json:"active"`
}

type OrderStatus struct {
	ID int64 `json:"id"`
	// values like ordered, proccessed and delivered
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaymentMethod struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	PaymentTypeID int64  `json:"payment_type_id"`
	Provider      string `json:"provider"`
	IsDefault     bool   `json:"is_default"`
}

type PaymentType struct {
	ID int64 `json:"id"`
	// for companies payment system like BCD
	Value string `json:"value"`
}

type Product struct {
	ID           int64  `json:"id"`
	CategoryID   int64  `json:"category_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ProductImage string `json:"product_image"`
	// default is false
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductCategory struct {
	ID               int64    `json:"id"`
	ParentCategoryID null.Int `json:"parent_category_id"`
	CategoryName     string   `json:"category_name"`
}

type ProductConfiguration struct {
	ProductItemID     int64 `json:"product_item_id"`
	VariationOptionID int64 `json:"variation_option_id"`
}

type ProductItem struct {
	ID         int64 `json:"id"`
	ProductID  int64 `json:"product_id"`
	ProductSku int64 `json:"product_sku"`
	QtyInStock int32 `json:"qty_in_stock"`
	// may be used to show different images than original
	ProductImage string `json:"product_image"`
	Price        string `json:"price"`
	// default is false
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductPromotion struct {
	ProductID   int64 `json:"product_id"`
	PromotionID int64 `json:"promotion_id"`
	// default is false
	Active bool `json:"active"`
}

type Promotion struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DiscountRate int64  `json:"discount_rate"`
	// default is false
	Active    bool      `json:"active"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type ShippingMethod struct {
	ID int64 `json:"id"`
	// values like normal, or free
	Name  string `json:"name"`
	Price string `json:"price"`
}

type ShopOrder struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"user_id"`
	PaymentMethodID   int64     `json:"payment_method_id"`
	ShippingAddressID int64     `json:"shipping_address_id"`
	OrderTotal        string    `json:"order_total"`
	ShippingMethodID  int64     `json:"shipping_method_id"`
	OrderStatusID     int64     `json:"order_status_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ShopOrderItem struct {
	ID            int64 `json:"id"`
	ProductItemID int64 `json:"product_item_id"`
	OrderID       int64 `json:"order_id"`
	Quantity      int32 `json:"quantity"`
	// price of product when ordered
	Price     string    `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ShoppingCart struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ShoppingCartItem struct {
	ID             int64     `json:"id"`
	ShoppingCartID int64     `json:"shopping_cart_id"`
	ProductItemID  int64     `json:"product_item_id"`
	Qty            int32     `json:"qty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	Telephone      int32     `json:"telephone"`
	DefaultPayment null.Int  `json:"default_payment"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserAddress struct {
	UserID         int64     `json:"user_id"`
	AddressID      int64     `json:"address_id"`
	DefaultAddress null.Int  `json:"default_address"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserReview struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	OrderedProductID int64     `json:"ordered_product_id"`
	RatingValue      int32     `json:"rating_value"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserSession struct {
	ID           uuid.UUID `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Variation struct {
	ID         int64 `json:"id"`
	CategoryID int64 `json:"category_id"`
	// variation names like color, and size
	Name string `json:"name"`
}

type VariationOption struct {
	ID          int64 `json:"id"`
	VariationID int64 `json:"variation_id"`
	// variation values like Red, ans Size XL
	Value string `json:"value"`
}

type WishList struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WishListItem struct {
	ID            int64     `json:"id"`
	WishListID    int64     `json:"wish_list_id"`
	ProductItemID int64     `json:"product_item_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
