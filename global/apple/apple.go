package apple

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	gjwt "github.com/dgrijalva/jwt-go/v4"

	"fotstat/global/config"
)

const (
	tokenURL  = "https://appleid.apple.com/auth/token"
	revokeURL = "https://appleid.apple.com/auth/revoke"
	audience  = "https://appleid.apple.com"
)

// parsePrivateKey decodes the PEM-encoded PKCS8 EC private key (.p8) from config.
func parsePrivateKey() (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(config.Apple.PrivateKey))
	if block == nil {
		return nil, errors.New("apple: invalid private key PEM")
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apple: parse private key: %w", err)
	}

	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("apple: private key is not ECDSA")
	}

	return key, nil
}

// clientSecret builds the short-lived JWT used as client_secret in token
// exchange / revocation requests, signed with ES256 using the .p8 key.
func clientSecret() (string, error) {
	key, err := parsePrivateKey()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := gjwt.MapClaims{
		"iss": config.Apple.TeamID,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": audience,
		"sub": config.Apple.ClientID,
	}

	token := gjwt.NewWithClaims(gjwt.SigningMethodES256, claims)
	token.Header["kid"] = config.Apple.KeyID
	token.Header["alg"] = "ES256"

	return token.SignedString(key)
}

// ExchangeCode trades an authorization code (from Sign in with Apple) for a
// refresh token. The refresh token is stored so the account can later be
// revoked on deletion.
func ExchangeCode(authorizationCode string) (refreshToken string, err error) {
	secret, err := clientSecret()
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("client_id", config.Apple.ClientID)
	form.Set("client_secret", secret)
	form.Set("code", authorizationCode)
	form.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("apple token exchange failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Minimal parse to avoid pulling in a struct: only refresh_token is needed.
	var parsed struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.RefreshToken == "" {
		return "", errors.New("apple: no refresh_token in response")
	}

	return parsed.RefreshToken, nil
}

// Revoke invalidates the user's Apple refresh token, fully severing the Sign
// in with Apple link as required for account deletion.
func Revoke(refreshToken string) error {
	secret, err := clientSecret()
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Set("client_id", config.Apple.ClientID)
	form.Set("client_secret", secret)
	form.Set("token", refreshToken)
	form.Set("token_type_hint", "refresh_token")

	resp, err := http.PostForm(revokeURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("apple revoke failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
