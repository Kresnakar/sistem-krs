package krs

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrSemesterNotFound = errors.New("semester tidak ditemukan")
	ErrPeriodeNotFound  = errors.New("periode KRS tidak ditemukan")
)

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *Repository) ListSemester(ctx context.Context) ([]Semester, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tahun_akademik, jenis, status_aktif
		FROM semester
		ORDER BY tahun_akademik DESC, jenis ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Semester{}
	for rows.Next() {
		var s Semester
		if err := rows.Scan(&s.ID, &s.TahunAkademik, &s.Jenis, &s.StatusAktif); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *Repository) GetSemester(ctx context.Context, id int) (*Semester, error) {
	var s Semester
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tahun_akademik, jenis, status_aktif
		FROM semester
		WHERE id = ?`, id).Scan(&s.ID, &s.TahunAkademik, &s.Jenis, &s.StatusAktif)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSemesterNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) GetActiveSemester(ctx context.Context) (*Semester, error) {
	var s Semester
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tahun_akademik, jenis, status_aktif
		FROM semester
		WHERE status_aktif = 1
		ORDER BY id DESC
		LIMIT 1`).Scan(&s.ID, &s.TahunAkademik, &s.Jenis, &s.StatusAktif)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSemesterNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) CreateSemester(ctx context.Context, s Semester) (int, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO semester (tahun_akademik, jenis, status_aktif)
		VALUES (?, ?, ?)`, s.TahunAkademik, s.Jenis, s.StatusAktif)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *Repository) UpdateSemester(ctx context.Context, id int, s Semester) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE semester
		SET tahun_akademik = ?, jenis = ?, status_aktif = ?
		WHERE id = ?`, s.TahunAkademik, s.Jenis, s.StatusAktif, id)
	return err
}

