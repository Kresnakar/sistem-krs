package auth

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/gorilla/sessions"
)

const SessionName = "sistem_krs_session"

func NewSessionStore() (*sessions.CookieStore, error) {
	authKey, err := base64.StdEncoding.DecodeString(
		os.Getenv("SESSION_AUTH_KEY"),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca SESSION_AUTH_KEY: %w", err)
	}

	encryptionKey, err := base64.StdEncoding.DecodeString(
		os.Getenv("SESSION_ENCRYPTION_KEY"),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca SESSION_ENCRYPTION_KEY: %w", err)
	}

	if len(authKey) < 32 {
		return nil, fmt.Errorf("SESSION_AUTH_KEY minimal 32 byte")
	}

	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("SESSION_ENCRYPTION_KEY harus 32 byte")
	}

	store := sessions.NewCookieStore(authKey, encryptionKey)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 8,
		HttpOnly: true,
		Secure:   false, // Untuk pengembangan lokal HTTP
		SameSite: 2,     // SameSite Lax
	}

	return store, nil
}
