package krs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrJendelaKRSTutup       = errors.New("periode pengisian KRS sedang ditutup")
	ErrJendelaPerubahanTutup = errors.New("periode perubahan KRS sedang ditutup")
)

func (s *Service) DaftarSemester(ctx context.Context) ([]Semester, error) {
	return s.repo.ListSemester(ctx)
}

func (s *Service) TambahSemester(ctx context.Context, req SemesterRequest) (*Semester, error) {
	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)
	req.Jenis = strings.TrimSpace(req.Jenis)

	if req.TahunAkademik == "" || req.Jenis == "" {
		return nil, errors.New("tahun_akademik dan jenis wajib diisi")
	}
	if req.Jenis != "Ganjil" && req.Jenis != "Genap" && req.Jenis != "Antara" {
		return nil, errors.New("jenis semester harus Ganjil, Genap, atau Antara")
	}

	sem := Semester{
		TahunAkademik: req.TahunAkademik,
		Jenis:         req.Jenis,
		StatusAktif:   req.StatusAktif,
	}

	id, err := s.repo.CreateSemester(ctx, sem)
	if err != nil {
		return nil, err
	}
	if req.StatusAktif {
		if err := s.repo.AktifkanSemester(ctx, id); err != nil {
			return nil, err
		}
	}
	return s.repo.GetSemester(ctx, id)
}

func (s *Service) UbahSemester(ctx context.Context, id int, req SemesterRequest) (*Semester, error) {
	if _, err := s.repo.GetSemester(ctx, id); err != nil {
		return nil, err
	}

	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)
	req.Jenis = strings.TrimSpace(req.Jenis)
	if req.TahunAkademik == "" || req.Jenis == "" {
		return nil, errors.New("tahun_akademik dan jenis wajib diisi")
	}
	if req.Jenis != "Ganjil" && req.Jenis != "Genap" && req.Jenis != "Antara" {
		return nil, errors.New("jenis semester tidak valid")
	}

	sem := Semester{
		ID:            id,
		TahunAkademik: req.TahunAkademik,
		Jenis:         req.Jenis,
		StatusAktif:   req.StatusAktif,
	}
	if err := s.repo.UpdateSemester(ctx, id, sem); err != nil {
		return nil, err
	}
	if req.StatusAktif {
		if err := s.repo.AktifkanSemester(ctx, id); err != nil {
			return nil, err
		}
	}
	return s.repo.GetSemester(ctx, id)
}

// AmbilSemesterAktif dipakai halaman mahasiswa untuk mengetahui semester berjalan.
func (s *Service) AmbilSemesterAktif(ctx context.Context) (*Semester, *PeriodeKRS, error) {
	sem, err := s.repo.GetActiveSemester(ctx)
	if err != nil {
		return nil, nil, err
	}

	periode, err := s.repo.GetPeriodeBySemester(ctx, sem.ID)
	if err != nil && !errors.Is(err, ErrPeriodeNotFound) {
		return nil, nil, err
	}
	return sem, periode, nil
}

func parsePeriodeRequest(req PeriodeRequest) (PeriodeKRS, error) {
	var p PeriodeKRS
	p.SemesterID = req.SemesterID
	if p.SemesterID <= 0 {
		return p, errors.New("semester_id wajib diisi")
	}

	var err error
	p.MulaiKRS, err = parseDateTime(req.MulaiKRS)
	if err != nil {
		return p, fmt.Errorf("mulai_krs: %w", err)
	}
	p.SelesaiKRS, err = parseDateTime(req.SelesaiKRS)
	if err != nil {
		return p, fmt.Errorf("selesai_krs: %w", err)
	}
	p.MulaiPerubahan, err = parseDateTime(req.MulaiPerubahan)
	if err != nil {
		return p, fmt.Errorf("mulai_perubahan: %w", err)
	}
	p.SelesaiPerubahan, err = parseDateTime(req.SelesaiPerubahan)
	if err != nil {
		return p, fmt.Errorf("selesai_perubahan: %w", err)
	}

	if strings.TrimSpace(req.TanggalMulaiKuliah) == "" {
		p.TanggalMulaiKuliah = p.MulaiKRS
	} else {
		p.TanggalMulaiKuliah, err = parseDateTime(req.TanggalMulaiKuliah)
		if err != nil {
			return p, fmt.Errorf("tanggal_mulai_kuliah: %w", err)
		}
	}

	if !p.SelesaiKRS.After(p.MulaiKRS) {
		return p, errors.New("selesai_krs harus setelah mulai_krs")
	}
	if !p.SelesaiPerubahan.After(p.MulaiPerubahan) {
		return p, errors.New("selesai_perubahan harus setelah mulai_perubahan")
	}
	return p, nil
}

