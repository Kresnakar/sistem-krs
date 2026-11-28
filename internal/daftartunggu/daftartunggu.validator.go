package daftartunggu

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/kresnakar/sistem-krs/internal/krs"
)

// KesetaraanLookup merupakan kontrak untuk mengambil
// seluruh kode mata kuliah yang saling ekuivalen.
type KesetaraanLookup interface {
	GetSetaraCodes(
		ctx context.Context,
		kodeMK string,
	) ([]string, error)
}

// KRSValidatorAdapter menghubungkan modul daftar tunggu
// dengan validator akademik dan repository kesetaraan.
type KRSValidatorAdapter struct {
	db         *sql.DB
	krsService *krs.Service
	kesetaraan KesetaraanLookup
}

// NewKRSValidatorAdapter membuat adapter validator.
func NewKRSValidatorAdapter(
	db *sql.DB,
	krsService *krs.Service,
	kesetaraan KesetaraanLookup,
) *KRSValidatorAdapter {
	return &KRSValidatorAdapter{
		db:         db,
		krsService: krsService,
		kesetaraan: kesetaraan,
	}
}

// Memastikan adapter memenuhi kontrak KRSValidator.
var _ KRSValidator = (*KRSValidatorAdapter)(nil)

// EvaluasiUntukDaftarTunggu memeriksa kelayakan akademik
// mahasiswa untuk masuk atau dipromosikan dari daftar tunggu.
func (v *KRSValidatorAdapter) EvaluasiUntukDaftarTunggu(
	ctx context.Context,
	nim string,
	kelasID int,
) (bool, string, error) {

	if v.db == nil {
		return false, "", errors.New("database belum tersedia")
	}

	if v.krsService == nil {
		return false, "", errors.New("service KRS belum tersedia")
	}

	if v.kesetaraan == nil {
		return false, "", errors.New("repository kesetaraan belum tersedia")
	}

	nim = strings.TrimSpace(nim)

	if nim == "" || kelasID <= 0 {
		return false, "NIM atau ID kelas tidak valid", nil
	}

	// 1. Ambil informasi kelas dan semester.
	var semesterID int
	var kodeMK string
	var statusKelas string
	var statusSemester int

	err := v.db.QueryRowContext(ctx, `
		SELECT
			k.semester_id,
			k.kode_mk,
			k.status,
			s.status_aktif
		FROM kelas k
		JOIN semester s ON s.id = k.semester_id
		WHERE k.id = ?
	`, kelasID).Scan(
		&semesterID,
		&kodeMK,
		&statusKelas,
		&statusSemester,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return false, "Kelas tidak ditemukan", nil
	}

	if err != nil {
		return false, "", err
	}

	if statusSemester != 1 {
		return false, "Semester kelas tidak aktif", nil
	}

	if statusKelas != "Dibuka" {
		return false, "Kelas tidak dibuka", nil
	}

	// 2. Ambil KRS mahasiswa pada semester kelas.
	dataKRS, err := v.krsService.AmbilKRS(
		ctx,
		nim,
		semesterID,
	)

	if errors.Is(err, krs.ErrKRSNotFound) {
		return false, "KRS semester aktif belum tersedia", nil
	}

	if err != nil {
		return false, "", err
	}

	// KRS yang dapat digunakan untuk proses daftar tunggu:
	// Draft/Ditolak pada periode pengisian,
	// Disetujui pada periode perubahan.
	switch dataKRS.Status {
	case "Draft", "Ditolak", "Disetujui":
	default:
		return false, "Status KRS belum memungkinkan perubahan kelas", nil
	}

	// 3. Ambil seluruh mata kuliah yang ekuivalen.
	kodeSetara, err := v.kesetaraan.GetSetaraCodes(
		ctx,
		kodeMK,
	)

	if err != nil {
		return false, "", err
	}

	// Pastikan kode asal ikut diperiksa dan tidak ada duplikasi.
	daftarKode := make([]string, 0, len(kodeSetara)+1)
	kodeTerlihat := make(map[string]bool)

	for _, kode := range kodeSetara {
		kode = strings.TrimSpace(kode)

		if kode != "" && !kodeTerlihat[kode] {
			daftarKode = append(daftarKode, kode)
			kodeTerlihat[kode] = true
		}
	}

	if !kodeTerlihat[kodeMK] {
		daftarKode = append(daftarKode, kodeMK)
	}

	// Periksa apakah mahasiswa sudah mengantre pada
	// mata kuliah lain yang ekuivalen.
	antreanPlaceholders := make([]string, len(daftarKode))
	antreanArgs := []any{nim, kelasID}

	for i, kode := range daftarKode {
		antreanPlaceholders[i] = "?"
		antreanArgs = append(antreanArgs, kode)
	}

	queryAntrean := `
		SELECT EXISTS (
			SELECT 1
			FROM daftar_tunggu dt
			JOIN kelas k ON k.id = dt.kelas_id
			WHERE dt.nim = ?
			  AND dt.kelas_id <> ?
			  AND dt.status IN (
				'Menunggu',
				'MenungguKonfirmasi',
				'Dilewati',
				'Diterima'
			  )
			  AND k.kode_mk IN (` +
		strings.Join(antreanPlaceholders, ",") + `)
		)
	`

	var sudahMengantre bool

	err = v.db.QueryRowContext(
		ctx,
		queryAntrean,
		antreanArgs...,
	).Scan(&sudahMengantre)

	if err != nil {
		return false, "", err
	}

	if sudahMengantre {
		return false,
			"Sudah mengantre pada mata kuliah yang ekuivalen",
			nil
	}

	// 4. Periksa apakah mata kuliah atau ekuivalennya
	// sudah terdaftar di KRS mahasiswa.
	placeholders := make([]string, len(daftarKode))
	args := []any{dataKRS.ID}

	for i, kode := range daftarKode {
		placeholders[i] = "?"
		args = append(args, kode)
	}

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM krs_detail kd
			JOIN kelas k ON k.id = kd.kelas_id
			WHERE kd.krs_id = ?
			  AND kd.status = 'Terdaftar'
			  AND k.kode_mk IN (` +
		strings.Join(placeholders, ",") + `)
		)
	`

	var sudahTerdaftar bool

	err = v.db.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(&sudahTerdaftar)

	if err != nil {
		return false, "", err
	}

	if sudahTerdaftar {
		return false,
			"Mata kuliah atau mata kuliah ekuivalennya sudah terdaftar di KRS",
			nil
	}

	// 5. Validasi prasyarat menggunakan service KRS.
	err = v.krsService.ValidasiPrasyarat(
		ctx,
		dataKRS.ID,
		kelasID,
	)

	if err != nil {
		// ValidasiPrasyarat saat ini masih menggunakan
		// errors.New(), belum menyediakan sentinel error.
		if err.Error() == "prasyarat mata kuliah belum terpenuhi" {
			return false, err.Error(), nil
		}

		// Kesalahan teknis atau struktur prasyarat
		// tidak boleh membuat mahasiswa dilewati.
		return false, "", err
	}

	// 6. Validasi bentrok jadwal.
	err = v.krsService.ValidasiBentrokJadwal(
		ctx,
		dataKRS.ID,
		kelasID,
	)

	if err != nil {
		if err.Error() ==
			"jadwal kelas bentrok dengan kelas lain yang sudah terdaftar" {
			return false, err.Error(), nil
		}

		return false, "", err
	}

	// 7. Validasi batas SKS.
	err = v.krsService.ValidasiBatasSKS(
		ctx,
		dataKRS.ID,
		kelasID,
		dataKRS.BatasSKS,
	)

	if errors.Is(err, krs.ErrBatasSKS) {
		return false, err.Error(), nil
	}

	if err != nil {
		return false, "", err
	}

	return true, "", nil
}
