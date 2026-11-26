package nilai

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// UpsertNilai menginput/mengupdate nilai mahasiswa.
func (r *Repository) UpsertNilai(ctx context.Context, n *NilaiMahasiswa) error {
	// Ambil bobot berdasarkan nilai huruf
	var bobot float64
	err := r.db.QueryRowContext(ctx, `SELECT bobot FROM skala_nilai WHERE nilai_huruf = ?`, n.NilaiHuruf).Scan(&bobot)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("nilai huruf '%s' tidak valid", n.NilaiHuruf)
	}
	if err != nil {
		return err
	}

	// Gunakan ON DUPLICATE KEY UPDATE (karena ada UNIQUE KEY di tabel)
	query := `
        INSERT INTO riwayat_nilai (nim, kode_mk, semester_id, nilai_huruf)
        VALUES (?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
            nilai_huruf = VALUES(nilai_huruf)
    `
	_, err = r.db.ExecContext(ctx, query, n.NIM, n.KodeMK, n.SemesterID, n.NilaiHuruf)
	return err
}

// GetTranskripTerbaik (Inti Tugas Anda - FR-03)
// Mengambil daftar nilai terbaik dari setiap MK yang pernah diambil mahasiswa.
func (r *Repository) GetTranskripTerbaik(ctx context.Context, nim string) ([]TranskripEntry, error) {
	// CTE untuk mencari bobot tertinggi per MK
	query := `
        WITH RankedNilai AS (
            SELECT
                rn.nim, rn.kode_mk, rn.nilai_huruf, sn.bobot,
                ROW_NUMBER() OVER(PARTITION BY rn.kode_mk ORDER BY sn.bobot DESC) as rn_순위
            FROM riwayat_nilai rn
            JOIN skala_nilai sn ON rn.nilai_huruf = sn.nilai_huruf
            WHERE rn.nim = ?
        )
        SELECT
            rn.nim, rn.kode_mk, mk.nama, mk.sks, rn.nilai_huruf, rn.bobot
        FROM RankedNilai rn
        JOIN mata_kuliah mk ON rn.kode_mk = mk.kode_mk
        WHERE rn.rn_순위 = 1
        ORDER BY mk.kode_mk ASC
    `
	rows, err := r.db.QueryContext(ctx, query, nim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transkrip []TranskripEntry
	for rows.Next() {
		var t TranskripEntry
		if err := rows.Scan(&t.NIM, &t.KodeMK, &t.NamaMK, &t.SKS, &t.NilaiHuruf, &t.Bobot); err != nil {
			return nil, err
		}
		transkrip = append(transkrip, t)
	}
	return transkrip, rows.Err()
}

// GetRiwayatSemester (Untuk Perhitungan IPS - FR-04)
// Mengambil seluruh nilai mahasiswa pada semester spesifik.
func (r *Repository) GetRiwayatSemester(ctx context.Context, nim string, semesterID int) ([]NilaiMahasiswa, error) {
	query := `
        SELECT rn.kode_mk, mk.sks, rn.nilai_huruf, sn.bobot
        FROM riwayat_nilai rn
        JOIN mata_kuliah mk ON rn.kode_mk = mk.kode_mk
        JOIN skala_nilai sn ON rn.nilai_huruf = sn.nilai_huruf
        WHERE rn.nim = ? AND rn.semester_id = ?
    `
	rows, err := r.db.QueryContext(ctx, query, nim, semesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var riwayat []NilaiMahasiswa
	for rows.Next() {
		var n NilaiMahasiswa
		if err := rows.Scan(&n.KodeMK, &n.SKS, &n.NilaiHuruf, &n.Bobot); err != nil {
			return nil, err
		}
		riwayat = append(riwayat, n)
	}
	return riwayat, rows.Err()
}

// GetInfoSemester mengambil detail tahun akademik dan jenis semester.
// GetInfoSemester mengambil detail tahun akademik dan jenis semester.
func (r *Repository) GetInfoSemester(ctx context.Context, semesterID int) (tahun, jenis string, err error) {
	query := `SELECT tahun_akademik, jenis FROM semester WHERE id = ?` // <-- Fixed: Added colon for declaration
	err = r.db.QueryRowContext(ctx, query, semesterID).Scan(&tahun, &jenis)
	return
}
