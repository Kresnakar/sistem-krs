package daftartunggu

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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

// RegisterRoutes mendaftarkan endpoint daftar tunggu mahasiswa.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(
		"GET /api/krs/mahasiswa/daftar-tunggu",
		h.DaftarSaya,
	)

	mux.HandleFunc(
		"POST /api/krs/mahasiswa/daftar-tunggu",
		h.MasukAntrean,
	)

	mux.HandleFunc(
		"POST /api/krs/mahasiswa/daftar-tunggu/{id}/keluar",
		h.KeluarAntrean,
	)

	mux.HandleFunc(
		"POST /api/krs/mahasiswa/daftar-tunggu/{id}/konfirmasi",
		h.KonfirmasiTawaran,
	)
}

// nimMahasiswa mengambil NIM dari session login.
func (h *Handler) nimMahasiswa(
	r *http.Request,
) (string, bool) {

	session, err := h.store.Get(r, auth.SessionName)
	if err != nil {
		return "", false
	}

	role, ok := session.Values["role"].(string)
	if !ok || role != "Mahasiswa" {
		return "", false
	}

	nim, ok := session.Values["nim"].(string)
	if !ok || strings.TrimSpace(nim) == "" {
		return "", false
	}

	return strings.TrimSpace(nim), true
}

// Helper response JSON.
func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Helper response error.
func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

// Menentukan HTTP status berdasarkan jenis error.
func (h *Handler) handleError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrNIMTidakValid),
		errors.Is(err, ErrKelasIDTidakValid),
		errors.Is(err, ErrAksiKonfirmasiTidakValid):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, ErrKelasTidakDitemukan),
		errors.Is(err, ErrAntreanTidakDitemukan),
		errors.Is(err, ErrKRSKonfirmasiTidakDitemukan):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, ErrKonfirmasiKedaluwarsa):
		writeError(w, http.StatusGone, err.Error())

	case errors.Is(err, ErrKelasTidakDibuka),
		errors.Is(err, ErrKelasBelumPenuh),
		errors.Is(err, ErrSudahMengantre),
		errors.Is(err, ErrSudahMendapatKursi),
		errors.Is(err, ErrAntreanTidakAktif),
		errors.Is(err, ErrKonfirmasiTidakTersedia),
		errors.Is(err, ErrKonfirmasiHanyaPeriodePengisian),
		errors.Is(err, ErrPeriodePengisianTutup),
		errors.Is(err, ErrKapasitasTidakKonsisten),
		errors.Is(err, ErrKonfirmasiPeriodePerubahan),
		errors.Is(err, ErrPeriodePerubahanTutup),
		errors.Is(err, ErrPermintaanPerubahanSudahAda),
		errors.Is(err, ErrPeriodeDaftarTungguTutup):
		writeError(w, http.StatusConflict, err.Error())

	case errors.Is(err, ErrTidakMemenuhiSyarat),
		errors.Is(err, ErrKelayakanKonfirmasiBerubah):
		writeError(w, http.StatusUnprocessableEntity, err.Error())

	default:
		writeError(
			w,
			http.StatusInternalServerError,
			"Terjadi kesalahan pada server",
		)
	}
}

// GET: Melihat seluruh daftar tunggu milik mahasiswa.
func (h *Handler) DaftarSaya(
	w http.ResponseWriter,
	r *http.Request,
) {
	nim, ok := h.nimMahasiswa(r)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"Session mahasiswa tidak valid",
		)
		return
	}

	data, err := h.service.DaftarSaya(
		r.Context(),
		nim,
	)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// POST: Mendaftarkan mahasiswa ke daftar tunggu.
func (h *Handler) MasukAntrean(
	w http.ResponseWriter,
	r *http.Request,
) {
	nim, ok := h.nimMahasiswa(r)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"Session mahasiswa tidak valid",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req struct {
		KelasID int `json:"kelas_id"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Format request tidak valid",
		)
		return
	}

	data, err := h.service.MasukAntrean(
		r.Context(),
		nim,
		req.KelasID,
	)

	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, data)
}

// POST: Mengeluarkan mahasiswa dari antreannya sendiri.
func (h *Handler) KeluarAntrean(
	w http.ResponseWriter,
	r *http.Request,
) {
	nim, ok := h.nimMahasiswa(r)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"Session mahasiswa tidak valid",
		)
		return
	}

	antreanID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || antreanID <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"ID antrean tidak valid",
		)
		return
	}

	err = h.service.KeluarAntrean(
		r.Context(),
		nim,
		antreanID,
	)

	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"pesan": "Berhasil keluar dari daftar tunggu",
	})
}

func (h *Handler) KonfirmasiTawaran(
	w http.ResponseWriter,
	r *http.Request,
) {
	// 1. Pastikan pengguna adalah mahasiswa yang login.
	nim, ok := h.nimMahasiswa(r)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"Session mahasiswa tidak valid",
		)
		return
	}

	// 2. Ambil ID antrean dari URL.
	antreanID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || antreanID <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"ID antrean tidak valid",
		)
		return
	}

	// 3. Batasi ukuran request body.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// 4. Baca aksi dari request JSON.
	var req KonfirmasiRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"Format request tidak valid",
		)
		return
	}

	// 5. Proses konfirmasi melalui service.
	err = h.service.KonfirmasiTawaran(
		r.Context(),
		nim,
		antreanID,
		req.Aksi,
	)

	if err != nil {
		h.handleError(w, err)
		return
	}

	// 6. Kirim respons sesuai aksi.
	pesan := "Tawaran kursi berhasil dikonfirmasi"

	if strings.EqualFold(req.Aksi, "tolak") {
		pesan = "Tawaran kursi berhasil ditolak"
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"pesan": pesan,
	})
}
