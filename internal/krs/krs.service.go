package krs

import (
	"context"
	"database/sql"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

var (
	ErrMahasiswaTidakDitemukan = errors.New("mahasiswa tidak ditemukan")
	ErrKRSBelumAda             = errors.New("data KRS belum tersedia")
	ErrKRSBukanDraft           = errors.New("KRS tidak dalam status Draft")
	ErrBatasSKS                = errors.New("pengambilan SKS melebihi batas")
)

func (s *Service) AmbilKRS(
	ctx context.Context,
	nim string,
	semesterID int,
) (*KRS, error) {

	data, err := s.repo.GetByNIMSemester(
		ctx,
		nim,
		semesterID,
	)

	if err != nil {
		return nil, err
	}

	detail, err := s.repo.GetDetail(
		ctx,
		data.ID,
	)

	if err != nil {
		return nil, err
	}

	data.Detail = detail

	return data, nil
}

func (s *Service) PerbaruiTotalSKS(
	ctx context.Context,
	krsID int,
) error {

	totalSKS, err := s.repo.HitungTotalSKS(
		ctx,
		krsID,
	)

	if err != nil {
		return err
	}

	return s.repo.UpdateTotalSKS(
		ctx,
		krsID,
		totalSKS,
	)
}

func (s *Service) TambahKelas(
	ctx context.Context,
	krsID int,
	kelasID int,
	batasSKS int,
) error {
	err := s.ValidasiStatusKRS(ctx, krsID)
	if err != nil {
		return err
	}

	err = s.ValidasiStatusKelas(ctx, kelasID)
	if err != nil {
		return err
	}

	err = s.ValidasiBatasSKS(
		ctx,
		krsID,
		kelasID,
		batasSKS,
	)

	if err != nil {
		return err
	}

	terdaftar, err := s.repo.CekKelasTerdaftar(
		ctx,
		krsID,
		kelasID,
	)
	if err != nil {
		return err
	}

	if terdaftar {
		return errors.New("kelas sudah terdaftar di KRS")
	}

	err = s.ValidasiKuotaKelas(ctx, kelasID)
	if err != nil {
		return err
	}

	err = s.repo.TambahTerisi(ctx, kelasID)
	if err != nil {
		return err
	}

	err = s.repo.TambahDetail(ctx, krsID, kelasID)
	if err != nil {
		_ = s.repo.KurangiTerisi(ctx, kelasID)
		return err
	}

	return s.PerbaruiTotalSKS(ctx, krsID)
}

func (s *Service) BatalkanKelas(
	ctx context.Context,
	krsID int,
	kelasID int,
) error {

	err := s.ValidasiStatusKRS(ctx, krsID)
	if err != nil {
		return err
	}

	err = s.repo.HapusDetail(ctx, krsID, kelasID)
	if err != nil {
		return err
	}

	err = s.repo.KurangiTerisi(ctx, kelasID)
	if err != nil {
		return err
	}

	return s.PerbaruiTotalSKS(ctx, krsID)
}

func (s *Service) ValidasiBatasSKS(
	ctx context.Context,
	krsID int,
	kelasID int,
	batasSKS int,
) error {

	totalSKS, err := s.repo.HitungTotalSKS(
		ctx,
		krsID,
	)

	if err != nil {
		return err
	}

	var sksKelas int

	query := `
        SELECT mk.sks
        FROM kelas k
        JOIN mata_kuliah mk ON k.kode_mk = mk.kode_mk
        WHERE k.id = ?
    `

	err = s.repo.db.QueryRowContext(
		ctx,
		query,
		kelasID,
	).Scan(&sksKelas)

	if err != nil {
		return err
	}

	if totalSKS+sksKelas > batasSKS {
		return ErrBatasSKS
	}

	return nil
}

func (s *Service) ValidasiStatusKelas(
	ctx context.Context,
	kelasID int,
) error {

	var status string

	query := `
        SELECT status
        FROM kelas
        WHERE id = ?
    `

	err := s.repo.db.QueryRowContext(
		ctx,
		query,
		kelasID,
	).Scan(&status)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("kelas tidak ditemukan")
	}

	if err != nil {
		return err
	}

	if status != "Dibuka" {
		return errors.New("kelas tidak dibuka")
	}

	return nil
}

func (s *Service) ValidasiKuotaKelas(
	ctx context.Context,
	kelasID int,
) error {

	var kuota, terisi int

	query := `
        SELECT kuota, terisi
        FROM kelas
        WHERE id = ?
    `

	err := s.repo.db.QueryRowContext(
		ctx,
		query,
		kelasID,
	).Scan(&kuota, &terisi)

	if err != nil {
		return err
	}

	if terisi >= kuota {
		return errors.New("kuota kelas sudah penuh")
	}

	return nil
}

func (s *Service) ValidasiStatusKRS(
	ctx context.Context,
	krsID int,
) error {
	var status string

	query := `
        SELECT status
        FROM krs
        WHERE id = ?
    `

	err := s.repo.db.QueryRowContext(
		ctx,
		query,
		krsID,
	).Scan(&status)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("KRS tidak ditemukan")
	}

	if err != nil {
		return err
	}

	if status != "Draft" {
		return ErrKRSBukanDraft
	}

	return nil
}
