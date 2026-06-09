package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"fotstat/global/jwt"
	"fotstat/global/log"
	"fotstat/models"
)

// GuestAuth creates an anonymous (guest) account and returns a JWT, so users
// can access the app's features without entering any personal information.
// Required by App Store guideline 5.1.1(v): non-account-based features must be
// usable without forcing registration. A guest may later upgrade to a full
// account by registering or signing in with Apple.
func GuestAuth(c *fiber.Ctx) error {
	conn := models.NewConnection()
	defer conn.Close()

	userManager := models.NewUserManager(conn)

	// A unique, non-personal identifier. The "guest:" prefix lets us recognise
	// anonymous accounts; no email, password or name is collected.
	email := fmt.Sprintf("guest:%s", uuid.NewString())

	newUser := &models.UserUpdate{
		Email:    email,
		Password: "",
		Name:     "게스트",
	}

	if err := userManager.Insert(newUser); err != nil {
		log.Error().Str("error", err.Error()).Msg("Guest auth: insert user")
		return c.JSON(fiber.Map{"code": "error", "message": "게스트 시작에 실패했습니다"})
	}

	user := &models.User{
		Id:    userManager.GetIdentity(),
		Email: email,
		Name:  "게스트",
	}

	token := jwt.MakeToken(*user)

	return c.JSON(fiber.Map{
		"code":  "ok",
		"token": token,
		"user":  user,
	})
}
