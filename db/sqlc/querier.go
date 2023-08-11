// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/guregu/null"
)

type Querier interface {
	CheckUserAddressDefaultAddress(ctx context.Context, userID int64) (int64, error)
	CreateAddress(ctx context.Context, arg CreateAddressParams) (Address, error)
	CreateAdmin(ctx context.Context, arg CreateAdminParams) (Admin, error)
	CreateAdminType(ctx context.Context, adminType string) (AdminType, error)
	CreateBrandPromotion(ctx context.Context, arg CreateBrandPromotionParams) (BrandPromotion, error)
	CreateCategoryPromotion(ctx context.Context, arg CreateCategoryPromotionParams) (CategoryPromotion, error)
	CreateHomePageTextBanner(ctx context.Context, arg CreateHomePageTextBannerParams) (HomePageTextBanner, error)
	CreateOrderStatus(ctx context.Context, status string) (OrderStatus, error)
	CreatePaymentMethod(ctx context.Context, arg CreatePaymentMethodParams) (PaymentMethod, error)
	CreatePaymentType(ctx context.Context, value string) (PaymentType, error)
	CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error)
	CreateProductBrand(ctx context.Context, arg CreateProductBrandParams) (ProductBrand, error)
	CreateProductCategory(ctx context.Context, arg CreateProductCategoryParams) (ProductCategory, error)
	CreateProductColor(ctx context.Context, colorValue string) (ProductColor, error)
	CreateProductConfiguration(ctx context.Context, arg CreateProductConfigurationParams) (ProductConfiguration, error)
	CreateProductImage(ctx context.Context, arg CreateProductImageParams) (ProductImage, error)
	CreateProductItem(ctx context.Context, arg CreateProductItemParams) (ProductItem, error)
	CreateProductPromotion(ctx context.Context, arg CreateProductPromotionParams) (ProductPromotion, error)
	CreateProductSize(ctx context.Context, sizeValue string) (ProductSize, error)
	CreatePromotion(ctx context.Context, arg CreatePromotionParams) (Promotion, error)
	CreateShippingMethod(ctx context.Context, arg CreateShippingMethodParams) (ShippingMethod, error)
	CreateShopOrder(ctx context.Context, arg CreateShopOrderParams) (ShopOrder, error)
	CreateShopOrderItem(ctx context.Context, arg CreateShopOrderItemParams) (ShopOrderItem, error)
	CreateShoppingCart(ctx context.Context, userID int64) (ShoppingCart, error)
	CreateShoppingCartItem(ctx context.Context, arg CreateShoppingCartItemParams) (ShoppingCartItem, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	CreateUserAddress(ctx context.Context, arg CreateUserAddressParams) (UserAddress, error)
	CreateUserAddressWithAddress(ctx context.Context, arg CreateUserAddressWithAddressParams) (CreateUserAddressWithAddressRow, error)
	CreateUserReview(ctx context.Context, arg CreateUserReviewParams) (UserReview, error)
	CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error)
	CreateUserWithCartAndWishList(ctx context.Context, arg CreateUserWithCartAndWishListParams) (CreateUserWithCartAndWishListRow, error)
	CreateVariation(ctx context.Context, arg CreateVariationParams) (Variation, error)
	CreateVariationOption(ctx context.Context, arg CreateVariationOptionParams) (VariationOption, error)
	CreateWishList(ctx context.Context, userID int64) (WishList, error)
	CreateWishListItem(ctx context.Context, arg CreateWishListItemParams) (WishListItem, error)
	DeleteAddress(ctx context.Context, id int64) error
	DeleteAdmin(ctx context.Context, id int64) error
	DeleteAdminTypeByID(ctx context.Context, id int64) error
	DeleteAdminTypeByType(ctx context.Context, adminType string) error
	DeleteBrandPromotion(ctx context.Context, arg DeleteBrandPromotionParams) error
	DeleteCategoryPromotion(ctx context.Context, arg DeleteCategoryPromotionParams) error
	DeleteHomePageTextBanner(ctx context.Context, id int64) error
	DeleteOrderStatus(ctx context.Context, id int64) error
	DeletePaymentMethod(ctx context.Context, arg DeletePaymentMethodParams) (PaymentMethod, error)
	DeletePaymentType(ctx context.Context, id int64) error
	DeleteProduct(ctx context.Context, id int64) error
	DeleteProductBrand(ctx context.Context, id int64) error
	DeleteProductCategory(ctx context.Context, arg DeleteProductCategoryParams) error
	DeleteProductColor(ctx context.Context, id int64) error
	DeleteProductConfiguration(ctx context.Context, arg DeleteProductConfigurationParams) error
	DeleteProductImage(ctx context.Context, id int64) error
	DeleteProductItem(ctx context.Context, id int64) error
	DeleteProductPromotion(ctx context.Context, arg DeleteProductPromotionParams) error
	DeleteProductSize(ctx context.Context, id int64) error
	DeletePromotion(ctx context.Context, id int64) error
	DeleteShippingMethod(ctx context.Context, id int64) error
	DeleteShopOrder(ctx context.Context, id int64) error
	DeleteShopOrderItem(ctx context.Context, id int64) error
	DeleteShoppingCart(ctx context.Context, id int64) error
	DeleteShoppingCartItem(ctx context.Context, arg DeleteShoppingCartItemParams) error
	DeleteShoppingCartItemAllByUser(ctx context.Context, arg DeleteShoppingCartItemAllByUserParams) ([]ShoppingCartItem, error)
	DeleteUser(ctx context.Context, id int64) (User, error)
	// -- name: UpdateUserAddressWithAddress :one
	// WITH t1 AS (
	//     UPDATE "address" as a
	//     SET
	//     address_line = COALESCE(sqlc.narg(address_line),address_line),
	//     region = COALESCE(sqlc.narg(region),region),
	//     city= COALESCE(sqlc.narg(city),city)
	//     WHERE id = COALESCE(sqlc.arg(id),id)
	//     RETURNING a.id, a.address_line, a.region, a.city
	//    ),
	//     t2 AS (
	//     UPDATE "user_address"
	//     SET
	//     default_address = COALESCE(sqlc.narg(default_address),default_address)
	//     WHERE
	//     user_id = COALESCE(sqlc.arg(user_id),user_id)
	//     AND address_id = COALESCE(sqlc.arg(address_id),address_id)
	//     RETURNING user_id, address_id, default_address
	// 	)
	// SELECT
	// user_id,
	// address_id,
	// default_address,
	// address_line,
	// region,
	// city From t1,t2;
	DeleteUserAddress(ctx context.Context, arg DeleteUserAddressParams) (UserAddress, error)
	DeleteUserReview(ctx context.Context, arg DeleteUserReviewParams) (UserReview, error)
	DeleteVariation(ctx context.Context, id int64) error
	DeleteVariationOption(ctx context.Context, id int64) error
	DeleteWishList(ctx context.Context, id int64) error
	// WITH t1 AS (
	//   SELECT id FROM "wish_list" AS wl
	//   WHERE wl.user_id = sqlc.arg(user_id)
	// )
	DeleteWishListItem(ctx context.Context, arg DeleteWishListItemParams) error
	// WITH t1 AS(
	//   SELECT id FROM "wish_list" WHERE user_id = $1
	// )
	DeleteWishListItemAll(ctx context.Context, wishListID int64) ([]WishListItem, error)
	GetAddress(ctx context.Context, id int64) (Address, error)
	GetAddressByCity(ctx context.Context, city string) (Address, error)
	GetAdmin(ctx context.Context, id int64) (Admin, error)
	GetAdminByEmail(ctx context.Context, email string) (Admin, error)
	GetAdminType(ctx context.Context, id int64) (AdminType, error)
	GetBrandPromotion(ctx context.Context, arg GetBrandPromotionParams) (BrandPromotion, error)
	GetCategoryPromotion(ctx context.Context, arg GetCategoryPromotionParams) (CategoryPromotion, error)
	GetHomePageTextBanner(ctx context.Context, id int64) (HomePageTextBanner, error)
	GetOrderStatus(ctx context.Context, id int64) (OrderStatus, error)
	GetOrderStatusByUserID(ctx context.Context, arg GetOrderStatusByUserIDParams) (GetOrderStatusByUserIDRow, error)
	// id = $1
	GetPaymentMethod(ctx context.Context, arg GetPaymentMethodParams) (PaymentMethod, error)
	GetPaymentType(ctx context.Context, id int64) (PaymentType, error)
	GetProduct(ctx context.Context, id int64) (Product, error)
	GetProductBrand(ctx context.Context, id int64) (ProductBrand, error)
	GetProductCategory(ctx context.Context, id int64) (ProductCategory, error)
	GetProductCategoryByParent(ctx context.Context, arg GetProductCategoryByParentParams) (ProductCategory, error)
	GetProductColor(ctx context.Context, id int64) (ProductColor, error)
	GetProductConfiguration(ctx context.Context, arg GetProductConfigurationParams) (ProductConfiguration, error)
	GetProductImage(ctx context.Context, id int64) (ProductImage, error)
	GetProductItem(ctx context.Context, id int64) (ProductItem, error)
	GetProductItemForUpdate(ctx context.Context, id int64) (ProductItem, error)
	GetProductPromotion(ctx context.Context, arg GetProductPromotionParams) (ProductPromotion, error)
	GetProductSize(ctx context.Context, id int64) (ProductSize, error)
	GetPromotion(ctx context.Context, id int64) (Promotion, error)
	GetShippingMethod(ctx context.Context, id int64) (ShippingMethod, error)
	GetShippingMethodByUserID(ctx context.Context, arg GetShippingMethodByUserIDParams) (GetShippingMethodByUserIDRow, error)
	GetShopOrder(ctx context.Context, id int64) (ShopOrder, error)
	GetShopOrderItem(ctx context.Context, id int64) (ShopOrderItem, error)
	GetShopOrderItemByUserIDOrderID(ctx context.Context, arg GetShopOrderItemByUserIDOrderIDParams) (GetShopOrderItemByUserIDOrderIDRow, error)
	GetShoppingCart(ctx context.Context, id int64) (ShoppingCart, error)
	GetShoppingCartByUserIDCartID(ctx context.Context, arg GetShoppingCartByUserIDCartIDParams) (ShoppingCart, error)
	GetShoppingCartItem(ctx context.Context, id int64) (ShoppingCartItem, error)
	GetShoppingCartItemByUserIDCartID(ctx context.Context, arg GetShoppingCartItemByUserIDCartIDParams) ([]GetShoppingCartItemByUserIDCartIDRow, error)
	GetUser(ctx context.Context, id int64) (User, error)
	GetUserAddress(ctx context.Context, arg GetUserAddressParams) (UserAddress, error)
	GetUserAddressWithAddress(ctx context.Context, arg GetUserAddressWithAddressParams) (GetUserAddressWithAddressRow, error)
	// SELECT * FROM "user"
	// WHERE email = $1 LIMIT 1;
	GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error)
	GetUserReview(ctx context.Context, arg GetUserReviewParams) (UserReview, error)
	GetUserSession(ctx context.Context, id uuid.UUID) (UserSession, error)
	GetVariation(ctx context.Context, id int64) (Variation, error)
	GetVariationOption(ctx context.Context, id int64) (VariationOption, error)
	GetWishList(ctx context.Context, id int64) (WishList, error)
	GetWishListByUserID(ctx context.Context, userID int64) (WishList, error)
	GetWishListItem(ctx context.Context, id int64) (WishListItem, error)
	GetWishListItemByUserIDCartID(ctx context.Context, arg GetWishListItemByUserIDCartIDParams) (WishListItem, error)
	ListAddressesByCity(ctx context.Context, arg ListAddressesByCityParams) ([]Address, error)
	ListAddressesByID(ctx context.Context, addressesIds []int64) ([]Address, error)
	ListAdminTypes(ctx context.Context, arg ListAdminTypesParams) ([]AdminType, error)
	ListAdmins(ctx context.Context, arg ListAdminsParams) ([]Admin, error)
	ListBrandPromotions(ctx context.Context, arg ListBrandPromotionsParams) ([]BrandPromotion, error)
	ListBrandPromotionsWithImages(ctx context.Context) ([]ListBrandPromotionsWithImagesRow, error)
	ListCategoryPromotions(ctx context.Context, arg ListCategoryPromotionsParams) ([]CategoryPromotion, error)
	ListCategoryPromotionsWithImages(ctx context.Context) ([]ListCategoryPromotionsWithImagesRow, error)
	ListHomePageTextBanners(ctx context.Context) ([]HomePageTextBanner, error)
	ListOrderStatuses(ctx context.Context) ([]OrderStatus, error)
	// ORDER BY id
	// LIMIT $1
	// OFFSET $2;
	ListOrderStatusesByUserID(ctx context.Context, arg ListOrderStatusesByUserIDParams) ([]ListOrderStatusesByUserIDRow, error)
	ListPaymentMethods(ctx context.Context, arg ListPaymentMethodsParams) ([]PaymentMethod, error)
	ListPaymentTypes(ctx context.Context) ([]PaymentType, error)
	ListProductBrands(ctx context.Context) ([]ProductBrand, error)
	ListProductCategories(ctx context.Context) ([]ProductCategory, error)
	// LIMIT $1
	// OFFSET $2;
	ListProductCategoriesByParent(ctx context.Context, parentCategoryID null.Int) ([]ProductCategory, error)
	ListProductConfigurations(ctx context.Context, arg ListProductConfigurationsParams) ([]ProductConfiguration, error)
	ListProductItems(ctx context.Context, arg ListProductItemsParams) ([]ListProductItemsRow, error)
	ListProductItemsByIDs(ctx context.Context, productsIds []int64) ([]ListProductItemsByIDsRow, error)
	ListProductItemsNextPage(ctx context.Context, arg ListProductItemsNextPageParams) ([]ListProductItemsNextPageRow, error)
	ListProductItemsV2(ctx context.Context, arg ListProductItemsV2Params) ([]ListProductItemsV2Row, error)
	ListProductPromotions(ctx context.Context, arg ListProductPromotionsParams) ([]ProductPromotion, error)
	ListProductPromotionsWithImages(ctx context.Context) ([]ListProductPromotionsWithImagesRow, error)
	// WITH total_records AS (
	//   SELECT COUNT(id)
	//   FROM "product"
	// ),
	// list_products AS (
	ListProducts(ctx context.Context, arg ListProductsParams) ([]ListProductsRow, error)
	ListPromotions(ctx context.Context, arg ListPromotionsParams) ([]Promotion, error)
	ListShippingMethods(ctx context.Context) ([]ShippingMethod, error)
	// ORDER BY id
	// LIMIT $1
	// OFFSET $2;
	ListShippingMethodsByUserID(ctx context.Context, arg ListShippingMethodsByUserIDParams) ([]ListShippingMethodsByUserIDRow, error)
	ListShopOrderItems(ctx context.Context, arg ListShopOrderItemsParams) ([]ShopOrderItem, error)
	// ORDER BY soi.id;
	ListShopOrderItemsByUserID(ctx context.Context, arg ListShopOrderItemsByUserIDParams) ([]ListShopOrderItemsByUserIDRow, error)
	// SELECT * FROM "shop_order_item"
	// WHERE order_id = $1
	// ORDER BY id;
	// pi.product_image,
	ListShopOrderItemsByUserIDOrderID(ctx context.Context, arg ListShopOrderItemsByUserIDOrderIDParams) ([]ListShopOrderItemsByUserIDOrderIDRow, error)
	ListShopOrders(ctx context.Context, arg ListShopOrdersParams) ([]ShopOrder, error)
	ListShopOrdersByUserID(ctx context.Context, arg ListShopOrdersByUserIDParams) ([]ListShopOrdersByUserIDRow, error)
	// LIMIT 1;
	ListShoppingCartItems(ctx context.Context, arg ListShoppingCartItemsParams) ([]ShoppingCartItem, error)
	ListShoppingCartItemsByCartID(ctx context.Context, shoppingCartID int64) ([]ShoppingCartItem, error)
	ListShoppingCartItemsByUserID(ctx context.Context, userID int64) ([]ListShoppingCartItemsByUserIDRow, error)
	ListShoppingCarts(ctx context.Context, arg ListShoppingCartsParams) ([]ShoppingCart, error)
	ListUserAddresses(ctx context.Context, arg ListUserAddressesParams) ([]UserAddress, error)
	ListUserReviews(ctx context.Context, arg ListUserReviewsParams) ([]UserReview, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
	ListVariationOptions(ctx context.Context, arg ListVariationOptionsParams) ([]VariationOption, error)
	ListVariations(ctx context.Context, arg ListVariationsParams) ([]Variation, error)
	ListWishListItems(ctx context.Context, arg ListWishListItemsParams) ([]WishListItem, error)
	ListWishListItemsByCartID(ctx context.Context, wishListID int64) ([]WishListItem, error)
	ListWishListItemsByUserID(ctx context.Context, userID int64) ([]ListWishListItemsByUserIDRow, error)
	ListWishLists(ctx context.Context, arg ListWishListsParams) ([]WishList, error)
	SearchProductItems(ctx context.Context, arg SearchProductItemsParams) ([]SearchProductItemsRow, error)
	SearchProductItemsNextPage(ctx context.Context, arg SearchProductItemsNextPageParams) ([]SearchProductItemsNextPageRow, error)
	UpdateAddress(ctx context.Context, arg UpdateAddressParams) (Address, error)
	UpdateAdmin(ctx context.Context, arg UpdateAdminParams) (Admin, error)
	UpdateAdminType(ctx context.Context, arg UpdateAdminTypeParams) (AdminType, error)
	UpdateBrandPromotion(ctx context.Context, arg UpdateBrandPromotionParams) (BrandPromotion, error)
	UpdateCategoryPromotion(ctx context.Context, arg UpdateCategoryPromotionParams) (CategoryPromotion, error)
	UpdateHomePageTextBanner(ctx context.Context, arg UpdateHomePageTextBannerParams) (HomePageTextBanner, error)
	UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) (OrderStatus, error)
	UpdatePaymentMethod(ctx context.Context, arg UpdatePaymentMethodParams) (PaymentMethod, error)
	// ORDER BY id
	// LIMIT $1
	// OFFSET $2;
	UpdatePaymentType(ctx context.Context, arg UpdatePaymentTypeParams) (PaymentType, error)
	// )
	// SELECT *
	// FROM list_products, total_records;
	// product_image = COALESCE(sqlc.narg(product_image),product_image),
	UpdateProduct(ctx context.Context, arg UpdateProductParams) (Product, error)
	// LIMIT $1
	// OFFSET $2;
	UpdateProductBrand(ctx context.Context, arg UpdateProductBrandParams) (ProductBrand, error)
	// LIMIT $2
	// OFFSET $3;
	UpdateProductCategory(ctx context.Context, arg UpdateProductCategoryParams) (ProductCategory, error)
	UpdateProductColor(ctx context.Context, arg UpdateProductColorParams) (ProductColor, error)
	UpdateProductConfiguration(ctx context.Context, arg UpdateProductConfigurationParams) (ProductConfiguration, error)
	UpdateProductImage(ctx context.Context, arg UpdateProductImageParams) (ProductImage, error)
	UpdateProductItem(ctx context.Context, arg UpdateProductItemParams) (ProductItem, error)
	UpdateProductPromotion(ctx context.Context, arg UpdateProductPromotionParams) (ProductPromotion, error)
	UpdateProductSize(ctx context.Context, arg UpdateProductSizeParams) (ProductSize, error)
	UpdatePromotion(ctx context.Context, arg UpdatePromotionParams) (Promotion, error)
	UpdateShippingMethod(ctx context.Context, arg UpdateShippingMethodParams) (ShippingMethod, error)
	UpdateShopOrder(ctx context.Context, arg UpdateShopOrderParams) (ShopOrder, error)
	// -- name: ListShopOrderItemsByOrderID :many
	// SELECT * FROM "shop_order_item"
	// WHERE order_id = $1
	// ORDER BY id;
	UpdateShopOrderItem(ctx context.Context, arg UpdateShopOrderItemParams) (ShopOrderItem, error)
	UpdateShoppingCart(ctx context.Context, arg UpdateShoppingCartParams) (ShoppingCart, error)
	UpdateShoppingCartItem(ctx context.Context, arg UpdateShoppingCartItemParams) (UpdateShoppingCartItemRow, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserAddress(ctx context.Context, arg UpdateUserAddressParams) (UserAddress, error)
	UpdateUserReview(ctx context.Context, arg UpdateUserReviewParams) (UserReview, error)
	UpdateUserSession(ctx context.Context, arg UpdateUserSessionParams) (UserSession, error)
	UpdateVariation(ctx context.Context, arg UpdateVariationParams) (Variation, error)
	UpdateVariationOption(ctx context.Context, arg UpdateVariationOptionParams) (VariationOption, error)
	UpdateWishList(ctx context.Context, arg UpdateWishListParams) (WishList, error)
	// WITH t1 AS (
	//   SELECT user_id FROM "wish_list" AS wl
	//   WHERE wl.id = sqlc.arg(wish_list_id)
	// )
	UpdateWishListItem(ctx context.Context, arg UpdateWishListItemParams) (WishListItem, error)
}

var _ Querier = (*Queries)(nil)
