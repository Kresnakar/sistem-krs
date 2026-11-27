package krs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// GetKRSByID mengambil KRS berdasarkan ID.
func (r *Repository) GetKRSByID(ctx context.Context, krsID int) (*KRS, error) {
	query := `
		SELECT id, nim, semester_id, status, dosen_wali_id,
		       ips_acuan, batas_sks, total_sks, diajukan_pada,
		       disetujui_pada, catatan_penolakan
		FROM krs
		WHERE id = ?`

	var data KRS
	var dosenWaliID sql.NullInt64
	var diajukanPada sql.NullTime
	var disetujuiPada sql.NullTime
	var catatan sql.NullString

	err := r.db.QueryRowContext(ctx, query, krsID).Scan(
		&data.ID, &data.NIM, &data.SemesterID, &data.Status,
		&dosenWaliID, &data.IPSAcuan, &data.BatasSKS, &data.TotalSKS,
		&diajukanPada, &disetujuiPada, &catatan,
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

// GetOrCreateDraft mengambil KRS yang ada atau membuat draf baru.
func (r *Repository) GetOrCreateDraft(ctx context.Context, nim string, semesterID int) (*KRS, error) {
	existing, err := r.GetByNIMSemester(ctx, nim, semesterID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrKRSNotFound) {
		return nil, err
	}

	var dosenWaliID sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `
		SELECT dosen_wali_id FROM mahasiswa WHERE nim = ?`, nim,
	).Scan(&dosenWaliID); err != nil {
		return nil, fmt.Errorf("mahasiswa tidak ditemukan: %w", err)
	}

	var dosenArg any
	if dosenWaliID.Valid {
		dosenArg = dosenWaliID.Int64
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO krs (nim, semester_id, status, dosen_wali_id, ips_acuan, batas_sks, total_sks)
		VALUES (?, ?, 'Draft', ?, 0, 21, 0)`, nim, semesterID, dosenArg)
	if err != nil {
		// Bisa terjadi race condition; ambil data yang sudah dibuat.
		if again, e := r.GetByNIMSemester(ctx, nim, semesterID); e == nil {
			return again, nil
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetKRSByID(ctx, int(id))
}

// AjukanKRS mengubah status KRS menjadi Diajukan.
func (r *Repository) AjukanKRS(ctx context.Context, krsID int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE krs
		SET status = 'Diajukan', diajukan_pada = NOW(), catatan_penolakan = NULL
		WHERE id = ? AND status IN ('Draft', 'Ditolak')`, krsID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrKRSBukanDraft
	}
	return nil
}

// ListKRSByDosen mengambil KRS yang menjadi bimbingan seorang dosen wali.
func (r *Repository) ListKRSByDosen(ctx context.Context, dosenWaliID int) ([]KRSRingkas, error) {
	query := `
		SELECT k.id, k.nim, m.nama, k.semester_id, k.status,
		       k.total_sks, k.batas_sks, k.diajukan_pada,
		       k.disetujui_pada, k.catatan_penolakan
		FROM krs k
		JOIN mahasiswa m ON m.nim = k.nim
		WHERE k.dosen_wali_id = ? OR m.dosen_wali_id = ?
		ORDER BY k.diajukan_pada DESC, k.id DESC`

	rows, err := r.db.QueryContext(ctx, query, dosenWaliID, dosenWaliID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []KRSRingkas{}
	for rows.Next() {
		var item KRSRingkas
		var diajukan sql.NullTime
		var disetujui sql.NullTime
		var catatan sql.NullString

		if err := rows.Scan(
			&item.ID, &item.NIM, &item.NamaMhs, &item.SemesterID,
			&item.Status, &item.TotalSKS, &item.BatasSKS,
			&diajukan, &disetujui, &catatan,
		); err != nil {
			return nil, err
		}

		if diajukan.Valid {
			item.DiajukanPada = &diajukan.Time
		}
		if disetujui.Valid {
			item.DisetujuiPada = &disetujui.Time
		}
		if catatan.Valid {
			item.CatatanPenolakan = &catatan.String
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// GetKRSForDosen memastikan KRS memang bimbingan dosen tersebut.
func (r *Repository) GetKRSForDosen(ctx context.Context, krsID, dosenWaliID int) (*KRS, error) {
	data, err := r.GetKRSByID(ctx, krsID)
	if err != nil {
		return nil, err
	}

	var allowed bool
	err = r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM krs k
			JOIN mahasiswa m ON m.nim = k.nim
			WHERE k.id = ? AND (k.dosen_wali_id = ? OR m.dosen_wali_id = ?)
		)`, krsID, dosenWaliID, dosenWaliID).Scan(&allowed)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrKRSNotFound
	}
	return data, nil
}

// GetDetailLengkap mengambil detail KRS beserta info mata kuliah dan kelas.
func (r *Repository) GetDetailLengkap(ctx context.Context, krsID int) ([]KRSDetailLengkap, error) {
	query := `
		SELECT kd.id, kd.krs_id, kd.kelas_id,
		       mk.kode_mk, mk.nama, mk.sks, k.kode_paralel,
		       kd.status, kd.jenis_perubahan, kd.status_persetujuan
		FROM krs_detail kd
		JOIN kelas k ON k.id = kd.kelas_id
		JOIN mata_kuliah mk ON mk.kode_mk = k.kode_mk
		WHERE kd.krs_id = ?
		ORDER BY kd.id ASC`

	rows, err := r.db.QueryContext(ctx, query, krsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []KRSDetailLengkap{}
	for rows.Next() {
		var d KRSDetailLengkap
		if err := rows.Scan(
			&d.ID, &d.KRSID, &d.KelasID,
			&d.KodeMK, &d.NamaMK, &d.SKS, &d.KodeParalel,
			&d.Status, &d.JenisPerubahan, &d.StatusPersetujuan,
		); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (r *Repository) SetujuiKRS(ctx context.Context, krsID, dosenWaliID int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE krs k
		JOIN mahasiswa m ON m.nim = k.nim
		SET k.status = 'Disetujui',
		    k.disetujui_pada = NOW(),
		    k.catatan_penolakan = NULL
		WHERE k.id = ? AND k.status = 'Diajukan'
		  AND (k.dosen_wali_id = ? OR m.dosen_wali_id = ?)`,
		krsID, dosenWaliID, dosenWaliID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("KRS tidak ditemukan, bukan bimbingan Anda, atau status bukan Diajukan")
	}
	return nil
}

func (r *Repository) TolakKRS(ctx context.Context, krsID, dosenWaliID int, catatan string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE krs k
		JOIN mahasiswa m ON m.nim = k.nim
		SET k.status = 'Ditolak',
		    k.catatan_penolakan = ?,
		    k.disetujui_pada = NULL
		WHERE k.id = ? AND k.status = 'Diajukan'
		  AND (k.dosen_wali_id = ? OR m.dosen_wali_id = ?)`,
		catatan, krsID, dosenWaliID, dosenWaliID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("KRS tidak ditemukan, bukan bimbingan Anda, atau status bukan Diajukan")
	}
	return nil
}
