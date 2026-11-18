package matakuliah

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
	mux.HandleFunc("GET /api/mata-kuliah", h.GetAll)
	mux.HandleFunc("GET /api/mata-kuliah/{kode_mk}", h.GetByKode)
	mux.HandleFunc("POST /api/mata-kuliah", h.Create)
	mux.HandleFunc("PUT /api/mata-kuliah/{kode_mk}", h.Update)
	mux.HandleFunc("DELETE /api/mata-kuliah/{kode_mk}", h.Delete)
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

func (h *Handler) GetByKode(w http.ResponseWriter, r *http.Request) {
	kodeMK := r.PathValue("kode_mk")
	data, err := h.repo.GetByKode(r.Context(), kodeMK)
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req MataKuliah
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format JSON tidak valid"})
		return
	}

	req.KodeMK = strings.TrimSpace(req.KodeMK)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Kurikulum = strings.TrimSpace(req.Kurikulum)

	if req.KodeMK == "" || req.Nama == "" || req.Kurikulum == "" || req.SKS <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kode_mk, nama, kurikulum wajib diisi dan sks harus > 0"})
		return
	}

	if err := h.repo.Create(r.Context(), req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	kodeMK := r.PathValue("kode_mk")
	var req MataKuliah
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format JSON tidak valid"})
		return
	}

	req.Nama = strings.TrimSpace(req.Nama)
	req.Kurikulum = strings.TrimSpace(req.Kurikulum)
	if req.Nama == "" || req.Kurikulum == "" || req.SKS <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nama, kurikulum wajib diisi dan sks harus > 0"})
		return
	}

	err := h.repo.Update(r.Context(), kodeMK, req)
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	req.KodeMK = kodeMK
	writeJSON(w, http.StatusOK, req)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	kodeMK := r.PathValue("kode_mk")
	err := h.repo.Delete(r.Context(), kodeMK)
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "mata kuliah berhasil dihapus"})
}