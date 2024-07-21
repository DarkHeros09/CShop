package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
)

type dashboardJsonResponse struct {
	ActiveProducts int32  `json:"active_products"`
	TotalProducts  int32  `json:"total_products"`
	ActiveUsers    int32  `json:"active_users"`
	TotalUsers     int32  `json:"total_users"`
	ActiveOrders   int32  `json:"active_orders"`
	TotalOrders    int32  `json:"total_orders"`
	TodayRevenue   string `json:"today_revenue"`
	MonthRevenue   int32  `json:"total_revenue"`
}

type dashboardParamsResquest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

func (server *Server) getDashboardInfo(ctx *fiber.Ctx) error {
	params := &dashboardParamsResquest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}
	// active products query
	activeProducts, err := server.store.GetActiveProductItems(ctx.Context(), params.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "active products query error"})
		return nil
	}
	// total products query
	totalProducts, err := server.store.GetTotalProductItems(ctx.Context(), params.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "total products query error"})
		return nil
	}
	// active users query
	activeUsers, err := server.store.GetActiveUsersCount(ctx.Context(), params.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "active users query error"})
		return nil
	}
	// total users query
	totalUsers, err := server.store.GetTotalUsersCount(ctx.Context(), params.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "total users query error"})
		return nil
	}

	// active orders query
	arg := db.GetShopOrdersCountByStatusIdParams{
		OrderStatusID: null.IntFrom(2),
		AdminID:       params.AdminID,
	}
	activeOrders, err := server.store.GetShopOrdersCountByStatusId(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "active orders query error"})
		return nil
	}

	// total orders query
	totalOrders, err := server.store.GetTotalShopOrder(ctx.Context(), params.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "total orders query error"})
		return nil
	}

	// today revenue query
	dailyCompletedOrderTotal, err := server.store.GetCompletedDailyOrderTotal(ctx.Context(), params.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "today revenue query error"})
		return nil
	}

	// month revenue query
	rsp := dashboardJsonResponse{
		ActiveProducts: int32(activeProducts),
		TotalProducts:  int32(totalProducts),
		ActiveUsers:    int32(activeUsers),
		TotalUsers:     int32(totalUsers),
		ActiveOrders:   int32(activeOrders),
		TotalOrders:    int32(totalOrders),
		TodayRevenue:   dailyCompletedOrderTotal,
		MonthRevenue:   0,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)

	return nil
}
