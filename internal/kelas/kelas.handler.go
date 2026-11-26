package kelas

import (
	"encoding/json"
	"errors"
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
	mux.HandleFunc("GET /api/kelas/semester/{semester_id}", h.GetBySemester)
	mux.HandleFunc("POST /api/kelas", h.Create)
	mux.HandleFunc("PUT /api/kelas/{kelas_id}", h.Update)
	mux.HandleFunc("DELETE /api/kelas/{kelas_id}", h.Delete)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) GetBySemester(w http.ResponseWriter, r *http.Request) {
	semIDStr := r.PathValue("semester_id")
	semID, err := strconv.Atoi(strings.TrimSpace(semIDStr))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id semester tidak valid"})
		return
	}

	data, err := h.svc.AmbilKelasBySemester(r.Context(), semID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req KelasParalel
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format JSON tidak valid"})
		return
	}

	// Validasi dasar (bisa dipindah ke service untuk best practice)
	if req.KodeMK == "" || req.SemesterID == 0 || req.KodeParalel == "" || req.Kuota <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kode_mk, semester_id, kode_paralel, dan kuota (>0) wajib diisi"})
		return
	}

	err := h.svc.TambahKelasBaru(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("kelas_id")
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id kelas tidak valid"})
		return
	}

	var req KelasParalel
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "format JSON tidak valid"})
		return
	}

	if req.Kuota <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kuota harus > 0"})
		return
	}

	err = h.svc.PerbaruiDataKelas(r.Context(), id, &req)
	if errors.Is(err, ErrKelasNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	req.ID = id
	writeJSON(w, http.StatusOK, req)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("kelas_id")
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id kelas tidak valid"})
		return
	}

	err = h.svc.HapusKelas(r.Context(), id)
	if errors.Is(err, ErrKelasNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		// Error dari FK constraint (masih ada mahasiswa)
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "kelas berhasil dihapus"})
}
