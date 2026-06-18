package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

// refreshTokenTTLDays is how long a refresh token stays valid. Each successful
// use slides the expiry forward by this many days, so an actively-used app
// effectively never forces the user to sign in again.
const refreshTokenTTLDays = 90

// newRefreshToken returns a cryptographically random, URL-safe opaque token.
func newRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// CreateRefreshToken issues a new refresh token for a user and stores it with a
// fresh expiry. A user may hold several tokens at once (one per device).
func CreateRefreshToken(conn *Connection, userId int64) (string, error) {
	if !conn.IsConnect() {
		return "", sql.ErrConnDone
	}

	token, err := newRefreshToken()
	if err != nil {
		return "", err
	}

	query := `INSERT INTO refresh_token_tb (rt_user, rt_token, rt_expiredate)
	          VALUES (?, ?, DATE_ADD(NOW(), INTERVAL ? DAY))`
	if _, err := conn.Exec(query, userId, token, refreshTokenTTLDays); err != nil {
		return "", err
	}
	return token, nil
}

// LookupRefreshToken validates a refresh token. If it exists and has not
// expired, the expiry is slid forward and the owning user is returned. A nil
// user (with nil error) means the token is unknown or expired — the caller
// should treat that as "re-login required", not a server error.
func LookupRefreshToken(conn *Connection, token string) (*User, error) {
	if !conn.IsConnect() {
		return nil, sql.ErrConnDone
	}
	if token == "" {
		return nil, nil
	}

	rows, err := conn.Query(
		"SELECT rt_user FROM refresh_token_tb WHERE rt_token = ? AND rt_expiredate > NOW()",
		token,
	)
	if err != nil {
		return nil, err
	}

	var userId int64
	found := rows.Next()
	if found {
		if err := rows.Scan(&userId); err != nil {
			rows.Close()
			return nil, err
		}
	}
	// Free the connection before the UPDATE/SELECT below reuse it.
	rows.Close()

	if !found {
		return nil, nil
	}

	// Slide the expiry forward so active sessions stay alive.
	if _, err := conn.Exec(
		"UPDATE refresh_token_tb SET rt_expiredate = DATE_ADD(NOW(), INTERVAL ? DAY) WHERE rt_token = ?",
		refreshTokenTTLDays, token,
	); err != nil {
		return nil, err
	}

	return NewUserManager(conn).Get(userId), nil
}

// DeleteUserRefreshTokens revokes every refresh token belonging to a user
// (e.g. on logout). User deletion handles its own cleanup via ON DELETE CASCADE.
func DeleteUserRefreshTokens(conn *Connection, userId int64) error {
	if !conn.IsConnect() {
		return sql.ErrConnDone
	}
	_, err := conn.Exec("DELETE FROM refresh_token_tb WHERE rt_user = ?", userId)
	return err
}
