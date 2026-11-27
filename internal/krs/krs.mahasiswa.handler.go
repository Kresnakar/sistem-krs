package krs

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

// MahasiswaHandler menangani endpoint KRS untuk role Mahasiswa.
type MahasiswaHandler struct {
	service *Service
	store   *sessions.CookieStore
}

func NewMahasiswaHandler(service *Service, store *sessions.CookieStore) *MahasiswaHandler {
	return &MahasiswaHandler{service: service, store: store}
}

func (h *MahasiswaHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/krs/mahasiswa/semester-aktif", h.GetSemesterAktif)
	mux.HandleFunc("GET /api/krs/mahasiswa/{semester_id}", h.GetKRS)
	mux.HandleFunc("POST /api/krs/mahasiswa/draft", h.CreateDraft)
	mux.HandleFunc("GET /api/krs/mahasiswa/kelas/{semester_id}", h.ListKelas)
	mux.HandleFunc("POST /api/krs/mahasiswa/{krs_id}/kelas", h.TambahKelas)
	mux.HandleFunc("DELETE /api/krs/mahasiswa/{krs_id}/kelas/{kelas_id}", h.BatalkanKelas)
	mux.HandleFunc("POST /api/krs/mahasiswa/{krs_id}/ajukan", h.AjukanKRS)
	mux.HandleFunc("GET /api/krs/mahasiswa/krs/{krs_id}/perubahan", h.ListPerubahan)
	mux.HandleFunc("POST /api/krs/mahasiswa/krs/{krs_id}/perubahan", h.AjukanPerubahan)
}

func (h *MahasiswaHandler) nim(r *http.Request) (string, bool) {
	return sessionNIM(r, h.store)
}

func (h *MahasiswaHandler) GetSemesterAktif(w http.ResponseWriter, r *http.Request) {
	semester, periode, err := h.service.AmbilSemesterAktif(r.Context())
	if errors.Is(err, ErrSemesterNotFound) {
		writeError(w, http.StatusNotFound, "belum ada semester aktif")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"semester": semester,
		"periode":  periode,
	})
}

func (h *MahasiswaHandler) GetKRS(w http.ResponseWriter, r *http.Request) {
	nim, ok := h.nim(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session mahasiswa tidak valid")
		return
	}

	semesterID, err := parsePathID(r, "semester_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "semester id tidak valid")
		return
	}

	data, err := h.service.AmbilKRSLengkap(r.Context(), nim, semesterID)
	if errors.Is(err, ErrKRSNotFound) {
		writeError(w, http.StatusNotFound, "KRS belum dibuat untuk semester ini")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *MahasiswaHandler) CreateDraft(w http.ResponseWriter, r *http.Request) {
	nim, ok := h.nim(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session mahasiswa tidak valid")
		return
	}

	var req struct {
		SemesterID int `json:"semester_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.SemesterID <= 0 {
		semester, _, err := h.service.AmbilSemesterAktif(r.Context())
		if err != nil {
			writeError(w, http.StatusBadRequest, "semester_id wajib diisi atau belum ada semester aktif")
			return
		}
		req.SemesterID = semester.ID
	}

	data, err := h.service.BuatDraft(r.Context(), nim, req.SemesterID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, data)
}

func (h *MahasiswaHandler) ListKelas(w http.ResponseWriter, r *http.Request) {
	semesterID, err := parsePathID(r, "semester_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "semester id tidak valid")
		return
	}

	data, err := h.service.DaftarKelas(r.Context(), semesterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *MahasiswaHandler) TambahKelas(w http.ResponseWriter, r *http.Request) {
	nim, ok := h.nim(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session mahasiswa tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}

	var req struct {
		KelasID int `json:"kelas_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}
	if req.KelasID <= 0 {
		writeError(w, http.StatusBadRequest, "kelas_id wajib diisi")
		return
	}

	ok, err = h.service.ValidasiKepemilikanKRS(r.Context(), krsID, nim)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusForbidden, "KRS bukan milik Anda")
		return
	}

	batasSKS, err := h.service.BatasSKSKRS(r.Context(), krsID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.TambahKelas(r.Context(), krsID, req.KelasID, batasSKS); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"pesan": "kelas berhasil ditambahkan"})
}

func (h *MahasiswaHandler) BatalkanKelas(w http.ResponseWriter, r *http.Request) {
	nim, ok := h.nim(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session mahasiswa tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}
	kelasID, err := parsePathID(r, "kelas_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "kelas id tidak valid")
		return
	}

	ok, err = h.service.ValidasiKepemilikanKRS(r.Context(), krsID, nim)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusForbidden, "KRS bukan milik Anda")
		return
	}

	if err := h.service.BatalkanKelas(r.Context(), krsID, kelasID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "kelas berhasil dibatalkan"})
}

func (h *MahasiswaHandler) AjukanKRS(w http.ResponseWriter, r *http.Request) {
	nim, ok := h.nim(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session mahasiswa tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}

	if err := h.service.AjukanKRS(r.Context(), krsID, nim); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "KRS berhasil diajukan"})
}

func (h *MahasiswaHandler) ListPerubahan(w http.ResponseWriter, r *http.Request) {
	nim, ok := h.nim(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session mahasiswa tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}

	ok, err = h.service.ValidasiKepemilikanKRS(r.Context(), krsID, nim)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusForbidden, "KRS bukan milik Anda")
		return
	}

	data, err := h.service.DaftarPerubahanKRS(r.Context(), krsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *MahasiswaHandler) AjukanPerubahan(w http.ResponseWriter, r *http.Request) {
	nim, ok := h.nim(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session mahasiswa tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}

	var req AjukanPerubahanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	kelasIDs := req.KelasIDs
	if req.KelasID > 0 {
		kelasIDs = append(kelasIDs, req.KelasID)
	}

	id, err := h.service.AjukanPerubahan(
		r.Context(), krsID, nim,
		strings.TrimSpace(req.JenisPerubahan),
		req.Alasan, kelasIDs,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":    id,
		"pesan": "permintaan perubahan berhasil diajukan",
	})
}