// AktifkanSemester memastikan hanya satu semester yang aktif.
func (r *Repository) AktifkanSemester(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE semester SET status_aktif = 0 WHERE id <> ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE semester SET status_aktif = 1 WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

const periodeColumns = `id, semester_id, mulai_krs, selesai_krs, mulai_perubahan, selesai_perubahan, tanggal_mulai_kuliah`

func scanPeriode(scanner rowScanner) (*PeriodeKRS, error) {
	var p PeriodeKRS
	err := scanner.Scan(
		&p.ID, &p.SemesterID,
		&p.MulaiKRS, &p.SelesaiKRS,
		&p.MulaiPerubahan, &p.SelesaiPerubahan,
		&p.TanggalMulaiKuliah,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPeriodeNotFound
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()
	p.KRSDibuka = !now.Before(p.MulaiKRS) && !now.After(p.SelesaiKRS)
	p.PerubahanDibuka = !now.Before(p.MulaiPerubahan) && !now.After(p.SelesaiPerubahan)
	return &p, nil
}

func (r *Repository) ListPeriode(ctx context.Context) ([]PeriodeKRS, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+periodeColumns+`
		FROM periode_krs
		ORDER BY semester_id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []PeriodeKRS{}
	for rows.Next() {
		p, err := scanPeriode(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *p)
	}
	return list, rows.Err()
}

func (r *Repository) GetPeriode(ctx context.Context, id int) (*PeriodeKRS, error) {
	return scanPeriode(r.db.QueryRowContext(ctx, `
		SELECT `+periodeColumns+`
		FROM periode_krs
		WHERE id = ?`, id))
}

func (r *Repository) GetPeriodeBySemester(ctx context.Context, semesterID int) (*PeriodeKRS, error) {
	return scanPeriode(r.db.QueryRowContext(ctx, `
		SELECT `+periodeColumns+`
		FROM periode_krs
		WHERE semester_id = ?`, semesterID))
}

func (r *Repository) CreatePeriode(ctx context.Context, p PeriodeKRS) (int, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO periode_krs
			(semester_id, mulai_krs, selesai_krs, mulai_perubahan, selesai_perubahan, tanggal_mulai_kuliah)
		VALUES (?, ?, ?, ?, ?, ?)`,
		p.SemesterID, p.MulaiKRS, p.SelesaiKRS,
		p.MulaiPerubahan, p.SelesaiPerubahan, p.TanggalMulaiKuliah,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *Repository) UpdatePeriode(ctx context.Context, id int, p PeriodeKRS) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE periode_krs
		SET semester_id = ?, mulai_krs = ?, selesai_krs = ?,
		    mulai_perubahan = ?, selesai_perubahan = ?, tanggal_mulai_kuliah = ?
		WHERE id = ?`,
		p.SemesterID, p.MulaiKRS, p.SelesaiKRS,
		p.MulaiPerubahan, p.SelesaiPerubahan, p.TanggalMulaiKuliah, id,
	)
	return err
}

// UpdateJendelaPeriode membuka/mengatur ulang jendela waktu tanpa mengubah skema.
func (r *Repository) UpdateJendelaPeriode(ctx context.Context, id int, jenis string, mulai, selesai time.Time) error {
	var query string
	switch jenis {
	case "krs":
		query = `UPDATE periode_krs SET mulai_krs = ?, selesai_krs = ? WHERE id = ?`
	case "perubahan":
		query = `UPDATE periode_krs SET mulai_perubahan = ?, selesai_perubahan = ? WHERE id = ?`
	default:
		return errors.New("jenis periode harus 'krs' atau 'perubahan'")
	}

	_, err := r.db.ExecContext(ctx, query, mulai, selesai, id)
	return err
}

// TutupJendelaPeriode menutup jendela sekarang juga dengan menyetel selesai = waktu sekarang.
func (r *Repository) TutupJendelaPeriode(ctx context.Context, id int, jenis string) error {
	now := time.Now()

	var query string
	switch jenis {
	case "krs":
		query = `UPDATE periode_krs
			SET mulai_krs = IF(mulai_krs > ?, ?, mulai_krs), selesai_krs = ?
			WHERE id = ?`
	case "perubahan":
		query = `UPDATE periode_krs
			SET mulai_perubahan = IF(mulai_perubahan > ?, ?, mulai_perubahan), selesai_perubahan = ?
			WHERE id = ?`
	default:
		return errors.New("jenis periode harus 'krs' atau 'perubahan'")
	}

	_, err := r.db.ExecContext(ctx, query, now, now, now, id)
	return err
}

func (r *Repository) IsKRSWindowOpen(ctx context.Context, semesterID int) (bool, error) {
	var open bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM periode_krs
			WHERE semester_id = ? AND ? BETWEEN mulai_krs AND selesai_krs
		)`, semesterID, time.Now()).Scan(&open)
	return open, err
}

func (r *Repository) IsPerubahanWindowOpen(ctx context.Context, semesterID int) (bool, error) {
	var open bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM periode_krs
			WHERE semester_id = ? AND ? BETWEEN mulai_perubahan AND selesai_perubahan
		)`, semesterID, time.Now()).Scan(&open)
	return open, err
}

func (r *Repository) GetSemesterIDByKRS(ctx context.Context, krsID int) (int, error) {
	var semesterID int
	err := r.db.QueryRowContext(ctx, `SELECT semester_id FROM krs WHERE id = ?`, krsID).Scan(&semesterID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrKRSNotFound
	}
	if err != nil {
		return 0, err
	}
	return semesterID, nil
}

func (r *Repository) ListKelasBySemester(ctx context.Context, semesterID int) ([]KelasInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT k.id, k.kode_mk, mk.nama, k.semester_id, k.kode_paralel,
		       COALESCE(k.dosen_pengampu, ''), k.kuota, k.terisi, k.status
		FROM kelas k
		JOIN mata_kuliah mk ON mk.kode_mk = k.kode_mk
		WHERE k.semester_id = ? AND k.status <> 'Dibatalkan'
		ORDER BY k.kode_mk ASC, k.kode_paralel ASC`, semesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []KelasInfo{}
	for rows.Next() {
		var k KelasInfo
		if err := rows.Scan(
			&k.ID, &k.KodeMK, &k.NamaMK, &k.SemesterID, &k.KodeParalel,
			&k.DosenPengampu, &k.Kuota, &k.Terisi, &k.Status,
		); err != nil {
			return nil, err
		}
		list = append(list, k)
	}
	return list, rows.Err()
}
