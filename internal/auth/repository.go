package auth

import (
	"database/sql"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	NIM          sql.NullString
	DosenWaliID  sql.NullInt64
	IsActive     bool
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	query := `
		SELECT
			id,
			username,
			password_hash,
			role,
			nim,
			dosen_wali_id,
			is_active
		FROM users
		WHERE username = ?
		LIMIT 1
	`

	var user User

	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.NIM,
		&user.DosenWaliID,
		&user.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
