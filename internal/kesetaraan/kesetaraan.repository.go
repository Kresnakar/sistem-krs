package kesetaraan

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound   = errors.New("data kesetaraan mata kuliah tidak ditemukan")
	ErrDuplicate  = errors.New("data kesetaraan mata kuliah sudah terdaftar")
	ErrMKNotFound = errors.New("kode mk_lama atau mk_baru tidak ditemukan di tabel mata_kuliah")
	ErrSelfEqual  = errors.New("mk_lama dan mk_baru tidak boleh sama")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]KesetaraanMK, error) {
	query := `SELECT mk_lama, mk_baru, COALESCE(keterangan, '') FROM kesetaraan_mk ORDER BY mk_lama ASC, mk_baru ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []KesetaraanMK{}
	for rows.Next() {
		var item KesetaraanMK
		if err := rows.Scan(&item.MKLama, &item.MKBaru, &item.Keterangan); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *Repository) Create(ctx context.Context, item KesetaraanMK) error {
	if strings.EqualFold(item.MKLama, item.MKBaru) {
		return ErrSelfEqual
	}

	query := `INSERT INTO kesetaraan_mk (mk_lama, mk_baru, keterangan) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, item.MKLama, item.MKBaru, item.Keterangan)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			switch mysqlErr.Number {
			case 1062:
				return ErrDuplicate
			case 1452:
				return ErrMKNotFound
			}
		}
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, mkLama, mkBaru string) error {
	query := `DELETE FROM kesetaraan_mk WHERE mk_lama = ? AND mk_baru = ?`
	res, err := r.db.ExecContext(ctx, query, mkLama, mkBaru)
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

// GetSetaraCodes mencari seluruh kode mata kuliah yang setara dengan kodeMK (dua arah).
// Contoh: jika IF001 setara IF101, maka mencari IF101 akan mengembalikan ["IF101", "IF001"].
func (r *Repository) GetSetaraCodes(ctx context.Context, kodeMK string) ([]string, error) {
	semua, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	adj := make(map[string][]string)
	for _, k := range semua {
		adj[k.MKLama] = append(adj[k.MKLama], k.MKBaru)
		adj[k.MKBaru] = append(adj[k.MKBaru], k.MKLama)
	}

	visited := map[string]bool{kodeMK: true}
	queue := []string{kodeMK}
	result := []string{kodeMK}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, next := range adj[curr] {
			if !visited[next] {
				visited[next] = true
				result = append(result, next)
				queue = append(queue, next)
			}
		}
	}

	return result, nil
}