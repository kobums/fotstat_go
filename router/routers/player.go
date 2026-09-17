package routers

import (

	"strconv"

	"fotstat/global/log"


	"fotstat/controllers/rest"


	"fotstat/models"
	"github.com/gofiber/fiber/v2"
)

// SetupPlayerRoutes sets up routes for player domain
func SetupPlayerRoutes(group fiber.Router) {

	group.Get("/player", func(c *fiber.Ctx) error {
		page_, _ := strconv.Atoi(c.Query("page"))
		pagesize_, _ := strconv.Atoi(c.Query("pagesize"))
		var controller rest.PlayerController
		controller.Init(c)
		controller.Index(page_, pagesize_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Get("/player/:id", func(c *fiber.Ctx) error {
		id_, _ := strconv.ParseInt(c.Params("id"), 10, 64)
		var controller rest.PlayerController
		controller.Init(c)
		controller.Read(id_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	// GET /api/player/:id/stats?start=YYYY-MM-DD&end=YYYY-MM-DD
	// 선수 상세 통계 — 요약·스쿼드 집계·경기별 기록·부상 이력·훈련 참석을 한 번에.
	// 다른 JSON 라우트와 같이 항상 200 + 바디 code("ok"/"error")로 응답한다 —
	// iOS APIClient 는 비2xx 응답의 바디를 읽지 않아 메시지가 사라진다.
	group.Get("/player/:id/stats", func(c *fiber.Ctx) error {
		id_, _ := strconv.ParseInt(c.Params("id"), 10, 64)
		var controller rest.PlayerStatsController
		controller.Init(c)
		controller.Stats(id_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/player", func(c *fiber.Ctx) error {
		item_ := &models.Player{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.PlayerController
		controller.Init(c)
		controller.Insert(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/player/batch", func(c *fiber.Ctx) error {
		var items_ *[]models.Player
		items__ref := &items_
		err := c.BodyParser(items__ref)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.PlayerController
		controller.Init(c)
		controller.Insertbatch(items_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/player/count", func(c *fiber.Ctx) error {

		var controller rest.PlayerController
		controller.Init(c)
		controller.Count()
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Put("/player", func(c *fiber.Ctx) error {
		item_ := &models.Player{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.PlayerController
		controller.Init(c)
		controller.Update(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/player", func(c *fiber.Ctx) error {
		item_ := &models.Player{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.PlayerController
		controller.Init(c)
		controller.Delete(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/player/batch", func(c *fiber.Ctx) error {
		item_ := &[]models.Player{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.PlayerController
		controller.Init(c)
		controller.Deletebatch(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

}