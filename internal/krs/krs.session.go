package krs

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/kresnakar/sistem-krs/internal/auth"
)

// sessionNIM mengambil NIM mahasiswa dari session login.
func sessionNIM(r *http.Request, store *sessions.CookieStore) (string, bool) {
	session, err := store.Get(r, auth.SessionName)
	if err != nil {
		return "", false
	}

	nim, ok := session.Values["nim"].(string)
	if !ok || strings.TrimSpace(nim) == "" {
		return "", false
	}

	return strings.TrimSpace(nim), true
}

// sessionDosenWaliID mengambil ID dosen wali dari session login.
func sessionDosenWaliID(r *http.Request, store *sessions.CookieStore) (int, bool) {
	session, err := store.Get(r, auth.SessionName)
	if err != nil {
		return 0, false
	}

	return toInt(session.Values["dosen_wali_id"])
}

// toInt menangani tipe angka yang mungkin berbeda setelah decode session.
func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
