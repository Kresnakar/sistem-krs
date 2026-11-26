package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func Connect(ctx context.Context) (*sql.DB, error) {
	cfg := mysql.NewConfig()

	cfg.User = os.Getenv("DB_USER")
	cfg.Passwd = os.Getenv("DB_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)
	cfg.DBName = os.Getenv("DB_NAME")
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Params = map[string]string{
		"charset": "utf8mb4",
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf(
			"gagal membuka koneksi database: %w", err,
		)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf(
			"gagal menghubungi MySQL: %w", err,
		)
	}

	return db, nil
}
