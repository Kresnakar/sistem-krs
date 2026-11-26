package kelas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrKelasNotFound = errors.New("kelas paralel tidak ditemukan")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create menyimpan data kelas baru beserta jadwalnya dalam satu transaksi database.
func (r *Repository) Create(ctx context.Context, kelas *KelasParalel) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Kelas
	queryKelas := `INSERT INTO kelas (kode_mk, semester_id, kode_paralel, dosen_pengampu, kuota, status) VALUES (?, ?, ?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, queryKelas, kelas.KodeMK, kelas.SemesterID, kelas.KodeParalel, kelas.DosenPengampu, kelas.Kuota, kelas.Status)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	kelas.ID = int(id)

	// 2. Insert Jadwal (jika ada)
	if len(kelas.Jadwal) > 0 {
		queryJadwal := `INSERT INTO jadwal_kelas (kelas_id, hari, jam_mulai, jam_selesai, ruang) VALUES (?, ?, ?, ?, ?)`
		for _, j := range kelas.Jadwal {
			_, err = tx.ExecContext(ctx, queryJadwal, kelas.ID, j.Hari, j.JamMulai, j.JamSelesai, j.Ruang)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetBySemester mengambil daftar kelas untuk suatu semester beserta detail MK dan jadwalnya.
func (r *Repository) GetBySemester(ctx context.Context, semesterID int) ([]*KelasParalel, error) {
	// Query utama mengambil data kelas dan MK
	query := `
        SELECT k.id, k.kode_mk, mk.nama, k.semester_id, s.tahun_akademik, s.jenis,
               k.kode_paralel, k.dosen_pengampu, k.kuota, k.terisi, k.status
        FROM kelas k
        JOIN mata_kuliah mk ON k.kode_mk = mk.kode_mk
        JOIN semester s ON k.semester_id = s.id
        WHERE k.semester_id = ? AND k.status != 'Dibatalkan'
        ORDER BY mk.kode_mk ASC, k.kode_paralel ASC
    `
	rows, err := r.db.QueryContext(ctx, query, semesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listKelas []*KelasParalel
	mapKelas := make(map[int]*KelasParalel)
	var kelasIDs []interface{}

	for rows.Next() {
		var k KelasParalel
		var tahun, jenis string
		if err := rows.Scan(
			&k.ID, &k.KodeMK, &k.NamaMK, &k.SemesterID, &tahun, &jenis,
			&k.KodeParalel, &k.DosenPengampu, &k.Kuota, &k.Terisi, &k.Status,
		); err != nil {
			return nil, err
		}
		k.TahunSemester = fmt.Sprintf("%s %s", tahun, jenis)
		mapKelas[k.ID] = &k
		listKelas = append(listKelas, &k)
		kelasIDs = append(kelasIDs, k.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(listKelas) == 0 {
		return []*KelasParalel{}, nil
	}

	// Query kedua mengambil seluruh jadwal untuk kelas-kelas yang ditemukan dalam 1 kali jalan (optimasi)
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(kelasIDs)), ",")
	queryJadwal := fmt.Sprintf(
		`SELECT id, kelas_id, hari, TIME_FORMAT(jam_mulai, '%%H:%%i'), TIME_FORMAT(jam_selesai, '%%H:%%i'), COALESCE(ruang, '') FROM jadwal_kelas WHERE kelas_id IN (%s)`,
		placeholders,
	)

	rowsJ, err := r.db.QueryContext(ctx, queryJadwal, kelasIDs...)
	if err != nil {
		return nil, err
	}
	defer rowsJ.Close()

	for rowsJ.Next() {
		var j JadwalSesi
		if err := rowsJ.Scan(&j.ID, &j.KelasID, &j.Hari, &j.JamMulai, &j.JamSelesai, &j.Ruang); err != nil {
			return nil, err
		}
		if k, ok := mapKelas[j.KelasID]; ok {
			k.Jadwal = append(k.Jadwal, j)
		}
	}
	if err := rowsJ.Err(); err != nil {
		return nil, err
	}
	return listKelas, nil
}

// Update mengubah data dasar kelas dan memperbarui jadwalnya (hapus lama, insert baru).
func (r *Repository) Update(ctx context.Context, kelasID int, updateData *KelasParalel) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update data Kelas
	query := `UPDATE kelas SET dosen_pengampu = ?, kuota = ?, status = ? WHERE id = ?`
	res, err := tx.ExecContext(ctx, query, updateData.DosenPengampu, updateData.Kuota, updateData.Status, kelasID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrKelasNotFound
	}

	// 2. Hapus Jadwal lama
	_, err = tx.ExecContext(ctx, `DELETE FROM jadwal_kelas WHERE kelas_id = ?`, kelasID)
	if err != nil {
		return err
	}

	// 3. Insert Jadwal baru
	if len(updateData.Jadwal) > 0 {
		queryJadwal := `INSERT INTO jadwal_kelas (kelas_id, hari, jam_mulai, jam_selesai, ruang) VALUES (?, ?, ?, ?, ?)`
		for _, j := range updateData.Jadwal {
			_, err = tx.ExecContext(ctx, queryJadwal, kelasID, j.Hari, j.JamMulai, j.JamSelesai, j.Ruang)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// Delete membatalkan kelas secara fisik (diperbolehkan jika terisi=0 oleh constraint DB FK).
func (r *Repository) Delete(ctx context.Context, kelasID int) error {
	// Cek apakah ada mahasiswa terdaftar (FR-02: tidak boleh dihapus jika sudah ada isi)
	var terisi int
	err := r.db.QueryRowContext(ctx, `SELECT terisi FROM kelas WHERE id = ?`, kelasID).Scan(&terisi)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrKelasNotFound
	}
	if err != nil {
		return err
	}
	if terisi > 0 {
		return fmt.Errorf("kelas tidak dapat dihapus, sudah terisi %d mahasiswa", terisi)
	}

	// Hapus fisik (jadwal terhapus otomatis via ON DELETE CASCADE)
	_, err = r.db.ExecContext(ctx, `DELETE FROM kelas WHERE id = ?`, kelasID)
	return err
}
