package krs

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/kresnakar/sistem-krs/internal/auth"
)

type Handler struct {
	service *Service
	store   *sessions.CookieStore
}

func NewHandler(
	service *Service,
	store *sessions.CookieStore,
) *Handler {
	return &Handler{
		service: service,
		store:   store,
	}
}

func (h *Handler) ambilNIMSession(
	r *http.Request,
) (string, bool) {
	session, err := h.store.Get(r, auth.SessionName)
	if err != nil {
		return "", false
	}

	nim, ok := session.Values["nim"].(string)
	if !ok || nim == "" {
		return "", false
	}

	return nim, true
}

// Mendaftarkan endpoint API KRS
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/krs/{semester_id}", h.AmbilKRS)
	mux.HandleFunc("POST /api/krs/{krs_id}/kelas", h.TambahKelas)
	mux.HandleFunc("DELETE /api/krs/{krs_id}/kelas/{kelas_id}", h.BatalkanKelas)
}

// Mengambil data KRS mahasiswa
func (h *Handler) AmbilKRS(
	w http.ResponseWriter,
	r *http.Request,
) {
	nim, ok := h.ambilNIMSession(r)
	if !ok {
		http.Error(w, "Session mahasiswa tidak valid", http.StatusUnauthorized)
		return
	}

	semesterID, err := strconv.Atoi(
		r.PathValue("semester_id"),
	)

	if err != nil {
		http.Error(
			w,
			"Semester ID tidak valid",
			http.StatusBadRequest,
		)
		return
	}

	data, err := h.service.AmbilKRS(
		r.Context(),
		nim,
		semesterID,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Request untuk menambahkan kelas
type TambahKelasRequest struct {
	KelasID  int `json:"kelas_id"`
	BatasSKS int `json:"batas_sks"`
}

// Menambahkan kelas ke KRS
func (h *Handler) TambahKelas(
	w http.ResponseWriter,
	r *http.Request,
) {
	krsID, err := strconv.Atoi(
		r.PathValue("krs_id"),
	)

	if err != nil {
		http.Error(
			w,
			"KRS ID tidak valid",
			http.StatusBadRequest,
		)
		return
	}

	nim, ok := h.ambilNIMSession(r)
	if !ok {
		http.Error(w, "Session mahasiswa tidak valid", http.StatusUnauthorized)
		return
	}

	pemilik, err := h.service.ValidasiKepemilikanKRS(r.Context(), krsID, nim)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !pemilik {
		http.Error(w, "KRS bukan milik Anda", http.StatusForbidden)
		return
	}

	var req TambahKelasRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(
			w,
			"Format request tidak valid",
			http.StatusBadRequest,
		)
		return
	}

	if req.KelasID <= 0 {
		http.Error(
			w,
			"Kelas ID harus valid",
			http.StatusBadRequest,
		)
		return
	}

	batasSKS, err := h.service.BatasSKSKRS(r.Context(), krsID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.TambahKelas(
		r.Context(),
		krsID,
		req.KelasID,
		batasSKS,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pesan": "Kelas berhasil ditambahkan ke KRS",
	})
}

// Membatalkan kelas dari KRS
func (h *Handler) BatalkanKelas(
	w http.ResponseWriter,
	r *http.Request,
) {
	krsID, err := strconv.Atoi(
		r.PathValue("krs_id"),
	)

	if err != nil {
		http.Error(
			w,
			"KRS ID tidak valid",
			http.StatusBadRequest,
		)
		return
	}

	kelasID, err := strconv.Atoi(
		r.PathValue("kelas_id"),
	)

	if err != nil {
		http.Error(
			w,
			"Kelas ID tidak valid",
			http.StatusBadRequest,
		)
		return
	}

	nim, ok := h.ambilNIMSession(r)
	if !ok {
		http.Error(w, "Session mahasiswa tidak valid", http.StatusUnauthorized)
		return
	}

	pemilik, err := h.service.ValidasiKepemilikanKRS(r.Context(), krsID, nim)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !pemilik {
		http.Error(w, "KRS bukan milik Anda", http.StatusForbidden)
		return
	}

	err = h.service.BatalkanKelas(
		r.Context(),
		krsID,
		kelasID,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pesan": "Kelas berhasil dibatalkan",
	})
}
