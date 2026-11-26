package nilai

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) InputkanNilai(ctx context.Context, n *NilaiMahasiswa) error {
	// Validasi bisnis tambahan bisa ditambahkan di sini (misal: cek apakah semester sudah ditutup)
	return s.repo.UpsertNilai(ctx, n)
}

func (s *Service) AmbilTranskripTerbaik(ctx context.Context, nim string) ([]TranskripEntry, error) {
	return s.repo.GetTranskripTerbaik(ctx, nim)
}

// HitungIPS (Inti Tugas Anda - FR-04)
// Menghitung IPS berdasarkan seluruh MK yang diambil di semester tersebut.
func (s *Service) HitungIPS(ctx context.Context, nim string, semesterID int) (*RingkasanSemester, error) {
	riwayat, err := s.repo.GetRiwayatSemester(ctx, nim, semesterID)
	if err != nil {
		return nil, err
	}

	if len(riwayat) == 0 {
		// Kembalikan ringkasan kosong jika tidak ada data
		tahun, jenis, _ := s.repo.GetInfoSemester(ctx, semesterID)
		return &RingkasanSemester{
			SemesterID: semesterID, TahunAkademik: tahun, JenisSemester: jenis,
			TotalSKS: 0, TotalKualitas: 0, IPS: 0,
		}, nil
	}

	totalSKS := 0
	totalKualitas := 0.0

	for _, n := range riwayat {
		totalSKS += n.SKS
		totalKualitas += float64(n.SKS) * n.Bobot
	}

	ips := 0.0
	if totalSKS > 0 {
		ips = totalKualitas / float64(totalSKS)
	}

	tahun, jenis, err := s.repo.GetInfoSemester(ctx, semesterID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil info semester: %w", err)
	}

	return &RingkasanSemester{
		SemesterID:    semesterID,
		TahunAkademik: tahun,
		JenisSemester: jenis,
		TotalSKS:      totalSKS,
		TotalKualitas: totalKualitas,
		IPS:           ips,
	}, nil
}
