package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/kresnakar/sistem-krs/internal/database"
	"github.com/kresnakar/sistem-krs/internal/kelas"
	"github.com/kresnakar/sistem-krs/internal/kesetaraan"
	"github.com/kresnakar/sistem-krs/internal/krs"
	"github.com/kresnakar/sistem-krs/internal/matakuliah"
	"github.com/kresnakar/sistem-krs/internal/nilai"
	"github.com/kresnakar/sistem-krs/internal/prasyarat"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: file .env tidak ditemukan")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// 1. Registrasi modul Mata Kuliah (CRUD)
	mkRepo := matakuliah.NewRepository(db)
	mkHandler := matakuliah.NewHandler(mkRepo)
	mkHandler.RegisterRoutes(mux)

	// 2. Registrasi modul Kesetaraan Kurikulum
	ksRepo := kesetaraan.NewRepository(db)
	ksHandler := kesetaraan.NewHandler(ksRepo)
	ksHandler.RegisterRoutes(mux)

	// 3. Registrasi modul Prasyarat (Evaluasi AND/OR & Deteksi Siklus)
	psRepo := prasyarat.NewRepository(db)
	psService := prasyarat.NewService(psRepo, ksRepo)
	psHandler := prasyarat.NewHandler(psService, psRepo)
	psHandler.RegisterRoutes(mux)

	// 4. Registrasi Modul Kelas Paralel (Tugas Anda)
	kelasRepo := kelas.NewRepository(db)
	kelasService := kelas.NewService(kelasRepo)
	kelasHandler := kelas.NewHandler(kelasService)
	kelasHandler.RegisterRoutes(mux)

	// 5. Registrasi Modul Nilai Mahasiswa (Tugas Anda)

	nilaiRepo := nilai.NewRepository(db)
	nilaiService := nilai.NewService(nilaiRepo)
	nilaiHandler := nilai.NewHandler(nilaiService)
	nilaiHandler.RegisterRoutes(mux)

	// 6. Registrasi Modul KRS
	krsRepo := krs.NewRepository(db)
	krsService := krs.NewService(krsRepo)
	krsHandler := krs.NewHandler(krsService)
	krsHandler.RegisterRoutes(mux)

	// 7. Halaman Web Antarmuka
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Println("================================")
	log.Println("SISTEM KRS - Server berjalan di http://localhost:8081")
	log.Println("================================")

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}
