package kelas

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) TambahKelasBaru(ctx context.Context, kelas *KelasParalel) error {
	kelas.KodeMK = strings.TrimSpace(kelas.KodeMK)
	kelas.KodeParalel = strings.TrimSpace(kelas.KodeParalel)
	kelas.DosenPengampu = strings.TrimSpace(kelas.DosenPengampu)
	if kelas.Status == "" {
		kelas.Status = "Dibuka"
	}

	// Validasi Jadwal
	if len(kelas.Jadwal) == 0 {
		return errors.New("kelas harus memiliki setidaknya satu jadwal")
	}
	for i, j := range kelas.Jadwal {
		kelas.Jadwal[i].Hari = strings.TrimSpace(j.Hari)
		kelas.Jadwal[i].Ruang = strings.TrimSpace(j.Ruang)
		// Validasi format HH:MM
		if _, err := time.Parse("15:04", j.JamMulai); err != nil {
			return fmt.Errorf("format jam_mulai sesi %d tidak valid (HH:MM)", i+1)
		}
		if _, err := time.Parse("15:04", j.JamSelesai); err != nil {
			return fmt.Errorf("format jam_selesai sesi %d tidak valid (HH:MM)", i+1)
		}
		if j.JamMulai >= j.JamSelesai {
			return fmt.Errorf("jam mulai sesi %d harus sebelum jam selesai", i+1)
		}
	}

	return s.repo.Create(ctx, kelas)
}

func (s *Service) AmbilKelasBySemester(ctx context.Context, semesterID int) ([]*KelasParalel, error) {
	return s.repo.GetBySemester(ctx, semesterID)
}

func (s *Service) PerbaruiDataKelas(ctx context.Context, kelasID int, updateData *KelasParalel) error {
	// Validasi Jadwal (sama seperti Create)
	for i, j := range updateData.Jadwal {
		if j.JamMulai >= j.JamSelesai {
			return fmt.Errorf("jam mulai sesi %d harus sebelum jam selesai", i+1)
		}
	}
	return s.repo.Update(ctx, kelasID, updateData)
}

func (s *Service) HapusKelas(ctx context.Context, kelasID int) error {
	return s.repo.Delete(ctx, kelasID)
}
