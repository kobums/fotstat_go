package router

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	gjwt "github.com/dgrijalva/jwt-go/v4"
	"github.com/gofiber/fiber/v2"

	"fotstat/global/apple"
	"fotstat/global/config"
	"fotstat/global/jwt"
	"fotstat/global/log"
	"fotstat/models"
)

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleJWKSet struct {
	Keys []appleJWK `json:"keys"`
}

type appleClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	gjwt.StandardClaims
}

// appleBundleID is the audience (aud) value Apple sets in the identity token.
// It must match the app's Bundle ID (or Service ID for web sign-in).
const (
	appleBundleID = "com.gowoobro.fotstat"
	appleIssuer   = "https://appleid.apple.com"
)

var (
	appleKeyCache     *appleJWKSet
	appleKeyCacheAt   time.Time
	appleKeyCacheMu   sync.Mutex
)

func getAppleKeys() (*appleJWKSet, error) {
	appleKeyCacheMu.Lock()
	defer appleKeyCacheMu.Unlock()

	if appleKeyCache != nil && time.Since(appleKeyCacheAt) < time.Hour {
		return appleKeyCache, nil
	}

	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var keySet appleJWKSet
	if err := json.Unmarshal(body, &keySet); err != nil {
		return nil, err
	}

	appleKeyCache = &keySet
	appleKeyCacheAt = time.Now()
	return &keySet, nil
}

func verifyAppleToken(identityToken string) (*appleClaims, error) {
	keySet, err := getAppleKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple public keys: %w", err)
	}

	token, err := gjwt.ParseWithClaims(identityToken, &appleClaims{}, func(token *gjwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*gjwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, _ := token.Header["kid"].(string)

		for _, key := range keySet.Keys {
			if key.Kid == kid {
				nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
				if err != nil {
					return nil, err
				}
				eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
				if err != nil {
					return nil, err
				}
				return &rsa.PublicKey{
					N: new(big.Int).SetBytes(nBytes),
					E: int(new(big.Int).SetBytes(eBytes).Int64()),
				}, nil
			}
		}

		return nil, fmt.Errorf("matching key not found for kid: %s", kid)
	}, gjwt.WithAudience(appleBundleID), gjwt.WithIssuer(appleIssuer))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*appleClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func AppleAuth(c *fiber.Ctx) error {
	var body struct {
		IdentityToken     string `json:"identityToken"`
		AuthorizationCode string `json:"authorizationCode"`
		Name              string `json:"name"`
	}

	if err := c.BodyParser(&body); err != nil || body.IdentityToken == "" {
		return c.JSON(fiber.Map{"code": "error", "message": "identityToken required"})
	}

	claims, err := verifyAppleToken(body.IdentityToken)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Apple auth")
		return c.JSON(fiber.Map{"code": "error", "message": "Apple 인증에 실패했습니다"})
	}

	email := claims.Email
	if email == "" {
		email = fmt.Sprintf("apple:%s", claims.Sub)
	}

	conn := models.NewConnection()
	defer conn.Close()

	userManager := models.NewUserManager(conn)

	var args []interface{}
	args = append(args, models.Where{Column: "email", Value: email, Compare: "="})
	user := userManager.GetWhere(args)

	if user != nil {
		// Apple only returns the real name on the first authorization. If we now
		// received a real name but the stored one is empty or the default
		// placeholder, backfill it onto the existing user.
		if body.Name != "" && (user.Name == "" || user.Name == "Apple 사용자") {
			if err := userManager.UpdateName(body.Name, user.Id); err != nil {
				log.Error().Str("error", err.Error()).Msg("Apple auth: update name")
			} else {
				user.Name = body.Name
			}
		}
	} else {
		name := body.Name
		if name == "" {
			name = "Apple 사용자"
		}

		newUser := &models.UserUpdate{
			Email:    email,
			Password: "",
			Name:     name,
		}

		if err := userManager.Insert(newUser); err != nil {
			log.Error().Str("error", err.Error()).Msg("Apple auth: insert user")
			return c.JSON(fiber.Map{"code": "error", "message": "사용자 생성에 실패했습니다"})
		}

		id := userManager.GetIdentity()
		user = &models.User{
			Id:    id,
			Email: email,
			Name:  name,
		}
	}

	// Exchange the authorization code for a refresh token and store it so the
	// account can be revoked on deletion (required by App Store guideline
	// 5.1.1(v)). Best-effort: never block login if this fails or is unconfigured.
	if body.AuthorizationCode != "" && config.AppleConfigured() {
		if refresh, err := apple.ExchangeCode(body.AuthorizationCode); err != nil {
			log.Error().Str("error", err.Error()).Msg("Apple auth: code exchange")
		} else if err := models.SaveAppleRefreshToken(conn, user.Id, refresh); err != nil {
			log.Error().Str("error", err.Error()).Msg("Apple auth: save refresh token")
		}
	}

	user.Password = ""
	token := jwt.MakeToken(*user)

	// Issue our own long-lived refresh token (distinct from Apple's) so the app
	// can renew the access JWT without a fresh Sign in with Apple round-trip.
	refresh, err := models.CreateRefreshToken(conn, user.Id)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Apple auth: create app refresh token")
	}

	resp := fiber.Map{
		"code":  "ok",
		"token": token,
		"user":  user,
	}
	if refresh != "" {
		resp["refresh"] = refresh
	}
	return c.JSON(resp)
}
