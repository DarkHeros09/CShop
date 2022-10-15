// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	CreateAddress(ctx context.Context, arg CreateAddressParams) (Address, error)
	CreateAdmin(ctx context.Context, arg CreateAdminParams) (Admin, error)
	CreateAdminType(ctx context.Context, adminType string) (AdminType, error)
	CreateCategoryPromotion(ctx context.Context, arg CreateCategoryPromotionParams) (CategoryPromotion, error)
	CreateOrderStatus(ctx context.Context, status string) (OrderStatus, error)
	CreatePaymentMethod(ctx context.Context, arg CreatePaymentMethodParams) (PaymentMethod, error)
	CreatePaymentType(ctx context.Context, value string) (PaymentType, error)
	CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error)
	CreateProductCategory(ctx context.Context, arg CreateProductCategoryParams) (ProductCategory, error)
	CreateProductConfiguration(ctx context.Context, arg CreateProductConfigurationParams) (ProductConfiguration, error)
	CreateProductItem(ctx context.Context, arg CreateProductItemParams) (ProductItem, error)
	CreateProductPromotion(ctx context.Context, arg CreateProductPromotionParams) (ProductPromotion, error)
	CreatePromotion(ctx context.Context, arg CreatePromotionParams) (Promotion, error)
	CreateShippingMethod(ctx context.Context, arg CreateShippingMethodParams) (ShippingMethod, error)
	CreateShopOrder(ctx context.Context, arg CreateShopOrderParams) (ShopOrder, error)
	CreateShopOrderItem(ctx context.Context, arg CreateShopOrderItemParams) (ShopOrderItem, error)
	CreateShoppingCart(ctx context.Context, userID int64) (ShoppingCart, error)
	CreateShoppingCartItem(ctx context.Context, arg CreateShoppingCartItemParams) (ShoppingCartItem, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	CreateUserAddress(ctx context.Context, arg CreateUserAddressParams) (UserAddress, error)
	CreateUserReview(ctx context.Context, arg CreateUserReviewParams) (UserReview, error)
	CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error)
	CreateVariation(ctx context.Context, arg CreateVariationParams) (Variation, error)
	CreateVariationOption(ctx context.Context, arg CreateVariationOptionParams) (VariationOption, error)
	CreateWishList(ctx context.Context, userID int64) (WishList, error)
	CreateWishListItem(ctx context.Context, arg CreateWishListItemParams) (WishListItem, error)
	DeleteAddress(ctx context.Context, id int64) error
	DeleteAdmin(ctx context.Context, id int64) error
	DeleteAdminTypeByID(ctx context.Context, id int64) error
	DeleteAdminTypeByType(ctx context.Context, adminType string) error
	DeleteCategoryPromotion(ctx context.Context, categoryID int64) error
	DeleteOrderStatus(ctx context.Context, id int64) error
	DeletePaymentMethod(ctx context.Context, id int64) error
	DeletePaymentType(ctx context.Context, id int64) error
	DeleteProduct(ctx context.Context, id int64) error
	DeleteProductCategory(ctx context.Context, arg DeleteProductCategoryParams) error
	DeleteProductConfiguration(ctx context.Context, productItemID int64) error
	DeleteProductItem(ctx context.Context, id int64) error
	DeleteProductPromotion(ctx context.Context, productID int64) error
	DeletePromotion(ctx context.Context, id int64) error
	DeleteShippingMethod(ctx context.Context, id int64) error
	DeleteShopOrder(ctx context.Context, id int64) error
	DeleteShopOrderItem(ctx context.Context, id int64) error
	DeleteShoppingCart(ctx context.Context, id int64) error
	DeleteShoppingCartItem(ctx context.Context, id int64) error
	DeleteUser(ctx context.Context, id int64) error
	DeleteUserAddress(ctx context.Context, arg DeleteUserAddressParams) error
	DeleteUserReview(ctx context.Context, id int64) error
	DeleteVariation(ctx context.Context, id int64) error
	DeleteVariationOption(ctx context.Context, id int64) error
	DeleteWishList(ctx context.Context, id int64) error
	DeleteWishListItem(ctx context.Context, id int64) error
	GetAddress(ctx context.Context, id int64) (Address, error)
	GetAddressByCity(ctx context.Context, city string) (Address, error)
	GetAdmin(ctx context.Context, id int64) (Admin, error)
	GetAdminByEmail(ctx context.Context, email string) (Admin, error)
	GetAdminType(ctx context.Context, id int64) (AdminType, error)
	GetCategoryPromotion(ctx context.Context, categoryID int64) (CategoryPromotion, error)
	GetOrderStatus(ctx context.Context, id int64) (OrderStatus, error)
	GetPaymentMethod(ctx context.Context, id int64) (PaymentMethod, error)
	GetPaymentType(ctx context.Context, id int64) (PaymentType, error)
	GetProduct(ctx context.Context, id int64) (Product, error)
	GetProductCategory(ctx context.Context, id int64) (ProductCategory, error)
	GetProductCategoryByParent(ctx context.Context, arg GetProductCategoryByParentParams) (ProductCategory, error)
	GetProductConfiguration(ctx context.Context, productItemID int64) (ProductConfiguration, error)
	GetProductItem(ctx context.Context, id int64) (ProductItem, error)
	GetProductItemForUpdate(ctx context.Context, id int64) (ProductItem, error)
	GetProductPromotion(ctx context.Context, productID int64) (ProductPromotion, error)
	GetPromotion(ctx context.Context, id int64) (Promotion, error)
	GetShippingMethod(ctx context.Context, id int64) (ShippingMethod, error)
	GetShopOrder(ctx context.Context, id int64) (ShopOrder, error)
	GetShopOrderItem(ctx context.Context, id int64) (ShopOrderItem, error)
	GetShoppingCart(ctx context.Context, id int64) (ShoppingCart, error)
	GetShoppingCartItem(ctx context.Context, id int64) (ShoppingCartItem, error)
	GetUser(ctx context.Context, id int64) (User, error)
	GetUserAddress(ctx context.Context, arg GetUserAddressParams) (UserAddress, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserReview(ctx context.Context, id int64) (UserReview, error)
	GetUserSession(ctx context.Context, id uuid.UUID) (UserSession, error)
	GetVariation(ctx context.Context, id int64) (Variation, error)
	GetVariationOption(ctx context.Context, id int64) (VariationOption, error)
	GetWishList(ctx context.Context, id int64) (WishList, error)
	GetWishListItem(ctx context.Context, id int64) (WishListItem, error)
	ListAddressesByCity(ctx context.Context, arg ListAddressesByCityParams) ([]Address, error)
	ListAdminTypes(ctx context.Context, arg ListAdminTypesParams) ([]AdminType, error)
	ListAdmins(ctx context.Context, arg ListAdminsParams) ([]Admin, error)
	ListCategoryPromotions(ctx context.Context, arg ListCategoryPromotionsParams) ([]CategoryPromotion, error)
	ListOrderStatuses(ctx context.Context, arg ListOrderStatusesParams) ([]OrderStatus, error)
	ListPaymentMethods(ctx context.Context, arg ListPaymentMethodsParams) ([]PaymentMethod, error)
	ListPaymentTypes(ctx context.Context, arg ListPaymentTypesParams) ([]PaymentType, error)
	ListProductCategories(ctx context.Context, arg ListProductCategoriesParams) ([]ProductCategory, error)
	ListProductCategoriesByParent(ctx context.Context, arg ListProductCategoriesByParentParams) ([]ProductCategory, error)
	ListProductConfigurations(ctx context.Context, arg ListProductConfigurationsParams) ([]ProductConfiguration, error)
	ListProductItems(ctx context.Context, arg ListProductItemsParams) ([]ProductItem, error)
	ListProductPromotions(ctx context.Context, arg ListProductPromotionsParams) ([]ProductPromotion, error)
	ListProducts(ctx context.Context, arg ListProductsParams) ([]Product, error)
	ListPromotions(ctx context.Context, arg ListPromotionsParams) ([]Promotion, error)
	ListShippingMethods(ctx context.Context, arg ListShippingMethodsParams) ([]ShippingMethod, error)
	ListShopOrderItems(ctx context.Context, arg ListShopOrderItemsParams) ([]ShopOrderItem, error)
	ListShopOrderItemsByOrderID(ctx context.Context, orderID int64) ([]ShopOrderItem, error)
	ListShopOrders(ctx context.Context, arg ListShopOrdersParams) ([]ShopOrder, error)
	ListShoppingCartItems(ctx context.Context, arg ListShoppingCartItemsParams) ([]ShoppingCartItem, error)
	ListShoppingCartItemsByCartID(ctx context.Context, shoppingCartID int64) ([]ShoppingCartItem, error)
	ListShoppingCarts(ctx context.Context, arg ListShoppingCartsParams) ([]ShoppingCart, error)
	ListUserAddresses(ctx context.Context, arg ListUserAddressesParams) ([]UserAddress, error)
	ListUserReviews(ctx context.Context, arg ListUserReviewsParams) ([]UserReview, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
	ListVariationOptions(ctx context.Context, arg ListVariationOptionsParams) ([]VariationOption, error)
	ListVariations(ctx context.Context, arg ListVariationsParams) ([]Variation, error)
	ListWishListItems(ctx context.Context, arg ListWishListItemsParams) ([]WishListItem, error)
	ListWishListItemsByCartID(ctx context.Context, wishListID int64) ([]WishListItem, error)
	ListWishLists(ctx context.Context, arg ListWishListsParams) ([]WishList, error)
	UpdateAddress(ctx context.Context, arg UpdateAddressParams) (Address, error)
	UpdateAdmin(ctx context.Context, arg UpdateAdminParams) (Admin, error)
	UpdateAdminType(ctx context.Context, arg UpdateAdminTypeParams) (AdminType, error)
	UpdateCategoryPromotion(ctx context.Context, arg UpdateCategoryPromotionParams) (CategoryPromotion, error)
	UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) (OrderStatus, error)
	UpdatePaymentMethod(ctx context.Context, arg UpdatePaymentMethodParams) (PaymentMethod, error)
	UpdatePaymentType(ctx context.Context, arg UpdatePaymentTypeParams) (PaymentType, error)
	UpdateProduct(ctx context.Context, arg UpdateProductParams) (Product, error)
	UpdateProductCategory(ctx context.Context, arg UpdateProductCategoryParams) (ProductCategory, error)
	UpdateProductConfiguration(ctx context.Context, arg UpdateProductConfigurationParams) (ProductConfiguration, error)
	UpdateProductItem(ctx context.Context, arg UpdateProductItemParams) (ProductItem, error)
	UpdateProductPromotion(ctx context.Context, arg UpdateProductPromotionParams) (ProductPromotion, error)
	UpdatePromotion(ctx context.Context, arg UpdatePromotionParams) (Promotion, error)
	UpdateShippingMethod(ctx context.Context, arg UpdateShippingMethodParams) (ShippingMethod, error)
	UpdateShopOrder(ctx context.Context, arg UpdateShopOrderParams) (ShopOrder, error)
	UpdateShopOrderItem(ctx context.Context, arg UpdateShopOrderItemParams) (ShopOrderItem, error)
	UpdateShoppingCart(ctx context.Context, arg UpdateShoppingCartParams) (ShoppingCart, error)
	UpdateShoppingCartItem(ctx context.Context, arg UpdateShoppingCartItemParams) (ShoppingCartItem, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserAddress(ctx context.Context, arg UpdateUserAddressParams) (UserAddress, error)
	UpdateUserReview(ctx context.Context, arg UpdateUserReviewParams) (UserReview, error)
	UpdateVariation(ctx context.Context, arg UpdateVariationParams) (Variation, error)
	UpdateVariationOption(ctx context.Context, arg UpdateVariationOptionParams) (VariationOption, error)
	UpdateWishList(ctx context.Context, arg UpdateWishListParams) (WishList, error)
	UpdateWishListItem(ctx context.Context, arg UpdateWishListItemParams) (WishListItem, error)
}

var _ Querier = (*Queries)(nil)
