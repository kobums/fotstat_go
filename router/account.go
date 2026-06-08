package router

import (
	"fotstat/global/apple"
	"fotstat/global/config"
	"fotstat/global/log"
	"fotstat/models"

	"github.com/gofiber/fiber/v2"
)

// DeleteAccount removes the authenticated user's account along with all
// data owned by them. Foreign keys are declared ON DELETE CASCADE, so deleting
// the user_tb row cascades to teams, players, matches, quarters and records.
func DeleteAccount(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "error",
			"message": "unauthorized",
		})
	}

	conn := models.NewConnection()
	defer conn.Close()

	// Revoke the Apple refresh token before deleting, so the Sign in with Apple
	// link is fully severed (App Store guideline 5.1.1(v)). Best-effort: a
	// revoke failure must not prevent the user from deleting their account.
	if config.AppleConfigured() {
		if refresh, err := models.GetAppleRefreshToken(conn, user.Id); err != nil {
			log.Error().Str("error", err.Error()).Msg("DeleteAccount: get refresh token")
		} else if refresh != "" {
			if err := apple.Revoke(refresh); err != nil {
				log.Error().Str("error", err.Error()).Msg("DeleteAccount: apple revoke")
			}
		}
	}

	userManager := models.NewUserManager(conn)
	if err := userManager.Delete(user.Id); err != nil {
		log.Error().Str("error", err.Error()).Msg("DeleteAccount")
		return c.JSON(fiber.Map{"code": "error", "message": "계정 삭제에 실패했습니다"})
	}

	return c.JSON(fiber.Map{"code": "ok"})
}
