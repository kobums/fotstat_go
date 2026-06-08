package models

import "database/sql"

// SaveAppleRefreshToken upserts the Apple refresh token for a user. It is kept
// in a dedicated table (apple_token_tb) so the code-generated user model is not
// affected. The row is removed automatically when the user is deleted via the
// ON DELETE CASCADE foreign key.
func SaveAppleRefreshToken(conn *Connection, userId int64, refreshToken string) error {
	if !conn.IsConnect() {
		return sql.ErrConnDone
	}

	query := `INSERT INTO apple_token_tb (at_user, at_refresh)
	          VALUES (?, ?)
	          ON DUPLICATE KEY UPDATE at_refresh = VALUES(at_refresh)`
	_, err := conn.Exec(query, userId, refreshToken)
	return err
}

// GetAppleRefreshToken returns the stored Apple refresh token for a user, or an
// empty string if none exists.
func GetAppleRefreshToken(conn *Connection, userId int64) (string, error) {
	if !conn.IsConnect() {
		return "", sql.ErrConnDone
	}

	rows, err := conn.Query("SELECT at_refresh FROM apple_token_tb WHERE at_user = ?", userId)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var token string
	if rows.Next() {
		if err := rows.Scan(&token); err != nil {
			return "", err
		}
	}
	return token, nil
}
