package routers

import (

	"strconv"

	"fotstat/global/log"


	"fotstat/controllers/rest"


	"fotstat/models"
	"github.com/gofiber/fiber/v2"
)

// SetupAttendanceRoutes sets up routes for attendance domain
func SetupAttendanceRoutes(group fiber.Router) {

	group.Get("/attendance", func(c *fiber.Ctx) error {
		page_, _ := strconv.Atoi(c.Query("page"))
		pagesize_, _ := strconv.Atoi(c.Query("pagesize"))
		var controller rest.AttendanceController
		controller.Init(c)
		controller.Index(page_, pagesize_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Get("/attendance/:id", func(c *fiber.Ctx) error {
		id_, _ := strconv.ParseInt(c.Params("id"), 10, 64)
		var controller rest.AttendanceController
		controller.Init(c)
		controller.Read(id_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/attendance", func(c *fiber.Ctx) error {
		item_ := &models.Attendance{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.AttendanceController
		controller.Init(c)
		controller.Insert(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/attendance/batch", func(c *fiber.Ctx) error {
		var items_ *[]models.Attendance
		items__ref := &items_
		err := c.BodyParser(items__ref)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.AttendanceController
		controller.Init(c)
		controller.Insertbatch(items_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/attendance/count", func(c *fiber.Ctx) error {

		var controller rest.AttendanceController
		controller.Init(c)
		controller.Count()
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Put("/attendance", func(c *fiber.Ctx) error {
		item_ := &models.Attendance{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.AttendanceController
		controller.Init(c)
		controller.Update(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/attendance", func(c *fiber.Ctx) error {
		item_ := &models.Attendance{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.AttendanceController
		controller.Init(c)
		controller.Delete(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/attendance/batch", func(c *fiber.Ctx) error {
		item_ := &[]models.Attendance{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.AttendanceController
		controller.Init(c)
		controller.Deletebatch(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

}
