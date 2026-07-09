package routers

import (
	"net/http"
	"net/url"
	"strconv"

	"fotstat/controllers/rest"

	"github.com/gofiber/fiber/v2"
)

// SetupReportRoutes 는 리포트(엑셀 다운로드) 라우트를 등록한다.
func SetupReportRoutes(group fiber.Router) {

	// GET /api/report/matchrecord?team=<id>&start=<YYYY-MM-DD>&end=<YYYY-MM-DD>
	// 경기기록표 xlsx 를 스트리밍한다(웹·iOS 공용).
	group.Get("/report/matchrecord", func(c *fiber.Ctx) error {
		var controller rest.ReportController
		controller.Init(c)
		defer controller.Close()

		data, filename, ok := controller.BuildMatchRecord()
		if !ok {
			status := controller.Code
			if status == http.StatusOK {
				status = http.StatusBadRequest
			}
			return c.Status(status).JSON(controller.Result)
		}

		// filename* 로 UTF-8 파일명(한글) 을 전달하고, 구형 클라이언트용 ASCII fallback 도 둔다.
		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename=\"report.xlsx\"; filename*=UTF-8''"+url.PathEscape(filename))
		c.Set("Content-Length", strconv.Itoa(len(data)))
		return c.Send(data)
	})
}
