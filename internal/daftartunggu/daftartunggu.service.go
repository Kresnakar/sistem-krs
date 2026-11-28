package daftartunggu

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNIMTidakValid           = errors.New("NIM tidak valid")
	ErrKelasIDTidakValid       = errors.New("ID kelas tidak valid")
	ErrRepositoryTidakTersedia = errors.New("repository belum tersedia")
	ErrValidatorTidakTersedia  = errors.New("validator KRS belum tersedia")
	ErrTidakMemenuhiSyarat     = errors.New("mahasiswa tidak memenuhi syarat masuk daftar tunggu")
)

// KRSValidator merupakan kontrak untuk menggunakan
// validasi akademik yang dimiliki modul KRS.
type KRSValidator interface {
	EvaluasiUntukDaftarTunggu(
		ctx context.Context,
		nim string,
		kelasID int,
	) (bool, string, error)
}

type Service struct {
	repo      *Repository
	validator KRSValidator
}

// NewService membuat service daftar tunggu.
func NewService(
	repo *Repository,
	validator KRSValidator,
) *Service {
	return &Service{
		repo:      repo,
		validator: validator,
	}
}

// DaftarSaya mengambil seluruh antrean milik mahasiswa.
func (s *Service) DaftarSaya(
	ctx context.Context,
	nim string,
) ([]DaftarTunggu, error) {

	if s.repo == nil {
		return nil, ErrRepositoryTidakTersedia
	}

	nim = strings.TrimSpace(nim)
	if nim == "" {
		return nil, ErrNIMTidakValid
	}

	return s.repo.ListByNIM(ctx, nim)
}

// DaftarKelas mengambil seluruh antrean pada kelas tertentu.
// Fungsi ini nantinya digunakan oleh Admin Prodi.
func (s *Service) DaftarKelas(
	ctx context.Context,
	kelasID int,
) ([]DaftarTunggu, error) {

	if s.repo == nil {
		return nil, ErrRepositoryTidakTersedia
	}

	if kelasID <= 0 {
		return nil, ErrKelasIDTidakValid
	}

	return s.repo.ListByKelas(ctx, kelasID)
}

