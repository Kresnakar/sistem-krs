package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kresnakar/sistem-krs/internal/database"
)

func main() {
	// Membaca konfigurasi dari file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan")
	}

	// Memberikan batas waktu koneksi 5 detik
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// Menghubungkan aplikasi dengan MySQL
	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Menampilkan status koneksi
	log.Println("================================")
	log.Println("SISTEM KRS")
	log.Println("Koneksi database berhasil!")
	log.Println("Database: sistem_krs")
	log.Println("================================")
}
