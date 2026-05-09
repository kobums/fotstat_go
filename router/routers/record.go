package routers

import (

	"strconv"

	"fotstat/global/log"


	"fotstat/controllers/rest"


	"fotstat/models"
	"github.com/gofiber/fiber/v2"
)

// SetupRecordRoutes sets up routes for record domain
func SetupRecordRoutes(group fiber.Router) {

	group.Get("/record", func(c *fiber.Ctx) error {
		page_, _ := strconv.Atoi(c.Query("page"))
		pagesize_, _ := strconv.Atoi(c.Query("pagesize"))
		var controller rest.RecordController
		controller.Init(c)
		controller.Index(page_, pagesize_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Get("/record/:id", func(c *fiber.Ctx) error {
		id_, _ := strconv.ParseInt(c.Params("id"), 10, 64)
		var controller rest.RecordController
		controller.Init(c)
		controller.Read(id_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/record", func(c *fiber.Ctx) error {
		item_ := &models.Record{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.RecordController
		controller.Init(c)
		controller.Insert(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/record/batch", func(c *fiber.Ctx) error {
		var items_ *[]models.Record
		items__ref := &items_
		err := c.BodyParser(items__ref)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.RecordController
		controller.Init(c)
		controller.Insertbatch(items_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/record/count", func(c *fiber.Ctx) error {

		var controller rest.RecordController
		controller.Init(c)
		controller.Count()
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Put("/record/stats", func(c *fiber.Ctx) error {
		item_ := &models.Record{}
		err := c.BodyParser(item_)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		var controller rest.RecordController
		controller.Init(c)
		controller.UpdateStats(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Put("/record", func(c *fiber.Ctx) error {
		item_ := &models.Record{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.RecordController
		controller.Init(c)
		controller.Update(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/record", func(c *fiber.Ctx) error {
		item_ := &models.Record{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.RecordController
		controller.Init(c)
		controller.Delete(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/record/batch", func(c *fiber.Ctx) error {
		item_ := &[]models.Record{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.RecordController
		controller.Init(c)
		controller.Deletebatch(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

}