package krs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (s *Service) BuatDraft(ctx context.Context, nim string, semesterID int) (*KRS, error) {
	if semesterID <= 0 {
		return nil, errors.New("semester_id wajib diisi")
	}
	if _, err := s.repo.GetSemester(ctx, semesterID); err != nil {
		return nil, err
	}
	return s.repo.GetOrCreateDraft(ctx, nim, semesterID)
}

func (s *Service) AjukanKRS(ctx context.Context, krsID int, nim string) error {
	ok, err := s.ValidasiKepemilikanKRS(ctx, krsID, nim)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("KRS bukan milik Anda")
	}

	data, err := s.repo.GetKRSByID(ctx, krsID)
	if err != nil {
		return err
	}
	if data.Status != "Draft" && data.Status != "Ditolak" {
		return ErrKRSBukanDraft
	}
	if data.TotalSKS <= 0 {
		return errors.New("KRS belum memiliki kelas yang diambil")
	}
	if err := s.ValidasiJendelaKRS(ctx, data.SemesterID); err != nil {
		return err
	}

	return s.repo.AjukanKRS(ctx, krsID)
}

func (s *Service) AmbilKRSLengkap(ctx context.Context, nim string, semesterID int) (*KRSLengkap, error) {
	data, err := s.AmbilKRS(ctx, nim, semesterID)
	if err != nil {
		return nil, err
	}

	detail, err := s.repo.GetDetailLengkap(ctx, data.ID)
	if err != nil {
		return nil, err
	}

	return &KRSLengkap{KRS: *data, DetailLengkap: detail}, nil
}

func (s *Service) AmbilKRSByIDLengkap(ctx context.Context, krsID, dosenWaliID int) (*KRSLengkap, error) {
	data, err := s.repo.GetKRSForDosen(ctx, krsID, dosenWaliID)
	if err != nil {
		return nil, err
	}

	detail, err := s.repo.GetDetailLengkap(ctx, data.ID)
	if err != nil {
		return nil, err
	}

	return &KRSLengkap{KRS: *data, DetailLengkap: detail}, nil
}

// BatasSKSKRS mengambil batas SKS dari data KRS (bukan dari input client).
func (s *Service) BatasSKSKRS(ctx context.Context, krsID int) (int, error) {
	data, err := s.repo.GetKRSByID(ctx, krsID)
	if err != nil {
		return 0, err
	}
	if data.BatasSKS <= 0 {
		return 21, nil
	}
	return data.BatasSKS, nil
}

func (s *Service) DaftarKRSDosen(ctx context.Context, dosenWaliID int) ([]KRSRingkas, error) {
	return s.repo.ListKRSByDosen(ctx, dosenWaliID)
}

func (s *Service) SetujuiKRS(ctx context.Context, krsID, dosenWaliID int) error {
	return s.repo.SetujuiKRS(ctx, krsID, dosenWaliID)
}

func (s *Service) TolakKRS(ctx context.Context, krsID, dosenWaliID int, catatan string) error {
	catatan = strings.TrimSpace(catatan)
	if catatan == "" {
		return errors.New("catatan penolakan wajib diisi")
	}
	return s.repo.TolakKRS(ctx, krsID, dosenWaliID, catatan)
}

func (s *Service) AjukanPerubahan(
	ctx context.Context,
	krsID int,
	nim string,
	jenis string,
	alasan string,
	kelasIDs []int,
) (int, error) {
	ok, err := s.ValidasiKepemilikanKRS(ctx, krsID, nim)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errors.New("KRS bukan milik Anda")
	}

	data, err := s.repo.GetKRSByID(ctx, krsID)
	if err != nil {
		return 0, err
	}
	if data.Status != "Disetujui" {
		return 0, errors.New("perubahan hanya dapat diajukan untuk KRS yang sudah disetujui")
	}
	if err := s.ValidasiJendelaPerubahan(ctx, data.SemesterID); err != nil {
		return 0, err
	}

	jenis = strings.TrimSpace(jenis)
	if jenis != "Tambah Kelas" && jenis != "Batal Kelas" {
		return 0, errors.New("jenis_perubahan harus 'Tambah Kelas' atau 'Batal Kelas'")
	}
	if len(kelasIDs) == 0 {
		return 0, errors.New("minimal satu kelas harus dipilih")
	}

	for _, kelasID := range kelasIDs {
		var semesterID int
		var status string
		err := s.repo.db.QueryRowContext(ctx, `
			SELECT semester_id, status FROM kelas WHERE id = ?`, kelasID,
		).Scan(&semesterID, &status)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("kelas %d tidak ditemukan", kelasID)
		}
		if err != nil {
			return 0, err
		}
		if semesterID != data.SemesterID {
			return 0, fmt.Errorf("kelas %d bukan dari semester yang sama", kelasID)
		}
		if status == "Dibatalkan" {
			return 0, fmt.Errorf("kelas %d sudah dibatalkan", kelasID)
		}
		if jenis == "Tambah Kelas" && status != "Dibuka" {
			return 0, fmt.Errorf("kelas %d sedang tidak dibuka", kelasID)
		}

		var terdaftar bool
		err = s.repo.db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM krs_detail
				WHERE krs_id = ? AND kelas_id = ? AND status = 'Terdaftar'
			)`, krsID, kelasID).Scan(&terdaftar)
		if err != nil {
			return 0, err
		}

		if jenis == "Tambah Kelas" && terdaftar {
			return 0, fmt.Errorf("kelas %d sudah terdaftar di KRS", kelasID)
		}
		if jenis == "Batal Kelas" && !terdaftar {
			return 0, fmt.Errorf("kelas %d belum terdaftar di KRS", kelasID)
		}
	}

	return s.repo.CreatePermintaanPerubahan(ctx, krsID, jenis, strings.TrimSpace(alasan), kelasIDs)
}

func (s *Service) DaftarPerubahanMahasiswa(ctx context.Context, nim string) ([]PermintaanPerubahan, error) {
	return s.repo.ListPermintaanByNIM(ctx, nim)
}

func (s *Service) DaftarPerubahanKRS(ctx context.Context, krsID int) ([]PermintaanPerubahan, error) {
	return s.repo.ListPermintaanByKRS(ctx, krsID)
}

func (s *Service) DaftarPerubahanDosen(ctx context.Context, dosenWaliID int) ([]PermintaanPerubahan, error) {
	return s.repo.ListPermintaanByDosen(ctx, dosenWaliID)
}

func (s *Service) SetujuiPerubahan(ctx context.Context, permintaanID, dosenWaliID int) error {
	return s.repo.SetujuiPermintaan(ctx, permintaanID, dosenWaliID)
}

func (s *Service) TolakPerubahan(ctx context.Context, permintaanID, dosenWaliID int, catatan string) error {
	catatan = strings.TrimSpace(catatan)
	if catatan == "" {
		return errors.New("catatan penolakan wajib diisi")
	}
	return s.repo.TolakPermintaan(ctx, permintaanID, dosenWaliID, catatan)
}
