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

	// 로그인 — 자격증명을 본문(JSON 또는 form)으로 받는다.
	r.Post("/api/jwt", JwtLogin)

	// Deprecated: GET /api/jwt?email=&password= 는 비밀번호가 URL(쿼리스트링)에
	// 실려 프록시·접근 로그·브라우저 히스토리에 남을 수 있다. 이미 배포된 iOS 앱
	// 호환을 위해서만 남겨 두며, 구버전 앱이 사라지면 제거한다.
	// 요청 로그는 path 만 기록하므로(services/http.go) 쿼리스트링은 로그에 남지 않는다.
	r.Get("/api/jwt", func(c *fiber.Ctx) error {
		c.Set("Deprecation", "true")
		c.Set("Cache-Control", "no-store")
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