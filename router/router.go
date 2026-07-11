package router

import (
	"fotstat/router/routers"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func getArrayCommal(name string) []int64 {
	values := strings.Split(name, ",")

	var items []int64
	for _, item := range values {
        n, _ := strconv.ParseInt(item, 10, 64)
		items = append(items, n)
	}

	return items
}

func getArrayCommai(name string) []int {
	values := strings.Split(name, ",")

	var items []int
	for _, item := range values {
        n, _ := strconv.Atoi(item)
		items = append(items, n)
	}

	return items
}

func SetRouter(r *fiber.App) {

    r.Get("/api/jwt", func(c *fiber.Ctx) error {
		email := c.Query("email")
        password := c.Query("password")
        return c.JSON(JwtAuth(c, email, password))
	})

	r.Post("/api/apple-auth", AppleAuth)

	r.Post("/api/guest", GuestAuth)

	r.Post("/api/refresh", RefreshToken)

	apiGroup := r.Group("/api")

	apiGroup.Use(JwtAuthRequired)

	apiGroup.Delete("/account", DeleteAccount)
	apiGroup.Post("/account/upgrade", UpgradeAccount)
	apiGroup.Post("/logout", Logout)


	// Setup domain-specific routes
	routers.SetupUploadRoutes(apiGroup)
	routers.SetupUserRoutes(apiGroup)
	routers.SetupMatchRoutes(apiGroup)
	routers.SetupPlayerRoutes(apiGroup)
	routers.SetupInjuryRoutes(apiGroup)
	routers.SetupInbodyRoutes(apiGroup)
	routers.SetupQuarterRoutes(apiGroup)
	routers.SetupRecordRoutes(apiGroup)
	routers.SetupTrainingRoutes(apiGroup)
	routers.SetupAttendanceRoutes(apiGroup)
	routers.SetupTeamRoutes(apiGroup)
	routers.SetupReportRoutes(apiGroup)
}