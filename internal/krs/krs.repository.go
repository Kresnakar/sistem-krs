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

type PrasyaratNode struct {
	ID           int
	ParentID     *int
	Tipe         string
	KodeMKSyarat string
	NilaiMinimal string
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

func (r *Repository) GetNIMByKRS(
	ctx context.Context,
	krsID int,
) (string, error) {
	var nim string

	query := `SELECT nim FROM krs WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, krsID).Scan(&nim)
	return nim, err
}

func (r *Repository) GetKodeMKByKelas(
	ctx context.Context,
	kelasID int,
) (string, error) {
	var kodeMK string

	query := `SELECT kode_mk FROM kelas WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, kelasID).Scan(&kodeMK)
	return kodeMK, err
}

func (r *Repository) GetPrasyaratNodes(
	ctx context.Context,
	kodeMK string,
) ([]PrasyaratNode, error) {

	query := `
		SELECT id, parent_id, tipe,
		       kode_mk_syarat, nilai_minimal
		FROM prasyarat_node
		WHERE kode_mk = ?
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, kodeMK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []PrasyaratNode

	for rows.Next() {
		var node PrasyaratNode
		var parentID sql.NullInt64
		var kodeSyarat sql.NullString
		var nilaiMinimal sql.NullString

		err := rows.Scan(
			&node.ID,
			&parentID,
			&node.Tipe,
			&kodeSyarat,
			&nilaiMinimal,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			id := int(parentID.Int64)
			node.ParentID = &id
		}

		if kodeSyarat.Valid {
			node.KodeMKSyarat = kodeSyarat.String
		}

		if nilaiMinimal.Valid {
			node.NilaiMinimal = nilaiMinimal.String
		}

		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

func (r *Repository) CekPrasyaratLulus(
	ctx context.Context,
	nim string,
	kodeMK string,
	nilaiMinimal string,
) (bool, error) {

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM riwayat_nilai rn
			JOIN skala_nilai nilai
				ON nilai.nilai_huruf = rn.nilai_huruf
			JOIN skala_nilai minimal
				ON minimal.nilai_huruf = ?
			WHERE rn.nim = ?
			  AND rn.kode_mk = ?
			  AND rn.lulus = 1
			  AND nilai.bobot >= minimal.bobot
		)
	`

	var lulus bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		nilaiMinimal,
		nim,
		kodeMK,
	).Scan(&lulus)

	return lulus, err
}

func (r *Repository) CekBentrokJadwal(
	ctx context.Context,
	krsID int,
	kelasID int,
) (bool, error) {

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM jadwal_kelas jadwal_baru
			JOIN jadwal_kelas jadwal_lama
				ON jadwal_baru.hari = jadwal_lama.hari
			   AND jadwal_baru.jam_mulai < jadwal_lama.jam_selesai
			   AND jadwal_baru.jam_selesai > jadwal_lama.jam_mulai
			JOIN krs_detail kd
				ON kd.kelas_id = jadwal_lama.kelas_id
			WHERE jadwal_baru.kelas_id = ?
			  AND kd.krs_id = ?
			  AND kd.status = 'Terdaftar'
			  AND kd.kelas_id <> ?
		)
	`

	var bentrok bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		kelasID,
		krsID,
		kelasID,
	).Scan(&bentrok)

	return bentrok, err
}
