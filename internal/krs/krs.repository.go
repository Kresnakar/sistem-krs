package krs

import (
	"context"
	"database/sql"
	"errors"
)

var ErrKRSNotFound = errors.New("data KRS tidak ditemukan")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByNIMSemester(
	ctx context.Context,
	nim string,
	semesterID int,
) (*KRS, error) {

	query := `
        SELECT
            id, nim, semester_id, status,
            dosen_wali_id, ips_acuan, batas_sks,
            total_sks, diajukan_pada,
            disetujui_pada, catatan_penolakan
        FROM krs
        WHERE nim = ? AND semester_id = ?
    `

	var data KRS
	var dosenWaliID sql.NullInt64
	var diajukanPada sql.NullTime
	var disetujuiPada sql.NullTime
	var catatan sql.NullString

	err := r.db.QueryRowContext(
		ctx, query, nim, semesterID,
	).Scan(
		&data.ID,
		&data.NIM,
		&data.SemesterID,
		&data.Status,
		&dosenWaliID,
		&data.IPSAcuan,
		&data.BatasSKS,
		&data.TotalSKS,
		&diajukanPada,
		&disetujuiPada,
		&catatan,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKRSNotFound
	}

	if err != nil {
		return nil, err
	}

	if dosenWaliID.Valid {
		id := int(dosenWaliID.Int64)
		data.DosenWaliID = &id
	}

	if diajukanPada.Valid {
		data.DiajukanPada = &diajukanPada.Time
	}

	if disetujuiPada.Valid {
		data.DisetujuiPada = &disetujuiPada.Time
	}

	if catatan.Valid {
		data.CatatanPenolakan = &catatan.String
	}

	return &data, nil
}

func (r *Repository) GetDetail(
	ctx context.Context,
	krsID int,
) ([]KRSDetail, error) {

	query := `
        SELECT id, krs_id, kelas_id, status,
               jenis_perubahan, status_persetujuan
        FROM krs_detail
        WHERE krs_id = ?
        ORDER BY id ASC
    `

	rows, err := r.db.QueryContext(ctx, query, krsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []KRSDetail

	for rows.Next() {
		var d KRSDetail

		err := rows.Scan(
			&d.ID,
			&d.KRSID,
			&d.KelasID,
			&d.Status,
			&d.JenisPerubahan,
			&d.StatusPersetujuan,
		)
		if err != nil {
			return nil, err
		}

		details = append(details, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return details, nil
}

func (r *Repository) TambahDetail(
	ctx context.Context,
	krsID int,
	kelasID int,
) error {

	query := `
        INSERT INTO krs_detail (
            krs_id,
            kelas_id,
            status,
            jenis_perubahan,
            status_persetujuan
        )
        VALUES (?, ?, 'Terdaftar', 'Awal', 'Final')
        ON DUPLICATE KEY UPDATE
            status = 'Terdaftar',
            jenis_perubahan = 'Tambah',
            status_persetujuan = 'Final'
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		krsID,
		kelasID,
	)

	return err
}

func (r *Repository) HapusDetail(
	ctx context.Context,
	krsID int,
	kelasID int,
) error {

	query := `
        UPDATE krs_detail
        SET status = 'Dibatalkan',
            jenis_perubahan = 'Batal'
        WHERE krs_id = ?
          AND kelas_id = ?
          AND status = 'Terdaftar'
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		krsID,
		kelasID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("kelas tidak ditemukan atau sudah dibatalkan")
	}

	return nil
}

func (r *Repository) HitungTotalSKS(
	ctx context.Context,
	krsID int,
) (int, error) {

	query := `
        SELECT COALESCE(SUM(mk.sks), 0)
        FROM krs_detail kd
        JOIN kelas k ON kd.kelas_id = k.id
        JOIN mata_kuliah mk ON k.kode_mk = mk.kode_mk
        WHERE kd.krs_id = ?
          AND kd.status = 'Terdaftar'
    `

	var totalSKS int

	err := r.db.QueryRowContext(
		ctx,
		query,
		krsID,
	).Scan(&totalSKS)

	if err != nil {
		return 0, err
	}

	return totalSKS, nil
}

func (r *Repository) UpdateTotalSKS(
	ctx context.Context,
	krsID int,
	totalSKS int,
) error {

	query := `
        UPDATE krs
        SET total_sks = ?
        WHERE id = ?
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		totalSKS,
		krsID,
	)

	return err
}

func (r *Repository) TambahTerisi(
	ctx context.Context,
	kelasID int,
) error {

	query := `
        UPDATE kelas
        SET terisi = terisi + 1
        WHERE id = ? AND terisi < kuota
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		kelasID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("kuota kelas sudah penuh")
	}

	return nil
}

func (r *Repository) KurangiTerisi(
	ctx context.Context,
	kelasID int,
) error {

	query := `
        UPDATE kelas
        SET terisi = GREATEST(terisi - 1, 0)
        WHERE id = ?
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		kelasID,
	)

	return err
}

func (r *Repository) CekKelasTerdaftar(
	ctx context.Context,
	krsID int,
	kelasID int,
) (bool, error) {

	query := `
        SELECT EXISTS (
            SELECT 1
            FROM krs_detail
            WHERE krs_id = ?
            AND kelas_id = ?
            AND status = 'Terdaftar'
        )
    `

	var terdaftar bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		krsID,
		kelasID,
	).Scan(&terdaftar)

	return terdaftar, err
}