// MasukAntrean mendaftarkan mahasiswa ke daftar tunggu.
func (s *Service) MasukAntrean(
	ctx context.Context,
	nim string,
	kelasID int,
) (*DaftarTunggu, error) {

	if s.repo == nil {
		return nil, ErrRepositoryTidakTersedia
	}

	if s.validator == nil {
		return nil, ErrValidatorTidakTersedia
	}

	nim = strings.TrimSpace(nim)
	if nim == "" {
		return nil, ErrNIMTidakValid
	}

	if kelasID <= 0 {
		return nil, ErrKelasIDTidakValid
	}

	// Periksa kelayakan akademik mahasiswa.
	// Validasi menggunakan service dari modul KRS,
	// bukan membuat ulang aturan validasi di sini.
	layak, alasan, err := s.validator.EvaluasiUntukDaftarTunggu(
		ctx,
		nim,
		kelasID,
	)

	if err != nil {
		return nil, fmt.Errorf("gagal memvalidasi mahasiswa: %w", err)
	}

	if !layak {
		return nil, fmt.Errorf(
			"%w: %s",
			ErrTidakMemenuhiSyarat,
			alasan,
		)
	}

	// Repository memeriksa kuota, mengunci baris kelas,
	// menentukan nomor antrean, dan menyimpan data.
	data, err := s.repo.MasukAntrean(
		ctx,
		nim,
		kelasID,
	)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// KeluarAntrean mengeluarkan mahasiswa dari antreannya sendiri.
func (s *Service) KeluarAntrean(
	ctx context.Context,
	nim string,
	antreanID int,
) error {

	if s.repo == nil {
		return ErrRepositoryTidakTersedia
	}

	nim = strings.TrimSpace(nim)
	if nim == "" {
		return ErrNIMTidakValid
	}

	if antreanID <= 0 {
		return errors.New("ID antrean tidak valid")
	}

	return s.repo.KeluarAntrean(
		ctx,
		nim,
		antreanID,
	)
}

const batasKonfirmasiDefault = 24 * time.Hour

// PromosikanAntrean memproses kandidat antrean berdasarkan FIFO
// ketika tersedia kursi pada suatu kelas.
func (s *Service) PromosikanAntrean(
	ctx context.Context,
	kelasID int,
) (*HasilPromosi, error) {

	if s.repo == nil {
		return nil, ErrRepositoryTidakTersedia
	}

	if s.validator == nil {
		return nil, ErrValidatorTidakTersedia
	}

	if kelasID <= 0 {
		return nil, ErrKelasIDTidakValid
	}

	// 1. Mulai transaksi.
	tx, err := s.repo.MulaiTransaksiPromosi(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 2. Kunci baris kelas agar proses promosi
	// pada kelas yang sama tidak berjalan bersamaan.
	kapasitas, err := tx.KunciKelas(ctx, kelasID)
	if err != nil {
		return nil, err
	}

	// 3. Hitung kursi yang sedang ditahan.
	ditahan, err := tx.HitungKursiDitahan(ctx, kelasID)
	if err != nil {
		return nil, err
	}

	kapasitas.Ditahan = ditahan

	hasil := &HasilPromosi{
		KelasID:      kelasID,
		Dipromosikan: []DaftarTunggu{},
		Dilewati:     []DaftarTunggu{},
	}

	sisaKursi := kapasitas.KursiTersedia()

	// Tidak ada kursi tersedia.
	if sisaKursi <= 0 {
		hasil.KursiTersedia = 0
		return hasil, nil
	}

	// 4. Ambil kandidat berdasarkan FIFO.
	kandidat, err := tx.DaftarKandidat(ctx, kelasID)
	if err != nil {
		return nil, err
	}

	// 5. Periksa kandidat satu per satu.
	for _, mahasiswa := range kandidat {

		if sisaKursi <= 0 {
			break
		}

		// Validasi ulang kondisi akademik mahasiswa.
		layak, alasan, err := s.validator.EvaluasiUntukDaftarTunggu(
			ctx,
			mahasiswa.NIM,
			kelasID,
		)

		// Error teknis membatalkan seluruh transaksi.
		if err != nil {
			return nil, fmt.Errorf(
				"gagal memvalidasi mahasiswa %s: %w",
				mahasiswa.NIM,
				err,
			)
		}

		// 6. Jika tidak memenuhi syarat, tandai Dilewati.
		if !layak {

			if strings.TrimSpace(alasan) == "" {
				alasan = "Tidak memenuhi persyaratan akademik"
			}

			err := tx.TandaiDilewati(
				ctx,
				mahasiswa.ID,
				alasan,
			)
			if err != nil {
				return nil, err
			}

			mahasiswa.Status = StatusDilewati
			mahasiswa.AlasanDilewati = &alasan

			hasil.Dilewati = append(
				hasil.Dilewati,
				mahasiswa,
			)

			// Lanjutkan ke kandidat berikutnya.
			continue
		}

		// 7. Jika memenuhi syarat, tahan kursi selama 24 jam.
		batasKonfirmasi := time.Now().Add(batasKonfirmasiDefault)

		err = tx.TahanKursi(
			ctx,
			mahasiswa.ID,
			batasKonfirmasi,
		)
		if err != nil {
			return nil, err
		}

		mahasiswa.Status = StatusMenungguKonfirmasi
		mahasiswa.AlasanDilewati = nil
		mahasiswa.BatasKonfirmasi = &batasKonfirmasi

		hasil.Dipromosikan = append(
			hasil.Dipromosikan,
			mahasiswa,
		)

		// Satu kursi telah ditahan.
		sisaKursi--
	}

	hasil.KursiTersedia = sisaKursi

	// 8. Simpan seluruh perubahan secara atomik.
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return hasil, nil
}

// KonfirmasiTawaran memproses keputusan mahasiswa terhadap tawaran kursi.
func (s *Service) KonfirmasiTawaran(
	ctx context.Context,
	nim string,
	antreanID int,
	aksi string,
) error {

	if s.repo == nil {
		return ErrRepositoryTidakTersedia
	}

	nim = strings.TrimSpace(nim)
	if nim == "" {
		return ErrNIMTidakValid
	}

	if antreanID <= 0 {
		return errors.New("ID antrean tidak valid")
	}

	switch strings.ToLower(strings.TrimSpace(aksi)) {

	case "terima":
		if s.validator == nil {
			return ErrValidatorTidakTersedia
		}

		return s.terimaTawaran(ctx, nim, antreanID)

	case "tolak":
		kelasID, err := s.repo.TolakTawaran(
			ctx,
			nim,
			antreanID,
		)

		// Jika tawaran ternyata sudah kedaluwarsa,
		// kursi harus segera ditawarkan kepada kandidat berikutnya.
		if errors.Is(err, ErrKonfirmasiKedaluwarsa) {
			if promoErr := s.promosikanSetelahKursiLepas(
				ctx,
				kelasID,
			); promoErr != nil {
				return fmt.Errorf(
					"tawaran kedaluwarsa dan promosi berikutnya gagal: %w",
					promoErr,
				)
			}

			return ErrKonfirmasiKedaluwarsa
		}

		if err != nil {
			return err
		}

		// Kursi dilepaskan karena mahasiswa menolak tawaran.
		return s.promosikanSetelahKursiLepas(
			ctx,
			kelasID,
		)

	default:
		return ErrAksiKonfirmasiTidakValid
	}
}

func (s *Service) terimaTawaran(
	ctx context.Context,
	nim string,
	antreanID int,
) error {

	// 1. Ambil kelas dari antrean milik mahasiswa.
	kelasID, err := s.repo.GetKelasIDByAntrean(
		ctx,
		nim,
		antreanID,
	)
	if err != nil {
		return err
	}

	// 2. Validasi ulang kelayakan akademik.
	layak, alasan, err := s.validator.EvaluasiUntukDaftarTunggu(
		ctx,
		nim,
		kelasID,
	)
	if err != nil {
		return fmt.Errorf(
			"gagal memvalidasi kelayakan mahasiswa: %w",
			err,
		)
	}

	// 3. Jika tidak lagi memenuhi syarat, lepaskan kursi.
	if !layak {
		if strings.TrimSpace(alasan) == "" {
			alasan = "Tidak lagi memenuhi persyaratan akademik"
		}

		kelasID, err = s.repo.LewatiTawaran(
			ctx,
			nim,
			antreanID,
			alasan,
		)

		if errors.Is(err, ErrKonfirmasiKedaluwarsa) {
			if promoErr := s.promosikanSetelahKursiLepas(
				ctx,
				kelasID,
			); promoErr != nil {
				return fmt.Errorf(
					"tawaran kedaluwarsa dan promosi berikutnya gagal: %w",
					promoErr,
				)
			}

			return ErrKonfirmasiKedaluwarsa
		}

		if err != nil {
			return err
		}

		if err := s.promosikanSetelahKursiLepas(
			ctx,
			kelasID,
		); err != nil {
			return fmt.Errorf(
				"gagal mempromosikan antrean berikutnya: %w",
				err,
			)
		}

		return fmt.Errorf(
			"%w: %s",
			ErrTidakMemenuhiSyarat,
			alasan,
		)
	}

	// 4. Coba penerimaan langsung untuk KRS Draft/Ditolak.
	kelasID, err = s.repo.TerimaTawaran(
		ctx,
		nim,
		antreanID,
	)

	// 5. Jika KRS bukan Draft/Ditolak, coba jalur perubahan KRS.
	// Repository memvalidasi ulang status Disetujui dan
	// periode perubahan di dalam transaksi.
	if errors.Is(err, ErrKonfirmasiHanyaPeriodePengisian) {
		kelasID, err = s.repo.TerimaTawaranPerubahan(
			ctx,
			nim,
			antreanID,
		)
	}

	// 6. Jika tawaran kedaluwarsa, lepaskan kursi.
	if errors.Is(err, ErrKonfirmasiKedaluwarsa) {
		if promoErr := s.promosikanSetelahKursiLepas(
			ctx,
			kelasID,
		); promoErr != nil {
			return fmt.Errorf(
				"tawaran kedaluwarsa dan promosi berikutnya gagal: %w",
				promoErr,
			)
		}

		return ErrKonfirmasiKedaluwarsa
	}

	return err
}

func (s *Service) promosikanSetelahKursiLepas(
	ctx context.Context,
	kelasID int,
) error {

	if kelasID <= 0 {
		return ErrKelasIDTidakValid
	}

	_, err := s.PromosikanAntrean(ctx, kelasID)
	if err != nil {
		return fmt.Errorf(
			"gagal mempromosikan antrean kelas %d: %w",
			kelasID,
			err,
		)
	}

	return nil
}
