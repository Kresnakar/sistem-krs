package krs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrPermintaanNotFound = errors.New("permintaan perubahan tidak ditemukan")
	ErrPermintaanDiproses = errors.New("permintaan perubahan sudah diproses")
)

func (r *Repository) CreatePermintaanPerubahan(
	ctx context.Context,
	krsID int,
	jenis string,
	alasan string,
	kelasIDs []int,
) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO permintaan_perubahan (krs_id, jenis_perubahan, alasan, status)
		VALUES (?, ?, ?, 'Menunggu')`, krsID, jenis, alasan)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, kelasID := range kelasIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO permintaan_perubahan_detail (permintaan_id, kelas_id)
			VALUES (?, ?)`, id, kelasID); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(id), nil
}

const permintaanColumns = `
	p.id, p.krs_id, k.nim, m.nama, p.jenis_perubahan,
	COALESCE(p.alasan, ''), p.status, p.diajukan_pada,
	p.diproses_oleh, p.diproses_pada, p.catatan`

func scanPermintaanRows(rows *sql.Rows) ([]PermintaanPerubahan, error) {
	list := []PermintaanPerubahan{}
	for rows.Next() {
		var p PermintaanPerubahan
		var diprosesOleh sql.NullInt64
		var diprosesPada sql.NullTime
		var catatan sql.NullString

		if err := rows.Scan(
			&p.ID, &p.KRSID, &p.NIM, &p.NamaMhs,
			&p.JenisPerubahan, &p.Alasan, &p.Status, &p.DiajukanPada,
			&diprosesOleh, &diprosesPada, &catatan,
		); err != nil {
			return nil, err
		}

		if diprosesOleh.Valid {
			v := int(diprosesOleh.Int64)
			p.DiprosesOleh = &v
		}
		if diprosesPada.Valid {
			p.DiprosesPada = &diprosesPada.Time
		}
		if catatan.Valid {
			p.Catatan = &catatan.String
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *Repository) attachPermintaanDetails(ctx context.Context, list []PermintaanPerubahan) ([]PermintaanPerubahan, error) {
	for i := range list {
		details, err := r.GetPermintaanDetail(ctx, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Detail = details
	}
	return list, nil
}

func (r *Repository) ListPermintaanByNIM(ctx context.Context, nim string) ([]PermintaanPerubahan, error) {
	query := `
		SELECT ` + permintaanColumns + `
		FROM permintaan_perubahan p
		JOIN krs k ON k.id = p.krs_id
		JOIN mahasiswa m ON m.nim = k.nim
		WHERE k.nim = ?
		ORDER BY p.diajukan_pada DESC, p.id DESC`

	rows, err := r.db.QueryContext(ctx, query, nim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list, err := scanPermintaanRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return r.attachPermintaanDetails(ctx, list)
}

func (r *Repository) ListPermintaanByKRS(ctx context.Context, krsID int) ([]PermintaanPerubahan, error) {
	query := `
		SELECT ` + permintaanColumns + `
		FROM permintaan_perubahan p
		JOIN krs k ON k.id = p.krs_id
		JOIN mahasiswa m ON m.nim = k.nim
		WHERE p.krs_id = ?
		ORDER BY p.diajukan_pada DESC, p.id DESC`

	rows, err := r.db.QueryContext(ctx, query, krsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list, err := scanPermintaanRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return r.attachPermintaanDetails(ctx, list)
}

func (r *Repository) ListPermintaanByDosen(ctx context.Context, dosenWaliID int) ([]PermintaanPerubahan, error) {
	query := `
		SELECT ` + permintaanColumns + `
		FROM permintaan_perubahan p
		JOIN krs k ON k.id = p.krs_id
		JOIN mahasiswa m ON m.nim = k.nim
		WHERE k.dosen_wali_id = ? OR m.dosen_wali_id = ?
		ORDER BY p.diajukan_pada DESC, p.id DESC`

	rows, err := r.db.QueryContext(ctx, query, dosenWaliID, dosenWaliID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list, err := scanPermintaanRows(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return r.attachPermintaanDetails(ctx, list)
}

func (r *Repository) GetPermintaan(ctx context.Context, id int) (*PermintaanPerubahan, error) {
	query := `
		SELECT ` + permintaanColumns + `
		FROM permintaan_perubahan p
		JOIN krs k ON k.id = p.krs_id
		JOIN mahasiswa m ON m.nim = k.nim
		WHERE p.id = ?`

	var p PermintaanPerubahan
	var diprosesOleh sql.NullInt64
	var diprosesPada sql.NullTime
	var catatan sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.KRSID, &p.NIM, &p.NamaMhs,
		&p.JenisPerubahan, &p.Alasan, &p.Status, &p.DiajukanPada,
		&diprosesOleh, &diprosesPada, &catatan,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPermintaanNotFound
	}
	if err != nil {
		return nil, err
	}

	if diprosesOleh.Valid {
		v := int(diprosesOleh.Int64)
		p.DiprosesOleh = &v
	}
	if diprosesPada.Valid {
		p.DiprosesPada = &diprosesPada.Time
	}
	if catatan.Valid {
		p.Catatan = &catatan.String
	}

	details, err := r.GetPermintaanDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Detail = details
	return &p, nil
}

func (r *Repository) GetPermintaanDetail(ctx context.Context, permintaanID int) ([]PermintaanPerubahanDetail, error) {
	query := `
		SELECT pd.id, pd.permintaan_id, pd.kelas_id,
		       mk.kode_mk, mk.nama, k.kode_paralel
		FROM permintaan_perubahan_detail pd
		JOIN kelas k ON k.id = pd.kelas_id
		JOIN mata_kuliah mk ON mk.kode_mk = k.kode_mk
		WHERE pd.permintaan_id = ?
		ORDER BY pd.id ASC`

	rows, err := r.db.QueryContext(ctx, query, permintaanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []PermintaanPerubahanDetail{}
	for rows.Next() {
		var d PermintaanPerubahanDetail
		if err := rows.Scan(
			&d.ID, &d.PermintaanID, &d.KelasID,
			&d.KodeMK, &d.NamaMK, &d.KodeParalel,
		); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func hitungTotalSKSTx(ctx context.Context, tx *sql.Tx, krsID int) (int, error) {
	var total int
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(mk.sks), 0)
		FROM krs_detail kd
		JOIN kelas k ON k.id = kd.kelas_id
		JOIN mata_kuliah mk ON mk.kode_mk = k.kode_mk
		WHERE kd.krs_id = ? AND kd.status = 'Terdaftar'`, krsID).Scan(&total)
	return total, err
}

// SetujuiPermintaan menerapkan perubahan ke krs_detail lalu menandai permintaan disetujui.
func (r *Repository) SetujuiPermintaan(ctx context.Context, permintaanID, dosenWaliID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var krsID int
	var jenis, status string
	err = tx.QueryRowContext(ctx, `
		SELECT p.krs_id, p.jenis_perubahan, p.status
		FROM permintaan_perubahan p
		JOIN krs k ON k.id = p.krs_id
		JOIN mahasiswa m ON m.nim = k.nim
		WHERE p.id = ? AND (k.dosen_wali_id = ? OR m.dosen_wali_id = ?)
		FOR UPDATE`, permintaanID, dosenWaliID, dosenWaliID,
	).Scan(&krsID, &jenis, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrPermintaanNotFound
	}
	if err != nil {
		return err
	}
	if status != "Menunggu" {
		return ErrPermintaanDiproses
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT kelas_id FROM permintaan_perubahan_detail WHERE permintaan_id = ?`, permintaanID)
	if err != nil {
		return err
	}

	kelasIDs := []int{}
	for rows.Next() {
		var kelasID int
		if err := rows.Scan(&kelasID); err != nil {
			rows.Close()
			return err
		}
		kelasIDs = append(kelasIDs, kelasID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(kelasIDs) == 0 {
		return errors.New("permintaan perubahan tidak memiliki detail kelas")
	}

	for _, kelasID := range kelasIDs {
		var current sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT status FROM krs_detail
			WHERE krs_id = ? AND kelas_id = ?
			FOR UPDATE`, krsID, kelasID).Scan(&current)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		terdaftar := current.Valid && current.String == "Terdaftar"

		switch jenis {
		case "Tambah Kelas":
			if terdaftar {
				continue
			}

			var kuota, terisi int
			if err := tx.QueryRowContext(ctx, `
				SELECT kuota, terisi FROM kelas WHERE id = ? FOR UPDATE`, kelasID,
			).Scan(&kuota, &terisi); err != nil {
				return err
			}
			if terisi >= kuota {
				return fmt.Errorf("kuota kelas %d sudah penuh", kelasID)
			}

			if _, err := tx.ExecContext(ctx, `
				INSERT INTO krs_detail (krs_id, kelas_id, status, jenis_perubahan, status_persetujuan)
				VALUES (?, ?, 'Terdaftar', 'Tambah', 'Final')
				ON DUPLICATE KEY UPDATE
					status = 'Terdaftar',
					jenis_perubahan = 'Tambah',
					status_persetujuan = 'Final'`, krsID, kelasID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE kelas SET terisi = terisi + 1 WHERE id = ?`, kelasID); err != nil {
				return err
			}

		case "Batal Kelas":
			if !terdaftar {
				return fmt.Errorf("kelas %d tidak terdaftar di KRS", kelasID)
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE krs_detail
				SET status = 'Dibatalkan', jenis_perubahan = 'Batal'
				WHERE krs_id = ? AND kelas_id = ?`, krsID, kelasID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE kelas SET terisi = GREATEST(terisi - 1, 0) WHERE id = ?`, kelasID); err != nil {
				return err
			}

		default:
			return fmt.Errorf("jenis perubahan tidak dikenali: %s", jenis)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE permintaan_perubahan
		SET status = 'Disetujui', diproses_oleh = ?, diproses_pada = NOW()
		WHERE id = ?`, dosenWaliID, permintaanID); err != nil {
		return err
	}

	total, err := hitungTotalSKSTx(ctx, tx, krsID)
	if err != nil {
		return err
	}

	var batasSKS int
	if err := tx.QueryRowContext(ctx, `SELECT batas_sks FROM krs WHERE id = ?`, krsID).Scan(&batasSKS); err != nil {
		return err
	}
	if batasSKS > 0 && total > batasSKS {
		return fmt.Errorf("total SKS %d melebihi batas %d", total, batasSKS)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE krs SET total_sks = ? WHERE id = ?`, total, krsID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) TolakPermintaan(ctx context.Context, permintaanID, dosenWaliID int, catatan string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE permintaan_perubahan p
		JOIN krs k ON k.id = p.krs_id
		JOIN mahasiswa m ON m.nim = k.nim
		SET p.status = 'Ditolak',
		    p.catatan = ?,
		    p.diproses_oleh = ?,
		    p.diproses_pada = NOW()
		WHERE p.id = ? AND p.status = 'Menunggu'
		  AND (k.dosen_wali_id = ? OR m.dosen_wali_id = ?)`,
		catatan, dosenWaliID, permintaanID, dosenWaliID, dosenWaliID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("permintaan tidak ditemukan, bukan bimbingan Anda, atau sudah diproses")
	}
	return nil
}
