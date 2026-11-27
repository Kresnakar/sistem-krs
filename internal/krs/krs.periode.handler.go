package krs

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// PeriodeHandler menangani pengelolaan semester dan periode KRS (Admin Prodi).
type PeriodeHandler struct {
	service *Service
}

func NewPeriodeHandler(service *Service) *PeriodeHandler {
	return &PeriodeHandler{service: service}
}

func (h *PeriodeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/krs/admin/semester", h.ListSemester)
	mux.HandleFunc("POST /api/krs/admin/semester", h.CreateSemester)
	mux.HandleFunc("PUT /api/krs/admin/semester/{id}", h.UpdateSemester)

	mux.HandleFunc("GET /api/krs/admin/periode", h.ListPeriode)
	mux.HandleFunc("POST /api/krs/admin/periode", h.CreatePeriode)
	mux.HandleFunc("PUT /api/krs/admin/periode/{id}", h.UpdatePeriode)
	mux.HandleFunc("POST /api/krs/admin/periode/{id}/buka", h.BukaPeriode)
	mux.HandleFunc("POST /api/krs/admin/periode/{id}/tutup", h.TutupPeriode)
}

func parsePathID(r *http.Request, name string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(r.PathValue(name)))
}

func (h *PeriodeHandler) ListSemester(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.DaftarSemester(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *PeriodeHandler) CreateSemester(w http.ResponseWriter, r *http.Request) {
	var req SemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	data, err := h.service.TambahSemester(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, data)
}

func (h *PeriodeHandler) UpdateSemester(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id semester tidak valid")
		return
	}

	var req SemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	data, err := h.service.UbahSemester(r.Context(), id, req)
	if errors.Is(err, ErrSemesterNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *PeriodeHandler) ListPeriode(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.DaftarPeriode(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *PeriodeHandler) CreatePeriode(w http.ResponseWriter, r *http.Request) {
	var req PeriodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	data, err := h.service.TambahPeriode(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, data)
}

func (h *PeriodeHandler) UpdatePeriode(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id periode tidak valid")
		return
	}

	var req PeriodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	data, err := h.service.UbahPeriode(r.Context(), id, req)
	if errors.Is(err, ErrPeriodeNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *PeriodeHandler) BukaPeriode(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id periode tidak valid")
		return
	}

	var req BukaPeriodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	data, err := h.service.BukaPeriode(r.Context(), id, req)
	if errors.Is(err, ErrPeriodeNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *PeriodeHandler) TutupPeriode(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id periode tidak valid")
		return
	}

	var req TutupPeriodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	data, err := h.service.TutupPeriode(r.Context(), id, req)
	if errors.Is(err, ErrPeriodeNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}
