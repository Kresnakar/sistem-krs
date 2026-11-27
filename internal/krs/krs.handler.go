package krs

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Mendaftarkan endpoint API KRS
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/krs/{nim}/{semester_id}", h.AmbilKRS)
	mux.HandleFunc("POST /api/krs/{krs_id}/kelas", h.TambahKelas)
	mux.HandleFunc("DELETE /api/krs/{krs_id}/kelas/{kelas_id}", h.BatalkanKelas)
}

// Mengambil data KRS mahasiswa
func (h *Handler) AmbilKRS(
	w http.ResponseWriter,
	r *http.Request,
) {
	nim := r.PathValue("nim")

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

	if req.KelasID <= 0 || req.BatasSKS <= 0 {
		http.Error(
			w,
			"Kelas ID dan batas SKS harus valid",
			http.StatusBadRequest,
		)
		return
	}

	err = h.service.TambahKelas(
		r.Context(),
		krsID,
		req.KelasID,
		req.BatasSKS,
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
