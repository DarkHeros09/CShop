// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/google/uuid"
	null "github.com/guregu/null/v5"
)

type Querier interface {
	AdminCreateBrandPromotion(ctx context.Context, arg AdminCreateBrandPromotionParams) (BrandPromotion, error)
	AdminCreateCategoryPromotion(ctx context.Context, arg AdminCreateCategoryPromotionParams) (CategoryPromotion, error)
	AdminCreateFeaturedProductItem(ctx context.Context, arg AdminCreateFeaturedProductItemParams) (FeaturedProductItem, error)
	AdminCreatePaymentType(ctx context.Context, arg AdminCreatePaymentTypeParams) (PaymentType, error)
	AdminCreateProduct(ctx context.Context, arg AdminCreateProductParams) (Product, error)
	AdminCreateProductBrand(ctx context.Context, arg AdminCreateProductBrandParams) (ProductBrand, error)
	AdminCreateProductCategory(ctx context.Context, arg AdminCreateProductCategoryParams) (ProductCategory, error)
	AdminCreateProductColor(ctx context.Context, arg AdminCreateProductColorParams) (ProductColor, error)
	AdminCreateProductImages(ctx context.Context, arg AdminCreateProductImagesParams) (ProductImage, error)
	//   size_id,
	//   qty_in_stock,
	AdminCreateProductItem(ctx context.Context, arg AdminCreateProductItemParams) (ProductItem, error)
	AdminCreateProductPromotion(ctx context.Context, arg AdminCreateProductPromotionParams) (ProductPromotion, error)
	AdminCreateProductSize(ctx context.Context, arg AdminCreateProductSizeParams) (ProductSize, error)
	AdminCreatePromotion(ctx context.Context, arg AdminCreatePromotionParams) (Promotion, error)
	AdminCreateShippingMethod(ctx context.Context, arg AdminCreateShippingMethodParams) (ShippingMethod, error)
	AdminDeletePaymentType(ctx context.Context, arg AdminDeletePaymentTypeParams) error
	AdminDeleteProduct(ctx context.Context, arg AdminDeleteProductParams) error
	AdminListBrandPromotions(ctx context.Context, adminID int64) ([]AdminListBrandPromotionsRow, error)
	AdminListCategoryPromotions(ctx context.Context, adminID int64) ([]AdminListCategoryPromotionsRow, error)
	AdminListFeaturedProductItems(ctx context.Context, adminID int64) ([]AdminListFeaturedProductItemsRow, error)
	// ORDER BY id
	// LIMIT $1
	// OFFSET $2;
	AdminListOrderStatuses(ctx context.Context, adminID int64) ([]OrderStatus, error)
	// ORDER BY id
	// LIMIT $1
	// OFFSET $2;
	AdminListPaymentTypes(ctx context.Context, adminID int64) ([]PaymentType, error)
	AdminListProductPromotions(ctx context.Context, adminID int64) ([]AdminListProductPromotionsRow, error)
	AdminListShopOrdersNextPage(ctx context.Context, arg AdminListShopOrdersNextPageParams) ([]AdminListShopOrdersNextPageRow, error)
	AdminListShopOrdersV2(ctx context.Context, arg AdminListShopOrdersV2Params) ([]AdminListShopOrdersV2Row, error)
	AdminUpdateBrandPromotion(ctx context.Context, arg AdminUpdateBrandPromotionParams) (BrandPromotion, error)
	AdminUpdateCategoryPromotion(ctx context.Context, arg AdminUpdateCategoryPromotionParams) (CategoryPromotion, error)
	AdminUpdateFeaturedProductItem(ctx context.Context, arg AdminUpdateFeaturedProductItemParams) (FeaturedProductItem, error)
	AdminUpdatePaymentType(ctx context.Context, arg AdminUpdatePaymentTypeParams) (PaymentType, error)
	// product_image = COALESCE(sqlc.narg(product_image),product_image),
	AdminUpdateProduct(ctx context.Context, arg AdminUpdateProductParams) (Product, error)
	AdminUpdateProductColor(ctx context.Context, arg AdminUpdateProductColorParams) (ProductColor, error)
	AdminUpdateProductImage(ctx context.Context, arg AdminUpdateProductImageParams) (ProductImage, error)
	// qty_in_stock = COALESCE(sqlc.narg(qty_in_stock),qty_in_stock),
	// size_id = COALESCE(sqlc.narg(size_id),size_id),
	AdminUpdateProductItem(ctx context.Context, arg AdminUpdateProductItemParams) (ProductItem, error)
	AdminUpdateProductPromotion(ctx context.Context, arg AdminUpdateProductPromotionParams) (ProductPromotion, error)
	AdminUpdateProductSize(ctx context.Context, arg AdminUpdateProductSizeParams) (ProductSize, error)
	AdminUpdatePromotion(ctx context.Context, arg AdminUpdatePromotionParams) (Promotion, error)
	AdminUpdateShippingMethod(ctx context.Context, arg AdminUpdateShippingMethodParams) (ShippingMethod, error)
	CheckUserAddressDefaultAddress(ctx context.Context, userID int64) (int64, error)
	CreateAddress(ctx context.Context, arg CreateAddressParams) (Address, error)
	CreateAdmin(ctx context.Context, arg CreateAdminParams) (Admin, error)
	CreateAdminSession(ctx context.Context, arg CreateAdminSessionParams) (AdminSession, error)
	CreateAdminType(ctx context.Context, adminType string) (AdminType, error)
	CreateAppPolicy(ctx context.Context, arg CreateAppPolicyParams) (AppPolicy, error)
	CreateBrandPromotion(ctx context.Context, arg CreateBrandPromotionParams) (BrandPromotion, error)
	CreateCategoryPromotion(ctx context.Context, arg CreateCategoryPromotionParams) (CategoryPromotion, error)
	CreateHomePageTextBanner(ctx context.Context, arg CreateHomePageTextBannerParams) (HomePageTextBanner, error)
	// ON CONFLICT(user_id) DO UPDATE SET
	// device_id = EXCLUDED.device_id,
	// fcm_token = EXCLUDED.fcm_token
	CreateNotification(ctx context.Context, arg CreateNotificationParams) (Notification, error)
	CreateOrderStatus(ctx context.Context, status string) (OrderStatus, error)
	CreatePaymentMethod(ctx context.Context, arg CreatePaymentMethodParams) (PaymentMethod, error)
	CreatePaymentType(ctx context.Context, value string) (PaymentType, error)
	CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error)
	CreateProductBrand(ctx context.Context, arg CreateProductBrandParams) (ProductBrand, error)
	CreateProductCategory(ctx context.Context, arg CreateProductCategoryParams) (ProductCategory, error)
	CreateProductColor(ctx context.Context, colorValue string) (ProductColor, error)
	CreateProductConfiguration(ctx context.Context, arg CreateProductConfigurationParams) (ProductConfiguration, error)
	CreateProductImage(ctx context.Context, arg CreateProductImageParams) (ProductImage, error)
	//   size_id,
	//   qty_in_stock,
	CreateProductItem(ctx context.Context, arg CreateProductItemParams) (ProductItem, error)
	CreateProductPromotion(ctx context.Context, arg CreateProductPromotionParams) (ProductPromotion, error)
	CreateProductSize(ctx context.Context, arg CreateProductSizeParams) (ProductSize, error)
	CreatePromotion(ctx context.Context, arg CreatePromotionParams) (Promotion, error)
	CreateResetPassword(ctx context.Context, arg CreateResetPasswordParams) (ResetPassword, error)
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
	CreateVerifyEmail(ctx context.Context, arg CreateVerifyEmailParams) (VerifyEmail, error)
	CreateWishList(ctx context.Context, userID int64) (WishList, error)
	CreateWishListItem(ctx context.Context, arg CreateWishListItemParams) (WishListItem, error)
	DeleteAddress(ctx context.Context, id int64) error
	DeleteAdmin(ctx context.Context, id int64) error
	DeleteAdminTypeByID(ctx context.Context, id int64) error
	DeleteAdminTypeByType(ctx context.Context, adminType string) error
	DeleteAppPolicy(ctx context.Context, arg DeleteAppPolicyParams) (AppPolicy, error)
	DeleteBrandPromotion(ctx context.Context, arg DeleteBrandPromotionParams) error
	DeleteCategoryPromotion(ctx context.Context, arg DeleteCategoryPromotionParams) error
	DeleteFeaturedProductItem(ctx context.Context, arg DeleteFeaturedProductItemParams) error
	DeleteHomePageTextBanner(ctx context.Context, arg DeleteHomePageTextBannerParams) error
	DeleteNotification(ctx context.Context, arg DeleteNotificationParams) (Notification, error)
	DeleteNotificationAllByUser(ctx context.Context, userID int64) error
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
	DeleteProductSizeByProductItemID(ctx context.Context, productItemID int64) error
	DeletePromotion(ctx context.Context, id int64) error
	DeleteShippingMethod(ctx context.Context, id int64) error
	DeleteShopOrder(ctx context.Context, id int64) error
	DeleteShopOrderItem(ctx context.Context, arg DeleteShopOrderItemParams) (ShopOrderItem, error)
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
	DeleteUserByEmailNotVerified(ctx context.Context, email string) error
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
	GetActiveProductItems(ctx context.Context, adminID int64) (int64, error)
	GetActiveUsersCount(ctx context.Context, adminID int64) (int64, error)
	GetAddress(ctx context.Context, id int64) (Address, error)
	GetAddressByCity(ctx context.Context, city string) (Address, error)
	GetAdmin(ctx context.Context, id int64) (Admin, error)
	GetAdminByEmail(ctx context.Context, email string) (Admin, error)
	GetAdminSession(ctx context.Context, id uuid.UUID) (AdminSession, error)
	GetAdminType(ctx context.Context, id int64) (AdminType, error)
	GetAppPolicy(ctx context.Context) (AppPolicy, error)
	GetBrandPromotion(ctx context.Context, arg GetBrandPromotionParams) (BrandPromotion, error)
	GetCategoryPromotion(ctx context.Context, arg GetCategoryPromotionParams) (CategoryPromotion, error)
	GetCompletedDailyOrderTotal(ctx context.Context, adminID int64) (string, error)
	GetFeaturedProductItem(ctx context.Context, productItemID int64) (FeaturedProductItem, error)
	GetHomePageTextBanner(ctx context.Context, id int64) (HomePageTextBanner, error)
	// AND secret_code = $2
	GetLastUsedResetPassword(ctx context.Context, email string) (ResetPassword, error)
	GetNotification(ctx context.Context, arg GetNotificationParams) (Notification, error)
	GetNotificationV2(ctx context.Context, userID int64) (Notification, error)
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
	GetProductItem(ctx context.Context, productItemID int64) (GetProductItemRow, error)
	GetProductItemForUpdate(ctx context.Context, id int64) (ProductItem, error)
	GetProductItemSizeForUpdate(ctx context.Context, id int64) (ProductSize, error)
	GetProductItemWithPromotions(ctx context.Context, id int64) (GetProductItemWithPromotionsRow, error)
	GetProductPromotion(ctx context.Context, arg GetProductPromotionParams) (ProductPromotion, error)
	GetProductSize(ctx context.Context, id int64) (ProductSize, error)
	GetProductsByIDs(ctx context.Context, ids []int64) ([]Product, error)
	GetPromotion(ctx context.Context, id int64) (Promotion, error)
	GetResetPassword(ctx context.Context, id int64) (ResetPassword, error)
	GetResetPasswordUserIDByID(ctx context.Context, arg GetResetPasswordUserIDByIDParams) (int64, error)
	GetResetPasswordsByEmail(ctx context.Context, email string) (GetResetPasswordsByEmailRow, error)
	GetShippingMethod(ctx context.Context, id int64) (ShippingMethod, error)
	GetShippingMethodByUserID(ctx context.Context, arg GetShippingMethodByUserIDParams) (GetShippingMethodByUserIDRow, error)
	GetShopOrder(ctx context.Context, id int64) (ShopOrder, error)
	GetShopOrderItem(ctx context.Context, id int64) (ShopOrderItem, error)
	GetShopOrderItemByUserIDOrderID(ctx context.Context, arg GetShopOrderItemByUserIDOrderIDParams) (GetShopOrderItemByUserIDOrderIDRow, error)
	GetShopOrdersCountByStatusId(ctx context.Context, arg GetShopOrdersCountByStatusIdParams) (int64, error)
	GetShoppingCart(ctx context.Context, id int64) (ShoppingCart, error)
	GetShoppingCartByUserIDCartID(ctx context.Context, arg GetShoppingCartByUserIDCartIDParams) (ShoppingCart, error)
	GetShoppingCartItem(ctx context.Context, id int64) (ShoppingCartItem, error)
	GetShoppingCartItemByUserIDCartID(ctx context.Context, arg GetShoppingCartItemByUserIDCartIDParams) ([]GetShoppingCartItemByUserIDCartIDRow, error)
	GetTotalProductItems(ctx context.Context, adminID int64) (int64, error)
	GetTotalShopOrder(ctx context.Context, adminID int64) (int64, error)
	GetTotalUsersCount(ctx context.Context, adminID int64) (int64, error)
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
	GetVerifyEmail(ctx context.Context, id int64) (VerifyEmail, error)
	GetVerifyEmailByEmail(ctx context.Context, email string) (GetVerifyEmailByEmailRow, error)
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
	ListFeaturedProductItems(ctx context.Context, arg ListFeaturedProductItemsParams) ([]FeaturedProductItem, error)
	ListHomePageTextBanners(ctx context.Context) ([]HomePageTextBanner, error)
	ListOrderStatuses(ctx context.Context) ([]OrderStatus, error)
	ListOrderStatusesByUserID(ctx context.Context, arg ListOrderStatusesByUserIDParams) ([]ListOrderStatusesByUserIDRow, error)
	ListPaymentMethods(ctx context.Context, arg ListPaymentMethodsParams) ([]PaymentMethod, error)
	ListPaymentTypes(ctx context.Context) ([]PaymentType, error)
	ListProductBrands(ctx context.Context) ([]ProductBrand, error)
	ListProductCategories(ctx context.Context) ([]ProductCategory, error)
	// LIMIT $1
	// OFFSET $2;
	ListProductCategoriesByParent(ctx context.Context, parentCategoryID null.Int) ([]ProductCategory, error)
	ListProductColors(ctx context.Context) ([]ProductColor, error)
	ListProductConfigurations(ctx context.Context, arg ListProductConfigurationsParams) ([]ProductConfiguration, error)
	ListProductImagesNextPage(ctx context.Context, arg ListProductImagesNextPageParams) ([]ListProductImagesNextPageRow, error)
	ListProductImagesV2(ctx context.Context, limit int32) ([]ListProductImagesV2Row, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItems(ctx context.Context, arg ListProductItemsParams) ([]ListProductItemsRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsByIDs(ctx context.Context, productsIds []int64) ([]ListProductItemsByIDsRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	// AND CASE
	//     WHEN COALESCE(sqlc.narg(size_id), 0) > 0
	//     THEN pi.size_id = sqlc.narg(size_id)
	//     ELSE 1=1
	// END
	// CASE
	//     WHEN COALESCE(sqlc.narg(order_by_featured), FALSE) = TRUE
	//     THEN fpi.id END DESC,
	ListProductItemsNextPage(ctx context.Context, arg ListProductItemsNextPageParams) ([]ListProductItemsNextPageRow, error)
	// WITH t1 AS (
	// SELECT COUNT(*) OVER() AS total_count
	// FROM "product_item" AS pi
	// WHERE pi.active = TRUE
	// LIMIT 1
	// )
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsNextPageOld(ctx context.Context, arg ListProductItemsNextPageOldParams) ([]ListProductItemsNextPageOldRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	// AND CASE
	//     WHEN COALESCE(sqlc.narg(size_id), 0) > 0
	//     THEN pi.size_id = sqlc.narg(size_id)
	//     ELSE 1=1
	// END
	// CASE
	//     WHEN COALESCE(sqlc.narg(order_by_featured), FALSE) = TRUE
	//     THEN fpi.id END DESC,
	ListProductItemsV2(ctx context.Context, arg ListProductItemsV2Params) ([]ListProductItemsV2Row, error)
	// WITH t1 (total_count) AS (
	// SELECT COUNT(*) OVER() AS total_count
	// FROM "product_item" AS pi
	// WHERE pi.active = TRUE
	// LIMIT 1
	// )
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsV2Old(ctx context.Context, arg ListProductItemsV2OldParams) ([]ListProductItemsV2OldRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsWithBestSales(ctx context.Context, limit int32) ([]ListProductItemsWithBestSalesRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsWithBrandPromotions(ctx context.Context, arg ListProductItemsWithBrandPromotionsParams) ([]ListProductItemsWithBrandPromotionsRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsWithBrandPromotionsNextPage(ctx context.Context, arg ListProductItemsWithBrandPromotionsNextPageParams) ([]ListProductItemsWithBrandPromotionsNextPageRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsWithCategoryPromotions(ctx context.Context, arg ListProductItemsWithCategoryPromotionsParams) ([]ListProductItemsWithCategoryPromotionsRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsWithCategoryPromotionsNextPage(ctx context.Context, arg ListProductItemsWithCategoryPromotionsNextPageParams) ([]ListProductItemsWithCategoryPromotionsNextPageRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsWithPromotions(ctx context.Context, arg ListProductItemsWithPromotionsParams) ([]ListProductItemsWithPromotionsRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	ListProductItemsWithPromotionsNextPage(ctx context.Context, arg ListProductItemsWithPromotionsNextPageParams) ([]ListProductItemsWithPromotionsNextPageRow, error)
	ListProductPromotions(ctx context.Context, arg ListProductPromotionsParams) ([]ProductPromotion, error)
	ListProductPromotionsWithImages(ctx context.Context) ([]ListProductPromotionsWithImagesRow, error)
	ListProductSizes(ctx context.Context) ([]ProductSize, error)
	// JOIN "shopping_cart_item" AS sci ON sci.size_id = ps.product_item_id
	ListProductSizesByIDs(ctx context.Context, sizesIds []int64) ([]ProductSize, error)
	ListProductSizesByProductItemID(ctx context.Context, productItemID int64) ([]ProductSize, error)
	// WITH total_records AS (
	//   SELECT COUNT(id)
	//   FROM "product"
	// ),
	// list_products AS (
	ListProducts(ctx context.Context, arg ListProductsParams) ([]ListProductsRow, error)
	ListProductsNextPage(ctx context.Context, arg ListProductsNextPageParams) ([]ListProductsNextPageRow, error)
	ListProductsV2(ctx context.Context, limit int32) ([]ListProductsV2Row, error)
	ListPromotions(ctx context.Context) ([]Promotion, error)
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
	// , pt.value AS payment_type
	// LEFT JOIN "product_size" AS psize ON psize.id = pi.size_id
	// LEFT JOIN "payment_method" AS pm ON pm.id = so.payment_method_id
	// LEFT JOIN "shipping_method" AS sm ON sm.id = so.shipping_method_id
	ListShopOrderItemsByUserIDOrderID(ctx context.Context, arg ListShopOrderItemsByUserIDOrderIDParams) ([]ListShopOrderItemsByUserIDOrderIDRow, error)
	ListShopOrders(ctx context.Context, arg ListShopOrdersParams) ([]ShopOrder, error)
	ListShopOrdersByUserID(ctx context.Context, arg ListShopOrdersByUserIDParams) ([]ListShopOrdersByUserIDRow, error)
	// ROW_NUMBER() OVER(ORDER BY so.id) AS order_number,
	ListShopOrdersByUserIDNextPage(ctx context.Context, arg ListShopOrdersByUserIDNextPageParams) ([]ListShopOrdersByUserIDNextPageRow, error)
	// ROW_NUMBER() OVER(ORDER BY so.id) AS order_number,
	ListShopOrdersByUserIDV2(ctx context.Context, arg ListShopOrdersByUserIDV2Params) ([]ListShopOrdersByUserIDV2Row, error)
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
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	SearchProductItems(ctx context.Context, arg SearchProductItemsParams) ([]SearchProductItemsRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	SearchProductItemsNextPage(ctx context.Context, arg SearchProductItemsNextPageParams) ([]SearchProductItemsNextPageRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	SearchProductItemsNextPageOld(ctx context.Context, arg SearchProductItemsNextPageOldParams) ([]SearchProductItemsNextPageOldRow, error)
	// LEFT JOIN "product_size" AS ps ON ps.product_item_id = pi.id
	SearchProductItemsOld(ctx context.Context, arg SearchProductItemsOldParams) ([]SearchProductItemsOldRow, error)
	SearchProducts(ctx context.Context, arg SearchProductsParams) ([]SearchProductsRow, error)
	SearchProductsNextPage(ctx context.Context, arg SearchProductsNextPageParams) ([]SearchProductsNextPageRow, error)
	UpdateAddress(ctx context.Context, arg UpdateAddressParams) (Address, error)
	UpdateAdmin(ctx context.Context, arg UpdateAdminParams) (Admin, error)
	UpdateAdminSession(ctx context.Context, arg UpdateAdminSessionParams) (AdminSession, error)
	UpdateAdminType(ctx context.Context, arg UpdateAdminTypeParams) (AdminType, error)
	UpdateAppPolicy(ctx context.Context, arg UpdateAppPolicyParams) (AppPolicy, error)
	UpdateBrandPromotion(ctx context.Context, arg UpdateBrandPromotionParams) (BrandPromotion, error)
	UpdateCategoryPromotion(ctx context.Context, arg UpdateCategoryPromotionParams) (CategoryPromotion, error)
	UpdateHomePageTextBanner(ctx context.Context, arg UpdateHomePageTextBannerParams) (HomePageTextBanner, error)
	UpdateNotification(ctx context.Context, arg UpdateNotificationParams) (Notification, error)
	UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) (OrderStatus, error)
	UpdatePaymentMethod(ctx context.Context, arg UpdatePaymentMethodParams) (PaymentMethod, error)
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
	// qty_in_stock = COALESCE(sqlc.narg(qty_in_stock),qty_in_stock),
	// size_id = COALESCE(sqlc.narg(size_id),size_id),
	UpdateProductItem(ctx context.Context, arg UpdateProductItemParams) (ProductItem, error)
	UpdateProductPromotion(ctx context.Context, arg UpdateProductPromotionParams) (ProductPromotion, error)
	UpdateProductSize(ctx context.Context, arg UpdateProductSizeParams) (ProductSize, error)
	// LIMIT $1
	// OFFSET $2;
	UpdatePromotion(ctx context.Context, arg UpdatePromotionParams) (Promotion, error)
	UpdateResetPassword(ctx context.Context, arg UpdateResetPasswordParams) (ResetPassword, error)
	UpdateShippingMethod(ctx context.Context, arg UpdateShippingMethodParams) (ShippingMethod, error)
	UpdateShopOrder(ctx context.Context, arg UpdateShopOrderParams) (ShopOrder, error)
	// -- name: ListShopOrderItemsByOrderID :many
	// SELECT * FROM "shop_order_item"
	// WHERE order_id = $1
	// ORDER BY id;
	UpdateShopOrderItem(ctx context.Context, arg UpdateShopOrderItemParams) (ShopOrderItem, error)
	UpdateShoppingCart(ctx context.Context, arg UpdateShoppingCartParams) (ShoppingCart, error)
	UpdateShoppingCartItem(ctx context.Context, arg UpdateShoppingCartItemParams) (UpdateShoppingCartItemRow, error)
	// telephone = COALESCE(sqlc.narg(telephone),telephone),
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserAddress(ctx context.Context, arg UpdateUserAddressParams) (UserAddress, error)
	UpdateUserEmailisVerifiedForTest(ctx context.Context, id int64) error
	UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (User, error)
	UpdateUserReview(ctx context.Context, arg UpdateUserReviewParams) (UserReview, error)
	UpdateUserSession(ctx context.Context, arg UpdateUserSessionParams) (UserSession, error)
	UpdateVariation(ctx context.Context, arg UpdateVariationParams) (Variation, error)
	UpdateVariationOption(ctx context.Context, arg UpdateVariationOptionParams) (VariationOption, error)
	UpdateVerifyEmail(ctx context.Context, arg UpdateVerifyEmailParams) (UpdateVerifyEmailRow, error)
	UpdateWishList(ctx context.Context, arg UpdateWishListParams) (WishList, error)
	// WITH t1 AS (
	//   SELECT user_id FROM "wish_list" AS wl
	//   WHERE wl.id = sqlc.arg(wish_list_id)
	// )
	UpdateWishListItem(ctx context.Context, arg UpdateWishListItemParams) (WishListItem, error)
}

var _ Querier = (*Queries)(nil)
