package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Memastikan pengguna sudah login.
func RequireAuth(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, SessionName)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			_, loggedIn := GetUserID(session)
			if !loggedIn {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Memastikan pengguna memiliki salah satu role yang diizinkan.
func RequireRole(
	store *sessions.CookieStore,
	allowedRoles ...string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, SessionName)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			_, loggedIn := GetUserID(session)
			if !loggedIn {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			role, ok := session.Values["role"].(string)
			if !ok {
				http.Error(w, "Role pengguna tidak valid", http.StatusForbidden)
				return
			}

			authorized := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					authorized = true
					break
				}
			}

			if !authorized {
				http.Error(w, "Anda tidak memiliki hak akses", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
