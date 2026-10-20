package router

import (
	"fotstat/global/jwt"
	"fotstat/global/log"
	"fotstat/models"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

var JwtAuthRequired = func(c *fiber.Ctx) error {
	if values := c.Get("Authorization"); len(values) > 0 {
		str := values

		claims, err := jwt.Check(str)
		if err == nil {
			user := (*claims).User
			user.Password = ""
			c.Locals("user", &user)
			return c.Next()
		}
	}

	path := c.Path()
	u, _ := url.Parse(path)

	if u.Path == "/api/jwt" {
		return c.Next()
	}

	// 비로그인 가입은 POST /api/user 만 허용한다 (예전 접두사 비교는 /api/user/batch,
	// /api/user/count 까지 무인증으로 열어 이메일 존재 여부 조회가 가능했다).
	if c.Method() == "POST" && u.Path == "/api/user" {
		return c.Next()
	}

	log.Info().Msg("Jwt header is broken")

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"code":    "error",
		"message": "unauthorized",
	})
}

// jwtLoginBody 는 POST /api/jwt 본문. JSON 과 form(x-www-form-urlencoded,
// multipart) 모두 받는다.
type jwtLoginBody struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

// JwtLogin 은 POST /api/jwt 핸들러. 파싱 실패 시 에러 원문(본문 일부가 섞일 수 있음)을
// 로그나 응답에 싣지 않고 고정 메시지만 돌려준다.
func JwtLogin(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	var body jwtLoginBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "error",
			"message": "invalid request body",
		})
	}
	return c.JSON(JwtAuth(c, body.Email, body.Password))
}

func JwtAuth(c *fiber.Ctx, email string, password string) map[string]interface{} {
	// 소셜 로그인은 현재 User 모델에 ConnectId 컬럼이 없으므로 임시로 제외했습니다.
	// 추후 DB 구조 확장이 필요합니다.

	conn := models.NewConnection()
	defer conn.Close()

	userManager := models.NewUserManager(conn)

	// 이메일로 유저 검색 (GetByEmail 대신 GetWhere 사용)
	var args []interface{}
	args = append(args, models.Where{Column: "email", Value: email, Compare: "="})
	item := userManager.GetWhere(args)

	if item == nil {
		return map[string]interface{}{
			"code":    "error",
			"message": "user not found",
		}
	}

	// 비밀번호 확인 (Passwd -> Password 컬럼명 변경 적용)
	if !jwt.CheckPasswd(item.Password, password) {
		return map[string]interface{}{
			"code":    "error",
			"message": "wrong password",
		}
	}

	// 로그인 성공 처리
	// gym 프로젝트에 있던 ipblock, systemlog, loginlog 관련 로직은
	// fotstat 프로젝트에 해당 테이블들이 없으므로 제거했습니다.

	item.Password = ""

	token := jwt.MakeToken(*item)

	// Issue a long-lived refresh token so the client can stay signed in after
	// the access JWT expires. Best-effort: never block login if it fails.
	refresh, err := models.CreateRefreshToken(conn, item.Id)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("JwtAuth: create refresh token")
	}

	resp := map[string]interface{}{
		"code":  "ok",
		"token": token,
		"user":  item,
	}
	// Only advertise a refresh token when one was actually issued; an empty
	// string would otherwise be stored by the client as a useless token.
	if refresh != "" {
		resp["refresh"] = refresh
	}
	return resp
}
