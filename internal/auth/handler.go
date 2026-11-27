package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "web/login.html")
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Form tidak valid", http.StatusBadRequest)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		if username == "" || password == "" {
			http.Error(w, "Username dan password wajib diisi", http.StatusBadRequest)
			return
		}

		user, err := GetUserByUsername(db, username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Username atau password salah", http.StatusUnauthorized)
				return
			}

			http.Error(w, "Terjadi kesalahan server", http.StatusInternalServerError)
			return
		}

		if !user.IsActive {
			http.Error(w, "Akun tidak aktif", http.StatusForbidden)
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.PasswordHash),
			[]byte(password),
		); err != nil {
			http.Error(w, "Username atau password salah", http.StatusUnauthorized)
			return
		}

		session, err := store.Get(r, SessionName)
		if err != nil {
			http.Error(w, "Gagal membuat sesi login", http.StatusInternalServerError)
			return
		}

		session.Values["user_id"] = user.ID
		session.Values["username"] = user.Username
		session.Values["role"] = user.Role

		if user.NIM.Valid {
			session.Values["nim"] = user.NIM.String
		}

		if user.DosenWaliID.Valid {
			session.Values["dosen_wali_id"] = user.DosenWaliID.Int64
		}

		if err := session.Save(r, w); err != nil {
			http.Error(w, "Gagal menyimpan sesi login", http.StatusInternalServerError)
			return
		}

		// Tentukan halaman berdasarkan role pengguna
		var destination string

		switch user.Role {
		case "Mahasiswa":
			destination = "/mahasiswa"

		case "Dosen Wali":
			destination = "/dosen"

		case "Admin Prodi":
			destination = "/admin"

		case "Kaprodi":
			destination = "/kaprodi"

		default:
			http.Error(w, "Role pengguna tidak dikenali", http.StatusForbidden)
			return
		}

		http.Redirect(w, r, destination, http.StatusSeeOther)
	}
}

func LogoutHandler(store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
			return
		}

		session, err := store.Get(r, SessionName)
		if err != nil {
			http.Error(w, "Gagal membaca sesi", http.StatusInternalServerError)
			return
		}

		session.Values = make(map[interface{}]interface{})
		session.Options.MaxAge = -1

		if err := session.Save(r, w); err != nil {
			http.Error(w, "Gagal menghapus sesi", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// Helper untuk membaca ID pengguna dari sesi.
func GetUserID(session *sessions.Session) (int, bool) {
	value, ok := session.Values["user_id"]
	if !ok {
		return 0, false
	}

	switch id := value.(type) {
	case int:
		return id, true
	case int64:
		return int(id), true
	case string:
		parsed, err := strconv.Atoi(id)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
