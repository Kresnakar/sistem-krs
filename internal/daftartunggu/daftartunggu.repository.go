package daftartunggu

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrKelasTidakDitemukan   = errors.New("kelas tidak ditemukan")
	ErrKelasTidakDibuka      = errors.New("kelas tidak sedang dibuka")
	ErrKelasBelumPenuh       = errors.New("kelas masih memiliki kursi tersedia")
	ErrSudahMengantre        = errors.New("mahasiswa sudah berada di daftar tunggu")
	ErrSudahMendapatKursi    = errors.New("mahasiswa sudah mendapatkan tawaran atau kursi")
	ErrAntreanTidakDitemukan = errors.New("data antrean tidak ditemukan")
	ErrAntreanTidakAktif     = errors.New("antrean tidak dapat dikeluarkan")

	ErrKonfirmasiTidakTersedia         = errors.New("tawaran kursi tidak tersedia untuk dikonfirmasi")
	ErrKonfirmasiKedaluwarsa           = errors.New("batas waktu konfirmasi telah lewat")
	ErrKonfirmasiHanyaPeriodePengisian = errors.New("penerimaan langsung hanya tersedia pada periode pengisian KRS")
	ErrPeriodePengisianTutup           = errors.New("periode pengisian KRS sudah ditutup")
	ErrKRSKonfirmasiTidakDitemukan     = errors.New("KRS mahasiswa pada semester aktif tidak ditemukan")
	ErrKapasitasTidakKonsisten         = errors.New("kapasitas kelas tidak konsisten dengan jumlah kursi yang ditahan")
	ErrKelayakanKonfirmasiBerubah      = errors.New("kelayakan mahasiswa berubah sejak tawaran diberikan")
	ErrAksiKonfirmasiTidakValid        = errors.New("aksi konfirmasi harus berupa terima atau tolak")

	ErrKonfirmasiPeriodePerubahan  = errors.New("tawaran harus diproses melalui permintaan perubahan KRS")
	ErrPeriodePerubahanTutup       = errors.New("periode perubahan KRS sudah ditutup")
	ErrPermintaanPerubahanSudahAda = errors.New("permintaan tambah kelas untuk tawaran ini sudah tersedia")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// rowScanner memungkinkan fungsi scan menerima hasil Query maupun QueryRow.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanDaftarTunggu mengubah satu baris hasil query menjadi struct Go.
func scanDaftarTunggu(row rowScanner) (*DaftarTunggu, error) {
	var data DaftarTunggu
	var alasan sql.NullString
	var batasKonfirmasi sql.NullTime

	err := row.Scan(
		&data.ID,
		&data.NIM,
		&data.KelasID,
		&data.NomorAntrean,
		&data.Status,
		&data.WaktuDaftar,
		&alasan,
		&batasKonfirmasi,
		&data.PosisiAntrean,
	)
	if err != nil {
		return nil, err
	}

	if alasan.Valid {
		data.AlasanDilewati = &alasan.String
	}

	if batasKonfirmasi.Valid {
		data.BatasKonfirmasi = &batasKonfirmasi.Time
	}

	return &data, nil
}

// queryDaftarTunggu mengambil data antrean sekaligus menghitung posisi FIFO.
const queryDaftarTunggu = `
	SELECT
		dt.id,
		dt.nim,
		dt.kelas_id,
		dt.nomor_antrean,
		dt.status,
		dt.waktu_daftar,
		dt.alasan_dilewati,
		dt.batas_konfirmasi,
		CASE
			WHEN dt.status IN ('Menunggu', 'Dilewati') THEN (
				SELECT COUNT(*)
				FROM daftar_tunggu AS urutan
				WHERE urutan.kelas_id = dt.kelas_id
				  AND urutan.status IN ('Menunggu', 'Dilewati')
				  AND (
					urutan.waktu_daftar < dt.waktu_daftar
					OR (
						urutan.waktu_daftar = dt.waktu_daftar
						AND urutan.nomor_antrean <= dt.nomor_antrean
					)
				  )
			)
			ELSE 0
		END AS posisi_antrean
	FROM daftar_tunggu AS dt
`