func (s *Service) DaftarPeriode(ctx context.Context) ([]PeriodeKRS, error) {
	return s.repo.ListPeriode(ctx)
}

func (s *Service) TambahPeriode(ctx context.Context, req PeriodeRequest) (*PeriodeKRS, error) {
	p, err := parsePeriodeRequest(req)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetSemester(ctx, p.SemesterID); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetPeriodeBySemester(ctx, p.SemesterID); err == nil {
		return nil, errors.New("periode untuk semester tersebut sudah ada")
	} else if !errors.Is(err, ErrPeriodeNotFound) {
		return nil, err
	}

	id, err := s.repo.CreatePeriode(ctx, p)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPeriode(ctx, id)
}

func (s *Service) UbahPeriode(ctx context.Context, id int, req PeriodeRequest) (*PeriodeKRS, error) {
	if _, err := s.repo.GetPeriode(ctx, id); err != nil {
		return nil, err
	}

	p, err := parsePeriodeRequest(req)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetSemester(ctx, p.SemesterID); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePeriode(ctx, id, p); err != nil {
		return nil, err
	}
	return s.repo.GetPeriode(ctx, id)
}

// BukaPeriode mengatur jendela waktu (tanpa mengubah skema database).
func (s *Service) BukaPeriode(ctx context.Context, id int, req BukaPeriodeRequest) (*PeriodeKRS, error) {
	periode, err := s.repo.GetPeriode(ctx, id)
	if err != nil {
		return nil, err
	}

	jenis := strings.ToLower(strings.TrimSpace(req.Jenis))
	if jenis != "krs" && jenis != "perubahan" {
		return nil, errors.New("jenis harus 'krs' atau 'perubahan'")
	}

	mulai, ok, err := parseOptionalDateTime(req.Mulai)
	if err != nil {
		return nil, err
	}
	if !ok {
		mulai = time.Now()
	}

	selesai, ok, err := parseOptionalDateTime(req.Selesai)
	if err != nil {
		return nil, err
	}
	if !ok {
		selesai = mulai.AddDate(0, 0, 7)
	}
	if !selesai.After(mulai) {
		return nil, errors.New("selesai harus setelah mulai")
	}

	if err := s.repo.UpdateJendelaPeriode(ctx, periode.ID, jenis, mulai, selesai); err != nil {
		return nil, err
	}
	return s.repo.GetPeriode(ctx, id)
}

// TutupPeriode langsung menutup jendela dengan menyetel selesai = sekarang.
func (s *Service) TutupPeriode(ctx context.Context, id int, req TutupPeriodeRequest) (*PeriodeKRS, error) {
	if _, err := s.repo.GetPeriode(ctx, id); err != nil {
		return nil, err
	}

	jenis := strings.ToLower(strings.TrimSpace(req.Jenis))
	if jenis != "krs" && jenis != "perubahan" {
		return nil, errors.New("jenis harus 'krs' atau 'perubahan'")
	}

	if err := s.repo.TutupJendelaPeriode(ctx, id, jenis); err != nil {
		return nil, err
	}
	return s.repo.GetPeriode(ctx, id)
}

func (s *Service) DaftarKelas(ctx context.Context, semesterID int) ([]KelasInfo, error) {
	return s.repo.ListKelasBySemester(ctx, semesterID)
}

func (s *Service) ValidasiJendelaKRS(ctx context.Context, semesterID int) error {
	open, err := s.repo.IsKRSWindowOpen(ctx, semesterID)
	if err != nil {
		return err
	}
	if !open {
		return ErrJendelaKRSTutup
	}
	return nil
}

func (s *Service) ValidasiJendelaPerubahan(ctx context.Context, semesterID int) error {
	open, err := s.repo.IsPerubahanWindowOpen(ctx, semesterID)
	if err != nil {
		return err
	}
	if !open {
		return ErrJendelaPerubahanTutup
	}
	return nil
}
