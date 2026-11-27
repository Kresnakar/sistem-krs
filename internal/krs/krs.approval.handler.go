package krs

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

// DosenHandler menangani persetujuan KRS dan perubahan KRS (Dosen Wali).
type DosenHandler struct {
	service *Service
	store   *sessions.CookieStore
}

func NewDosenHandler(service *Service, store *sessions.CookieStore) *DosenHandler {
	return &DosenHandler{service: service, store: store}
}

func (h *DosenHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/krs/dosen/krs", h.ListKRS)
	mux.HandleFunc("GET /api/krs/dosen/krs/{krs_id}", h.DetailKRS)
	mux.HandleFunc("POST /api/krs/dosen/krs/{krs_id}/setujui", h.SetujuiKRS)
	mux.HandleFunc("POST /api/krs/dosen/krs/{krs_id}/tolak", h.TolakKRS)

	mux.HandleFunc("GET /api/krs/dosen/perubahan", h.ListPerubahan)
	mux.HandleFunc("POST /api/krs/dosen/perubahan/{id}/setujui", h.SetujuiPerubahan)
	mux.HandleFunc("POST /api/krs/dosen/perubahan/{id}/tolak", h.TolakPerubahan)
}

func (h *DosenHandler) dosenID(r *http.Request) (int, bool) {
	return sessionDosenWaliID(r, h.store)
}

func (h *DosenHandler) ListKRS(w http.ResponseWriter, r *http.Request) {
	dosenID, ok := h.dosenID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session dosen wali tidak valid")
		return
	}

	data, err := h.service.DaftarKRSDosen(r.Context(), dosenID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *DosenHandler) DetailKRS(w http.ResponseWriter, r *http.Request) {
	dosenID, ok := h.dosenID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session dosen wali tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}

	data, err := h.service.AmbilKRSByIDLengkap(r.Context(), krsID, dosenID)
	if errors.Is(err, ErrKRSNotFound) {
		writeError(w, http.StatusNotFound, "KRS tidak ditemukan atau bukan bimbingan Anda")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *DosenHandler) SetujuiKRS(w http.ResponseWriter, r *http.Request) {
	dosenID, ok := h.dosenID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session dosen wali tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}

	if err := h.service.SetujuiKRS(r.Context(), krsID, dosenID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "KRS berhasil disetujui"})
}

func (h *DosenHandler) TolakKRS(w http.ResponseWriter, r *http.Request) {
	dosenID, ok := h.dosenID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session dosen wali tidak valid")
		return
	}

	krsID, err := parsePathID(r, "krs_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "KRS id tidak valid")
		return
	}

	var req struct {
		Catatan string `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	if err := h.service.TolakKRS(r.Context(), krsID, dosenID, req.Catatan); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "KRS ditolak"})
}

func (h *DosenHandler) ListPerubahan(w http.ResponseWriter, r *http.Request) {
	dosenID, ok := h.dosenID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session dosen wali tidak valid")
		return
	}

	data, err := h.service.DaftarPerubahanDosen(r.Context(), dosenID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *DosenHandler) SetujuiPerubahan(w http.ResponseWriter, r *http.Request) {
	dosenID, ok := h.dosenID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session dosen wali tidak valid")
		return
	}

	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id permintaan tidak valid")
		return
	}

	if err := h.service.SetujuiPerubahan(r.Context(), id, dosenID); err != nil {
		if errors.Is(err, ErrPermintaanNotFound) || errors.Is(err, ErrPermintaanDiproses) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "perubahan KRS disetujui"})
}

func (h *DosenHandler) TolakPerubahan(w http.ResponseWriter, r *http.Request) {
	dosenID, ok := h.dosenID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session dosen wali tidak valid")
		return
	}

	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id permintaan tidak valid")
		return
	}

	var req struct {
		Catatan string `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	if err := h.service.TolakPerubahan(r.Context(), id, dosenID, req.Catatan); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pesan": "perubahan KRS ditolak"})
}