const queryHitungKursiDitahan = `
    SELECT COUNT(*)
    FROM daftar_tunggu dt
    WHERE dt.kelas_id = ?
      AND (
        dt.status = 'MenungguKonfirmasi'
        OR (
            dt.status = 'Diterima'
            AND EXISTS (
                SELECT 1
                FROM permintaan_perubahan p
                JOIN permintaan_perubahan_detail pd
                    ON pd.permintaan_id = p.id
                JOIN krs k
                    ON k.id = p.krs_id
                WHERE pd.daftar_tunggu_id = dt.id
                  AND pd.kelas_id = dt.kelas_id
                  AND k.nim = dt.nim
                  AND p.jenis_perubahan = 'Tambah Kelas'
                  AND p.status = 'Menunggu'
            )
        )
      )
`

// ListByNIM mengambil seluruh data antrean milik seorang mahasiswa.
func (r *Repository) ListByNIM(
	ctx context.Context,
	nim string,
) ([]DaftarTunggu, error) {

	query := queryDaftarTunggu + `
		WHERE dt.nim = ?
		ORDER BY dt.waktu_daftar DESC, dt.id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, nim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daftar []DaftarTunggu

	for rows.Next() {
		data, err := scanDaftarTunggu(rows)
		if err != nil {
			return nil, err
		}

		daftar = append(daftar, *data)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if daftar == nil {
		daftar = []DaftarTunggu{}
	}

	return daftar, nil
}

// ListByKelas mengambil seluruh antrean suatu kelas.
func (r *Repository) ListByKelas(
	ctx context.Context,
	kelasID int,
) ([]DaftarTunggu, error) {

	query := queryDaftarTunggu + `
		WHERE dt.kelas_id = ?
		ORDER BY
			CASE
				WHEN dt.status IN ('Menunggu', 'Dilewati') THEN 0
				ELSE 1
			END,
			dt.waktu_daftar ASC,
			dt.nomor_antrean ASC
	`

	rows, err := r.db.QueryContext(ctx, query, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daftar []DaftarTunggu

	for rows.Next() {
		data, err := scanDaftarTunggu(rows)
		if err != nil {
			return nil, err
		}

		daftar = append(daftar, *data)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if daftar == nil {
		daftar = []DaftarTunggu{}
	}

	return daftar, nil
}

// ListKandidatPromosi mengambil semua kandidat berdasarkan FIFO.
// Status Dilewati ikut diperiksa kembali pada kesempatan promosi berikutnya.
func (r *Repository) ListKandidatPromosi(
	ctx context.Context,
	kelasID int,
) ([]DaftarTunggu, error) {

	query := queryDaftarTunggu + `
		WHERE dt.kelas_id = ?
		  AND dt.status IN ('Menunggu', 'Dilewati')
		ORDER BY dt.waktu_daftar ASC, dt.nomor_antrean ASC
	`

	rows, err := r.db.QueryContext(ctx, query, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daftar []DaftarTunggu

	for rows.Next() {
		data, err := scanDaftarTunggu(rows)
		if err != nil {
			return nil, err
		}

		daftar = append(daftar, *data)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return daftar, nil
}

// MasukAntrean mendaftarkan mahasiswa ke antrean kelas yang penuh.
func (r *Repository) MasukAntrean(
	ctx context.Context,
	nim string,
	kelasID int,
) (*DaftarTunggu, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Kunci baris kelas agar dua proses pendaftaran
	// tidak memperoleh nomor antrean yang sama.
	var kuota, terisi int
	var statusKelas string

	err = tx.QueryRowContext(ctx, `
		SELECT kuota, terisi, status
		FROM kelas
		WHERE id = ?
		FOR UPDATE
	`, kelasID).Scan(&kuota, &terisi, &statusKelas)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKelasTidakDitemukan
	}
	if err != nil {
		return nil, err
	}

	if statusKelas != "Dibuka" {
		return nil, ErrKelasTidakDibuka
	}

	// Kursi yang sedang ditahan juga dihitung sebagai terisi.
	var jumlahDitahan int

	err = tx.QueryRowContext(
		ctx,
		queryHitungKursiDitahan,
		kelasID,
	).Scan(&jumlahDitahan)

	if err != nil {
		return nil, err
	}

	if kuota <= 0 || terisi+jumlahDitahan < kuota {
		return nil, ErrKelasBelumPenuh
	}

	// Periksa apakah mahasiswa sudah memiliki baris antrean.
	var existingID int
	var existingStatus string

	err = tx.QueryRowContext(ctx, `
		SELECT id, status
		FROM daftar_tunggu
		WHERE nim = ? AND kelas_id = ?
		FOR UPDATE
	`, nim, kelasID).Scan(&existingID, &existingStatus)

	exists := err == nil

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if exists {
		switch StatusDaftarTunggu(existingStatus) {
		case StatusMenunggu,
			StatusMenungguKonfirmasi,
			StatusDilewati:
			return nil, ErrSudahMengantre

		case StatusDiterima:
			return nil, ErrSudahMendapatKursi
		}
	}

	// Nomor baru selalu lebih besar daripada nomor sebelumnya,
	// termasuk catatan antrean yang sudah berakhir.
	var nomorAntrean int

	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(nomor_antrean), 0) + 1
		FROM daftar_tunggu
		WHERE kelas_id = ?
	`, kelasID).Scan(&nomorAntrean)

	if err != nil {
		return nil, err
	}

	var antreanID int64

	if exists {
		// Gunakan kembali baris lama karena terdapat UNIQUE (nim, kelas_id).
		// Waktu dan nomor antrean diperbarui sebagai pendaftaran baru.
		_, err = tx.ExecContext(ctx, `
			UPDATE daftar_tunggu
			SET nomor_antrean = ?,
				status = 'Menunggu',
				alasan_dilewati = NULL,
				waktu_daftar = CURRENT_TIMESTAMP,
				batas_konfirmasi = NULL
			WHERE id = ?
		`, nomorAntrean, existingID)

		if err != nil {
			return nil, err
		}

		antreanID = int64(existingID)
	} else {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO daftar_tunggu (
				nim,
				kelas_id,
				nomor_antrean,
				status,
				waktu_daftar
			)
			VALUES (?, ?, ?, 'Menunggu', CURRENT_TIMESTAMP)
		`, nim, kelasID, nomorAntrean)

		if err != nil {
			return nil, err
		}

		antreanID, err = result.LastInsertId()
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.getByID(ctx, int(antreanID))
}

// KeluarAntrean mengeluarkan mahasiswa dari antrean miliknya sendiri.
func (r *Repository) KeluarAntrean(
	ctx context.Context,
	nim string,
	antreanID int,
) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string

	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM daftar_tunggu
		WHERE id = ? AND nim = ?
		FOR UPDATE
	`, antreanID, nim).Scan(&status)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrAntreanTidakDitemukan
	}
	if err != nil {
		return err
	}

	if status != string(StatusMenunggu) &&
		status != string(StatusDilewati) {
		return ErrAntreanTidakAktif
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE daftar_tunggu
		SET status = 'Keluar'
		WHERE id = ? AND nim = ?
	`, antreanID, nim)

	if err != nil {
		return err
	}

	return tx.Commit()
}

type dataTawaran struct {
	KelasID     int
	SemesterID  int
	Status      string
	Kedaluwarsa bool
	Kuota       int
	Terisi      int
	StatusKelas string
}

// kunciTawaran mengunci kelas terlebih dahulu, kemudian antrean.
// Urutan ini dibuat konsisten dengan proses promosi.
func (r *Repository) kunciTawaran(
	ctx context.Context,
	tx *sql.Tx,
	nim string,
	antreanID int,
) (*dataTawaran, error) {

	var data dataTawaran

	// Ambil kelas_id terlebih dahulu tanpa mengunci baris antrean.
	err := tx.QueryRowContext(ctx, `
		SELECT kelas_id
		FROM daftar_tunggu
		WHERE id = ? AND nim = ?
	`, antreanID, nim).Scan(&data.KelasID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAntreanTidakDitemukan
	}
	if err != nil {
		return nil, err
	}

	// Kunci baris kelas.
	err = tx.QueryRowContext(ctx, `
		SELECT kuota, terisi, status, semester_id
		FROM kelas
		WHERE id = ?
		FOR UPDATE
	`, data.KelasID).Scan(
		&data.Kuota,
		&data.Terisi,
		&data.StatusKelas,
		&data.SemesterID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKelasTidakDitemukan
	}
	if err != nil {
		return nil, err
	}

	// Setelah kelas terkunci, kunci baris antrean.
	var sudahKedaluwarsa int

	err = tx.QueryRowContext(ctx, `
		SELECT
			status,
			CASE
				WHEN batas_konfirmasi IS NULL
				  OR batas_konfirmasi <= NOW()
				THEN 1
				ELSE 0
			END
		FROM daftar_tunggu
		WHERE id = ? AND nim = ?
		FOR UPDATE
	`, antreanID, nim).Scan(
		&data.Status,
		&sudahKedaluwarsa,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAntreanTidakDitemukan
	}
	if err != nil {
		return nil, err
	}

	data.Kedaluwarsa = sudahKedaluwarsa == 1

	return &data, nil
}

// ubahStatusTawaran mengubah status antrean yang sedang dikunci.
func ubahStatusTawaran(
	ctx context.Context,
	tx *sql.Tx,
	nim string,
	antreanID int,
	status string,
	alasan *string,
) error {

	result, err := tx.ExecContext(ctx, `
		UPDATE daftar_tunggu
		SET status = ?,
		    alasan_dilewati = ?,
		    batas_konfirmasi = NULL
		WHERE id = ?
		  AND nim = ?
		  AND status = 'MenungguKonfirmasi'
	`, status, alasan, antreanID, nim)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrKonfirmasiTidakTersedia
	}

	return nil
}

// terimaTawaran memvalidasi ulang kelayakan akademik,
// kemudian menentukan jalur penerimaan berdasarkan status KRS.
// terimaTawaran memvalidasi ulang kelayakan akademik,
// kemudian menentukan jalur penerimaan berdasarkan status KRS.
func (r *Repository) TerimaTawaran(
	ctx context.Context,
	nim string,
	antreanID int,
) (int, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Kunci kelas dan antrean.
	tawaran, err := r.kunciTawaran(ctx, tx, nim, antreanID)
	if err != nil {
		return 0, err
	}

	if tawaran.Status != string(StatusMenungguKonfirmasi) {
		return tawaran.KelasID, ErrKonfirmasiTidakTersedia
	}

	// 2. Jika sudah melewati tenggat, akhiri penahanan kursi.
	if tawaran.Kedaluwarsa {
		if err := ubahStatusTawaran(
			ctx,
			tx,
			nim,
			antreanID,
			string(StatusKedaluwarsa),
			nil,
		); err != nil {
			return tawaran.KelasID, err
		}

		if err := tx.Commit(); err != nil {
			return tawaran.KelasID, err
		}

		return tawaran.KelasID, ErrKonfirmasiKedaluwarsa
	}

	if tawaran.StatusKelas != "Dibuka" {
		return tawaran.KelasID, ErrKelasTidakDibuka
	}

	// 3. Kunci dan periksa KRS mahasiswa pada semester aktif.
	var krsID int
	var statusKRS string
	var batasSKS int
	var periodeTerbuka int

	err = tx.QueryRowContext(ctx, `
		SELECT
			k.id,
			k.status,
			k.batas_sks,
			CASE
				WHEN NOW() >= p.mulai_krs
				 AND NOW() <= p.selesai_krs
				THEN 1
				ELSE 0
			END
		FROM krs k
		JOIN semester s ON s.id = k.semester_id
		JOIN periode_krs p ON p.semester_id = k.semester_id
		WHERE k.nim = ?
		  AND s.status_aktif = 1
		FOR UPDATE
	`, nim).Scan(
		&krsID,
		&statusKRS,
		&batasSKS,
		&periodeTerbuka,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return tawaran.KelasID, ErrKRSKonfirmasiTidakDitemukan
	}
	if err != nil {
		return tawaran.KelasID, err
	}

	// Jalur ini hanya menangani pengisian KRS.
	// KRS Disetujui diproses melalui permintaan perubahan KRS.
	if statusKRS != "Draft" && statusKRS != "Ditolak" {
		return tawaran.KelasID, ErrKonfirmasiHanyaPeriodePengisian
	}

	if periodeTerbuka != 1 {
		return tawaran.KelasID, ErrPeriodePengisianTutup
	}

	if batasSKS <= 0 {
		return tawaran.KelasID, ErrKelayakanKonfirmasiBerubah
	}

	// 4. Pastikan kapasitas kursi yang ditahan masih konsisten.
	var jumlahDitahan int

	err = tx.QueryRowContext(
		ctx,
		queryHitungKursiDitahan,
		tawaran.KelasID,
	).Scan(&jumlahDitahan)

	if err != nil {
		return tawaran.KelasID, err
	}

	if jumlahDitahan <= 0 ||
		tawaran.Kuota <= 0 ||
		tawaran.Terisi+jumlahDitahan > tawaran.Kuota {
		return tawaran.KelasID, ErrKapasitasTidakKonsisten
	}

	// 5. Periksa apakah kelas sudah terdaftar di KRS.
	var sudahTerdaftar bool

	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM krs_detail
			WHERE krs_id = ?
			  AND kelas_id = ?
			  AND status = 'Terdaftar'
		)
	`, krsID, tawaran.KelasID).Scan(&sudahTerdaftar)

	if err != nil {
		return tawaran.KelasID, err
	}

	if sudahTerdaftar {
		return tawaran.KelasID, ErrKelayakanKonfirmasiBerubah
	}

	// 6. Masukkan kelas ke KRS.
	// Jika pernah dibatalkan, aktifkan kembali baris detail tersebut.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO krs_detail (
			krs_id,
			kelas_id,
			status,
			jenis_perubahan,
			status_persetujuan
		)
		VALUES (
			?,
			?,
			'Terdaftar',
			'Awal',
			'Menunggu Persetujuan'
		)
		ON DUPLICATE KEY UPDATE
			status = 'Terdaftar',
			jenis_perubahan = 'Awal',
			status_persetujuan = 'Menunggu Persetujuan'
	`, krsID, tawaran.KelasID)

	if err != nil {
		return tawaran.KelasID, err
	}

	// 7. Hitung ulang total SKS dari seluruh kelas aktif di KRS.
	var totalSKS int

	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(mk.sks), 0)
		FROM krs_detail kd
		JOIN kelas k ON k.id = kd.kelas_id
		JOIN mata_kuliah mk ON mk.kode_mk = k.kode_mk
		WHERE kd.krs_id = ?
		  AND kd.status = 'Terdaftar'
	`, krsID).Scan(&totalSKS)

	if err != nil {
		return tawaran.KelasID, err
	}

	if totalSKS > batasSKS {
		return tawaran.KelasID, ErrKelayakanKonfirmasiBerubah
	}

	// 8. Ubah kursi yang tadinya ditahan menjadi kursi terisi.
	result, err := tx.ExecContext(ctx, `
		UPDATE kelas
		SET terisi = terisi + 1
		WHERE id = ?
		  AND terisi < kuota
	`, tawaran.KelasID)

	if err != nil {
		return tawaran.KelasID, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return tawaran.KelasID, err
	}

	if affected == 0 {
		return tawaran.KelasID, ErrKapasitasTidakKonsisten
	}

	// 9. KRS kembali menjadi Draft.
	_, err = tx.ExecContext(ctx, `
		UPDATE krs
		SET total_sks = ?,
		    status = 'Draft',
		    catatan_penolakan = NULL
		WHERE id = ?
	`, totalSKS, krsID)

	if err != nil {
		return tawaran.KelasID, err
	}

	// 10. Tawaran antrean dinyatakan diterima.
	if err := ubahStatusTawaran(
		ctx,
		tx,
		nim,
		antreanID,
		string(StatusDiterima),
		nil,
	); err != nil {
		return tawaran.KelasID, err
	}

	// 11. Commit seluruh perubahan secara atomik.
	if err := tx.Commit(); err != nil {
		return tawaran.KelasID, err
	}

	return tawaran.KelasID, nil
}

