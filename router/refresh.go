package router

import (
	"github.com/gofiber/fiber/v2"

	"fotstat/global/jwt"
	"fotstat/global/log"
	"fotstat/models"
)

// RefreshToken exchanges a valid refresh token for a fresh access JWT, so the
// app can keep the user signed in after the short-lived access token expires.
// It sits outside the JwtAuthRequired group because the access token is, by
// definition, already expired when this is called.
func RefreshToken(c *fiber.Ctx) error {
	var body struct {
		Refresh string `json:"refresh"`
	}
	if err := c.BodyParser(&body); err != nil || body.Refresh == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": "error", "message": "refresh token required",
		})
	}

	conn := models.NewConnection()
	defer conn.Close()

	user, err := models.LookupRefreshToken(conn, body.Refresh)
	if err != nil {
		// A server-side failure (e.g. DB down) is retryable and must not be
		// confused with an invalid token, so return 500 rather than 401/200.
		log.Error().Str("error", err.Error()).Msg("RefreshToken: lookup")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": "error", "message": "토큰 갱신에 실패했습니다",
		})
	}
	if user == nil {
		// Unknown or expired token: the client must sign in again.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": "error", "message": "세션이 만료되었습니다. 다시 로그인해주세요",
		})
	}

	user.Password = ""
	token := jwt.MakeToken(*user)

	// The refresh token is non-rotating (its expiry is slid forward on use), so
	// echo it back unchanged for the client to keep.
	return c.JSON(fiber.Map{
		"code":    "ok",
		"token":   token,
		"refresh": body.Refresh,
		"user":    user,
	})
}
