package prasyarat

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	svc  *Service
	repo *Repository
}

func NewHandler(svc *Service, repo *Repository) *Handler {
	return &Handler{svc: svc, repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/prasyarat/{kode_mk}", h.GetPohon)
	mux.HandleFunc("POST /api/prasyarat", h.TambahNode)
	mux.HandleFunc("DELETE /api/prasyarat/{kode_mk}", h.HapusByKodeMK)
	mux.HandleFunc("GET /api/prasyarat/evaluasi/{nim}/{kode_mk}", h.Evaluasi)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) GetPohon(w http.ResponseWriter, r *http.Request) {
	kodeMK := strings.TrimSpace(r.PathValue("kode_mk"))
	pohon, err := h.svc.BangunPohon(r.Context(), kodeMK)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kode_mk": kodeMK,
		"pohon":   pohon,
	})
}

func (h *Handler) TambahNode(w http.ResponseWriter, r *http.Request) {
	var req PrasyaratNode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format JSON tidak valid"})
		return
	}

	hasil, err := h.svc.TambahNode(r.Context(), req)
	if errors.Is(err, ErrSiklusTerdeteksi) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, hasil)
}

func (h *Handler) HapusByKodeMK(w http.ResponseWriter, r *http.Request) {
	kodeMK := strings.TrimSpace(r.PathValue("kode_mk"))
	if err := h.repo.DeleteByKodeMK(r.Context(), kodeMK); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "aturan prasyarat berhasil dihapus"})
}

func (h *Handler) Evaluasi(w http.ResponseWriter, r *http.Request) {
	nim := strings.TrimSpace(r.PathValue("nim"))
	kodeMK := strings.TrimSpace(r.PathValue("kode_mk"))

	hasil, err := h.svc.Evaluasi(r.Context(), nim, kodeMK)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, hasil)
}