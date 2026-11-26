package prasyarat

import (
	"context"
	"database/sql"
)

type NilaiMahasiswa struct {
	KodeMK     string
	NilaiHuruf string
	Bobot      float64
	Lulus      bool
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAllNodes(ctx context.Context) ([]PrasyaratNode, error) {
	query := `SELECT id, kode_mk, parent_id, tipe, kode_mk_syarat, nilai_minimal FROM prasyarat_node ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNodes(rows)
}

func (r *Repository) GetNodesByKodeMK(ctx context.Context, kodeMK string) ([]PrasyaratNode, error) {
	query := `SELECT id, kode_mk, parent_id, tipe, kode_mk_syarat, nilai_minimal FROM prasyarat_node WHERE kode_mk = ? ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query, kodeMK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNodes(rows)
}

func scanNodes(rows *sql.Rows) ([]PrasyaratNode, error) {
	list := []PrasyaratNode{}
	for rows.Next() {
		var n PrasyaratNode
		var parentID sql.NullInt64
		var kodeSyarat, nilaiMin sql.NullString

		if err := rows.Scan(&n.ID, &n.KodeMK, &parentID, &n.Tipe, &kodeSyarat, &nilaiMin); err != nil {
			return nil, err
		}
		if parentID.Valid {
			val := int(parentID.Int64)
			n.ParentID = &val
		}
		if kodeSyarat.Valid {
			val := kodeSyarat.String
			n.KodeMKSyarat = &val
		}
		if nilaiMin.Valid {
			val := nilaiMin.String
			n.NilaiMinimal = &val
		}
		list = append(list, n)
	}
	return list, rows.Err()
}

func (r *Repository) CreateNode(ctx context.Context, n PrasyaratNode) (int, error) {
	query := `INSERT INTO prasyarat_node (kode_mk, parent_id, tipe, kode_mk_syarat, nilai_minimal) VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, n.KodeMK, n.ParentID, n.Tipe, n.KodeMKSyarat, n.NilaiMinimal)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *Repository) DeleteByKodeMK(ctx context.Context, kodeMK string) error {
	query := `DELETE FROM prasyarat_node WHERE kode_mk = ?`
	_, err := r.db.ExecContext(ctx, query, kodeMK)
	return err
}

func (r *Repository) GetSkalaNilai(ctx context.Context) (map[string]float64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT nilai_huruf, bobot FROM skala_nilai`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skala := make(map[string]float64)
	for rows.Next() {
		var huruf string
		var bobot float64
		if err := rows.Scan(&huruf, &bobot); err != nil {
			return nil, err
		}
		skala[huruf] = bobot
	}
	return skala, rows.Err()
}

func (r *Repository) GetRiwayatNilai(ctx context.Context, nim string) (map[string]NilaiMahasiswa, error) {
	query := `SELECT kode_mk, nilai_huruf, bobot, lulus FROM riwayat_nilai WHERE nim = ?`
	rows, err := r.db.QueryContext(ctx, query, nim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasil := make(map[string]NilaiMahasiswa)
	for rows.Next() {
		var nm NilaiMahasiswa
		if err := rows.Scan(&nm.KodeMK, &nm.NilaiHuruf, &nm.Bobot, &nm.Lulus); err != nil {
			return nil, err
		}
		// Simpan nilai tertinggi jika mahasiswa pernah mengulang mata kuliah
		if existing, ok := hasil[nm.KodeMK]; !ok || nm.Bobot > existing.Bobot {
			hasil[nm.KodeMK] = nm
		}
	}
	return hasil, rows.Err()
}