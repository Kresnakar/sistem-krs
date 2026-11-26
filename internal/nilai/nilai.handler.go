package nilai

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/nilai", h.InputNilai)
	mux.HandleFunc("GET /api/nilai/transkrip/{nim}", h.GetTranskrip)
	mux.HandleFunc("GET /api/nilai/ips/{nim}/{semester_id}", h.GetIPS)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) InputNilai(w http.ResponseWriter, r *http.Request) {
	var req NilaiMahasiswa
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format JSON tidak valid"})
		return
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.KodeMK = strings.TrimSpace(req.KodeMK)
	req.NilaiHuruf = strings.ToUpper(strings.TrimSpace(req.NilaiHuruf))

	if req.NIM == "" || req.KodeMK == "" || req.SemesterID == 0 || req.NilaiHuruf == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nim, kode_mk, semester_id, dan nilai_huruf wajib diisi"})
		return
	}

	err := h.svc.InputkanNilai(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) GetTranskrip(w http.ResponseWriter, r *http.Request) {
	nim := strings.TrimSpace(r.PathValue("nim"))
	if nim == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "NIM tidak valid"})
		return
	}

	transkrip, err := h.svc.AmbilTranskripTerbaik(r.Context(), nim)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, transkrip)
}

func (h *Handler) GetIPS(w http.ResponseWriter, r *http.Request) {
	nim := strings.TrimSpace(r.PathValue("nim"))
	semIDStr := strings.TrimSpace(r.PathValue("semester_id"))
	semID, err := strconv.Atoi(semIDStr)

	if nim == "" || err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "NIM atau ID Semester tidak valid"})
		return
	}

	ringkasan, err := h.svc.HitungIPS(r.Context(), nim, semID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ringkasan)
}
