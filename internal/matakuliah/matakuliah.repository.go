package matakuliah

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("mata kuliah tidak ditemukan")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]MataKuliah, error) {
	query := `SELECT kode_mk, nama, sks, kurikulum, aktif FROM mata_kuliah ORDER BY kode_mk ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []MataKuliah{}
	for rows.Next() {
		var mk MataKuliah
		if err := rows.Scan(&mk.KodeMK, &mk.Nama, &mk.SKS, &mk.Kurikulum, &mk.Aktif); err != nil {
			return nil, err
		}
		list = append(list, mk)
	}
	return list, rows.Err()
}

func (r *Repository) GetByKode(ctx context.Context, kodeMK string) (*MataKuliah, error) {
	query := `SELECT kode_mk, nama, sks, kurikulum, aktif FROM mata_kuliah WHERE kode_mk = ?`
	var mk MataKuliah
	err := r.db.QueryRowContext(ctx, query, kodeMK).Scan(
		&mk.KodeMK, &mk.Nama, &mk.SKS, &mk.Kurikulum, &mk.Aktif,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &mk, nil
}

func (r *Repository) Create(ctx context.Context, mk MataKuliah) error {
	query := `INSERT INTO mata_kuliah (kode_mk, nama, sks, kurikulum, aktif) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, mk.KodeMK, mk.Nama, mk.SKS, mk.Kurikulum, mk.Aktif)
	return err
}

func (r *Repository) Update(ctx context.Context, kodeMK string, mk MataKuliah) error {
	query := `UPDATE mata_kuliah SET nama = ?, sks = ?, kurikulum = ?, aktif = ? WHERE kode_mk = ?`
	res, err := r.db.ExecContext(ctx, query, mk.Nama, mk.SKS, mk.Kurikulum, mk.Aktif, kodeMK)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, kodeMK string) error {
	query := `DELETE FROM mata_kuliah WHERE kode_mk = ?`
	res, err := r.db.ExecContext(ctx, query, kodeMK)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}