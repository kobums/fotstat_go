package router

import (
	"strings"

	"fotstat/global/apple"
	"fotstat/global/config"
	"fotstat/global/jwt"
	"fotstat/global/log"
	"fotstat/models"

	"github.com/gofiber/fiber/v2"
)

// UpgradeAccount converts the authenticated guest account into a full account
// in place, keeping the same user id so all of the guest's data (teams,
// players, matches, records) is preserved. Only guest accounts may upgrade.
func UpgradeAccount(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "error",
			"message": "unauthorized",
		})
	}

	if !strings.HasPrefix(user.Email, "guest:") {
		return c.JSON(fiber.Map{"code": "error", "message": "이미 가입된 계정입니다"})
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.JSON(fiber.Map{"code": "error", "message": "잘못된 요청입니다"})
	}

	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" || body.Password == "" {
		return c.JSON(fiber.Map{"code": "error", "message": "이메일과 비밀번호를 입력해주세요"})
	}

	conn := models.NewConnection()
	defer conn.Close()

	userManager := models.NewUserManager(conn)

	// Reject if the email is already used by a different account.
	var args []interface{}
	args = append(args, models.Where{Column: "email", Value: body.Email, Compare: "="})
	if existing := userManager.GetWhere(args); existing != nil && existing.Id != user.Id {
		return c.JSON(fiber.Map{"code": "error", "message": "이미 사용 중인 이메일입니다"})
	}

	hashed, err := jwt.GeneratePasswd(body.Password)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("UpgradeAccount: hash password")
		return c.JSON(fiber.Map{"code": "error", "message": "가입에 실패했습니다"})
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "사용자"
	}

	if err := userManager.UpdateEmail(body.Email, user.Id); err != nil {
		log.Error().Str("error", err.Error()).Msg("UpgradeAccount: update email")
		return c.JSON(fiber.Map{"code": "error", "message": "가입에 실패했습니다"})
	}
	if err := userManager.UpdatePassword(hashed, user.Id); err != nil {
		log.Error().Str("error", err.Error()).Msg("UpgradeAccount: update password")
		return c.JSON(fiber.Map{"code": "error", "message": "가입에 실패했습니다"})
	}
	if err := userManager.UpdateName(name, user.Id); err != nil {
		log.Error().Str("error", err.Error()).Msg("UpgradeAccount: update name")
		return c.JSON(fiber.Map{"code": "error", "message": "가입에 실패했습니다"})
	}

	// The JWT embeds the user's email/name, so re-issue it after the upgrade.
	updated := &models.User{Id: user.Id, Email: body.Email, Name: name}
	token := jwt.MakeToken(*updated)

	// Revoke any refresh tokens issued while this was a guest account, then mint
	// a fresh one for the upgraded session so old guest tokens can't be reused.
	if err := models.DeleteUserRefreshTokens(conn, updated.Id); err != nil {
		log.Error().Str("error", err.Error()).Msg("UpgradeAccount: revoke old refresh tokens")
	}
	refresh, err := models.CreateRefreshToken(conn, updated.Id)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("UpgradeAccount: create refresh token")
	}

	return c.JSON(fiber.Map{
		"code":    "ok",
		"token":   token,
		"refresh": refresh,
		"user":    updated,
	})
}

// Logout revokes the authenticated user's refresh tokens server-side so a
// stolen or leaked token can no longer renew a session after sign-out. The
// client still discards its local copies; this severs the server end too.
func Logout(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": "error", "message": "unauthorized",
		})
	}

	conn := models.NewConnection()
	defer conn.Close()

	if err := models.DeleteUserRefreshTokens(conn, user.Id); err != nil {
		log.Error().Str("error", err.Error()).Msg("Logout: revoke refresh tokens")
		return c.JSON(fiber.Map{"code": "error", "message": "로그아웃에 실패했습니다"})
	}

	return c.JSON(fiber.Map{"code": "ok"})
}

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
