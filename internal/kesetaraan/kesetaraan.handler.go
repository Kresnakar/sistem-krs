package kesetaraan

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/kesetaraan", h.GetAll)
	mux.HandleFunc("GET /api/kesetaraan/{kode_mk}", h.GetSetaraByKode)
	mux.HandleFunc("POST /api/kesetaraan", h.Create)
	mux.HandleFunc("DELETE /api/kesetaraan/{mk_lama}/{mk_baru}", h.Delete)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	data, err := h.repo.GetAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) GetSetaraByKode(w http.ResponseWriter, r *http.Request) {
	kodeMK := strings.TrimSpace(r.PathValue("kode_mk"))
	daftar, err := h.repo.GetSetaraCodes(r.Context(), kodeMK)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kode_mk":       kodeMK,
		"daftar_setara": daftar,
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req KesetaraanMK
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format JSON tidak valid"})
		return
	}

	req.MKLama = strings.TrimSpace(req.MKLama)
	req.MKBaru = strings.TrimSpace(req.MKBaru)
	req.Keterangan = strings.TrimSpace(req.Keterangan)

	if req.MKLama == "" || req.MKBaru == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mk_lama dan mk_baru wajib diisi"})
		return
	}

	err := h.repo.Create(r.Context(), req)
	if errors.Is(err, ErrSelfEqual) || errors.Is(err, ErrDuplicate) || errors.Is(err, ErrMKNotFound) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	mkLama := r.PathValue("mk_lama")
	mkBaru := r.PathValue("mk_baru")

	err := h.repo.Delete(r.Context(), mkLama, mkBaru)
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "kesetaraan mata kuliah berhasil dihapus"})
}