// TerimaTawaranPerubahan membuat permintaan tambah kelas
// ketika tawaran antrean diterima pada periode perubahan KRS.
// Kursi tetap ditahan. Fungsi ini tidak mengubah krs_detail,
// total_sks, status KRS, maupun kelas.terisi.
func (r *Repository) TerimaTawaranPerubahan(
	ctx context.Context,
	nim string,
	antreanID int,
) (int, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Kunci kelas dan antrean.
	tawaran, err := r.kunciTawaran(
		ctx,
		tx,
		nim,
		antreanID,
	)
	if err != nil {
		return 0, err
	}

	if tawaran.Status != string(StatusMenungguKonfirmasi) {
		return tawaran.KelasID, ErrKonfirmasiTidakTersedia
	}

	if tawaran.Kedaluwarsa {
		if err := ubahStatusTawaran(
			ctx,
			tx,
			nim,
			antreanID,
			string(StatusKedaluwarsa),
			nil,
		); err != nil {
			return tawaran.KelasID, err
		}

		if err := tx.Commit(); err != nil {
			return tawaran.KelasID, err
		}

		return tawaran.KelasID, ErrKonfirmasiKedaluwarsa
	}

	if tawaran.StatusKelas != "Dibuka" {
		return tawaran.KelasID, ErrKelasTidakDibuka
	}

	// 2. Kunci KRS mahasiswa pada semester kelas tersebut.
	var krsID int
	var statusKRS string
	var periodeTerbuka int

	err = tx.QueryRowContext(ctx, `
		SELECT
			k.id,
			k.status,
			CASE
				WHEN NOW() >= p.mulai_perubahan
				 AND NOW() <= p.selesai_perubahan
				THEN 1
				ELSE 0
			END
		FROM krs k
		JOIN semester s
			ON s.id = k.semester_id
		JOIN periode_krs p
			ON p.semester_id = k.semester_id
		WHERE k.nim = ?
		  AND k.semester_id = ?
		  AND s.status_aktif = 1
		FOR UPDATE
	`, nim, tawaran.SemesterID).Scan(
		&krsID,
		&statusKRS,
		&periodeTerbuka,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return tawaran.KelasID, ErrKRSKonfirmasiTidakDitemukan
	}
	if err != nil {
		return tawaran.KelasID, err
	}

	if statusKRS != "Disetujui" {
		return tawaran.KelasID, ErrKelayakanKonfirmasiBerubah
	}

	if periodeTerbuka != 1 {
		return tawaran.KelasID, ErrPeriodePerubahanTutup
	}

	// 3. Pastikan kursi yang ditahan masih konsisten.
	var jumlahDitahan int

	err = tx.QueryRowContext(
		ctx,
		queryHitungKursiDitahan,
		tawaran.KelasID,
	).Scan(&jumlahDitahan)

	if err != nil {
		return tawaran.KelasID, err
	}

	if jumlahDitahan <= 0 ||
		tawaran.Kuota <= 0 ||
		tawaran.Terisi+jumlahDitahan > tawaran.Kuota {
		return tawaran.KelasID, ErrKapasitasTidakKonsisten
	}

	// 4. Periksa apakah kelas sudah ada dalam KRS.
	var sudahTerdaftar bool

	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM krs_detail
			WHERE krs_id = ?
			  AND kelas_id = ?
			  AND status = 'Terdaftar'
		)
	`, krsID, tawaran.KelasID).Scan(&sudahTerdaftar)

	if err != nil {
		return tawaran.KelasID, err
	}

	if sudahTerdaftar {
		return tawaran.KelasID, ErrKelayakanKonfirmasiBerubah
	}

	// 5. Cegah permintaan tambah kelas yang masih menunggu
	// untuk kelas yang sama pada KRS ini.
	var permintaanSudahAda bool

	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM permintaan_perubahan p
			JOIN permintaan_perubahan_detail pd
				ON pd.permintaan_id = p.id
			WHERE p.krs_id = ?
			  AND pd.kelas_id = ?
			  AND p.jenis_perubahan = 'Tambah Kelas'
			  AND p.status = 'Menunggu'
		)
	`, krsID, tawaran.KelasID).Scan(&permintaanSudahAda)

	if err != nil {
		return tawaran.KelasID, err
	}

	if permintaanSudahAda {
		return tawaran.KelasID, ErrPermintaanPerubahanSudahAda
	}

	// 6. Buat permintaan perubahan.
	result, err := tx.ExecContext(ctx, `
		INSERT INTO permintaan_perubahan (
			krs_id,
			jenis_perubahan,
			alasan,
			status
		)
		VALUES (
			?,
			'Tambah Kelas',
			'Penerimaan tawaran daftar tunggu',
			'Menunggu'
		)
	`, krsID)

	if err != nil {
		return tawaran.KelasID, err
	}

	permintaanID, err := result.LastInsertId()
	if err != nil {
		return tawaran.KelasID, err
	}

	// 7. Hubungkan detail permintaan dengan antrean asal.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO permintaan_perubahan_detail (
			permintaan_id,
			kelas_id,
			daftar_tunggu_id
		)
		VALUES (?, ?, ?)
	`, permintaanID, tawaran.KelasID, antreanID)

	if err != nil {
		return tawaran.KelasID, err
	}

	// 8. Ubah status antrean menjadi Diterima.
	// Kursi tetap ditahan karena permintaan masih Menunggu.
	if err := ubahStatusTawaran(
		ctx,
		tx,
		nim,
		antreanID,
		string(StatusDiterima),
		nil,
	); err != nil {
		return tawaran.KelasID, err
	}

	// 9. Commit seluruh perubahan.
	if err := tx.Commit(); err != nil {
		return tawaran.KelasID, err
	}

	return tawaran.KelasID, nil
}

func (r *Repository) TolakTawaran(
	ctx context.Context,
	nim string,
	antreanID int,
) (int, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	tawaran, err := r.kunciTawaran(ctx, tx, nim, antreanID)
	if err != nil {
		return 0, err
	}

	if tawaran.Status != string(StatusMenungguKonfirmasi) {
		return tawaran.KelasID, ErrKonfirmasiTidakTersedia
	}

	statusAkhir := string(StatusDitolak)
	var hasilErr error

	if tawaran.Kedaluwarsa {
		statusAkhir = string(StatusKedaluwarsa)
		hasilErr = ErrKonfirmasiKedaluwarsa
	}

	if err := ubahStatusTawaran(
		ctx,
		tx,
		nim,
		antreanID,
		statusAkhir,
		nil,
	); err != nil {
		return tawaran.KelasID, err
	}

	if err := tx.Commit(); err != nil {
		return tawaran.KelasID, err
	}

	return tawaran.KelasID, hasilErr
}

func (r *Repository) LewatiTawaran(
	ctx context.Context,
	nim string,
	antreanID int,
	alasan string,
) (int, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	tawaran, err := r.kunciTawaran(ctx, tx, nim, antreanID)
	if err != nil {
		return 0, err
	}

	if tawaran.Status != string(StatusMenungguKonfirmasi) {
		return tawaran.KelasID, ErrKonfirmasiTidakTersedia
	}

	if tawaran.Kedaluwarsa {
		if err := ubahStatusTawaran(
			ctx,
			tx,
			nim,
			antreanID,
			string(StatusKedaluwarsa),
			nil,
		); err != nil {
			return tawaran.KelasID, err
		}

		if err := tx.Commit(); err != nil {
			return tawaran.KelasID, err
		}

		return tawaran.KelasID, ErrKonfirmasiKedaluwarsa
	}

	if strings.TrimSpace(alasan) == "" {
		alasan = "Tidak lagi memenuhi persyaratan akademik"
	}

	if err := ubahStatusTawaran(
		ctx,
		tx,
		nim,
		antreanID,
		string(StatusDilewati),
		&alasan,
	); err != nil {
		return tawaran.KelasID, err
	}

	if err := tx.Commit(); err != nil {
		return tawaran.KelasID, err
	}

	return tawaran.KelasID, nil
}

// getByID mengambil satu antrean berdasarkan ID.
func (r *Repository) getByID(
	ctx context.Context,
	antreanID int,
) (*DaftarTunggu, error) {

	query := queryDaftarTunggu + ` WHERE dt.id = ?`

	data, err := scanDaftarTunggu(
		r.db.QueryRowContext(ctx, query, antreanID),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAntreanTidakDitemukan
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// transaksiPromosi merepresentasikan transaksi database
// khusus untuk proses promosi daftar tunggu.
type transaksiPromosi struct {
	tx *sql.Tx
}

// MulaiTransaksiPromosi memulai transaksi baru.
func (r *Repository) MulaiTransaksiPromosi(
	ctx context.Context,
) (*transaksiPromosi, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &transaksiPromosi{tx: tx}, nil
}

// KunciKelas mengambil informasi kapasitas sekaligus
// mengunci baris kelas selama transaksi berlangsung.
func (t *transaksiPromosi) KunciKelas(
	ctx context.Context,
	kelasID int,
) (*KapasitasKelas, error) {

	var kapasitas KapasitasKelas
	var status string

	err := t.tx.QueryRowContext(ctx, `
        SELECT kuota, terisi, status
        FROM kelas
        WHERE id = ?
        FOR UPDATE
    `, kelasID).Scan(
		&kapasitas.Kuota,
		&kapasitas.Terisi,
		&status,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKelasTidakDitemukan
	}

	if err != nil {
		return nil, err
	}

	if status != "Dibuka" {
		return nil, ErrKelasTidakDibuka
	}

	return &kapasitas, nil
}

// HitungKursiDitahan menghitung kursi yang sedang
// menunggu konfirmasi mahasiswa.
func (t *transaksiPromosi) HitungKursiDitahan(
	ctx context.Context,
	kelasID int,
) (int, error) {

	var jumlah int

	err := t.tx.QueryRowContext(
		ctx,
		queryHitungKursiDitahan,
		kelasID,
	).Scan(&jumlah)

	return jumlah, err
}

// DaftarKandidat mengambil antrean yang dapat dipromosikan.
//
// Mahasiswa berstatus Dilewati tetap diperiksa kembali
// pada kesempatan promosi berikutnya.
func (t *transaksiPromosi) DaftarKandidat(
	ctx context.Context,
	kelasID int,
) ([]DaftarTunggu, error) {

	query := `
        SELECT
            id,
            nim,
            kelas_id,
            nomor_antrean,
            status,
            waktu_daftar,
            alasan_dilewati,
            batas_konfirmasi,
            0 AS posisi_antrean
        FROM daftar_tunggu
        WHERE kelas_id = ?
          AND status IN ('Menunggu', 'Dilewati')
        ORDER BY waktu_daftar ASC, id ASC
        FOR UPDATE
    `

	rows, err := t.tx.QueryContext(ctx, query, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daftar []DaftarTunggu

	for rows.Next() {
		data, err := scanDaftarTunggu(rows)
		if err != nil {
			return nil, err
		}

		daftar = append(daftar, *data)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return daftar, nil
}

// TandaiDilewati menyimpan alasan mahasiswa tidak memenuhi
// syarat ketika diperiksa dalam proses promosi.
func (t *transaksiPromosi) TandaiDilewati(
	ctx context.Context,
	antreanID int,
	alasan string,
) error {

	_, err := t.tx.ExecContext(ctx, `
        UPDATE daftar_tunggu
        SET status = 'Dilewati',
            alasan_dilewati = ?,
            batas_konfirmasi = NULL
        WHERE id = ?
          AND status IN ('Menunggu', 'Dilewati')
    `, alasan, antreanID)

	return err
}

// TahanKursi mengalokasikan satu kursi kepada mahasiswa
// dan mengubah status menjadi MenungguKonfirmasi.
func (t *transaksiPromosi) TahanKursi(
	ctx context.Context,
	antreanID int,
	batasKonfirmasi time.Time,
) error {

	result, err := t.tx.ExecContext(ctx, `
        UPDATE daftar_tunggu
        SET status = 'MenungguKonfirmasi',
            alasan_dilewati = NULL,
            batas_konfirmasi = ?
        WHERE id = ?
          AND status IN ('Menunggu', 'Dilewati')
    `, batasKonfirmasi, antreanID)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrAntreanTidakAktif
	}

	return nil
}

// Commit mengesahkan seluruh perubahan transaksi.
func (t *transaksiPromosi) Commit() error {
	return t.tx.Commit()
}

// Rollback membatalkan seluruh perubahan transaksi.
func (t *transaksiPromosi) Rollback() error {
	return t.tx.Rollback()
}

func (r *Repository) GetKelasIDByAntrean(
	ctx context.Context,
	nim string,
	antreanID int,
) (int, error) {

	var kelasID int

	err := r.db.QueryRowContext(ctx, `
		SELECT kelas_id
		FROM daftar_tunggu
		WHERE id = ? AND nim = ?
	`, antreanID, nim).Scan(&kelasID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrAntreanTidakDitemukan
	}

	if err != nil {
		return 0, err
	}

	return kelasID, nil
}
