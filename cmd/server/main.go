package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/kresnakar/sistem-krs/internal/auth"
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

	// Inisialisasi session
	store, err := auth.NewSessionStore()
	if err != nil {
		log.Fatal("Gagal menginisialisasi session: ", err)
	}

	// Router khusus modul administrasi
	adminMux := http.NewServeMux()

	// 1. Modul Mata Kuliah
	mkRepo := matakuliah.NewRepository(db)
	mkHandler := matakuliah.NewHandler(mkRepo)
	mkHandler.RegisterRoutes(adminMux)

	// 2. Modul Kesetaraan Kurikulum
	ksRepo := kesetaraan.NewRepository(db)
	ksHandler := kesetaraan.NewHandler(ksRepo)
	ksHandler.RegisterRoutes(adminMux)

	// 3. Modul Prasyarat
	psRepo := prasyarat.NewRepository(db)
	psService := prasyarat.NewService(psRepo, ksRepo)
	psHandler := prasyarat.NewHandler(psService, psRepo)
	psHandler.RegisterRoutes(adminMux)

	// 4. Modul Kelas Paralel
	kelasRepo := kelas.NewRepository(db)
	kelasService := kelas.NewService(kelasRepo)
	kelasHandler := kelas.NewHandler(kelasService)
	kelasHandler.RegisterRoutes(adminMux)

	// 5. Modul Nilai Mahasiswa
	nilaiRepo := nilai.NewRepository(db)
	nilaiService := nilai.NewService(nilaiRepo)
	nilaiHandler := nilai.NewHandler(nilaiService)
	nilaiHandler.RegisterRoutes(adminMux)

	// Router utama
	rootMux := http.NewServeMux()

	// Login dan logout
	rootMux.HandleFunc("/login", auth.LoginHandler(db, store))
	rootMux.HandleFunc("/logout", auth.LogoutHandler(store))

	// Dashboard Admin Prodi
	adminPage := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	rootMux.Handle(
		"/admin",
		auth.RequireAuth(store)(
			auth.RequireRole(store, "Admin Prodi")(adminPage),
		),
	)

	// API modul hanya untuk Admin Prodi
	rootMux.Handle(
		"/api/",
		auth.RequireAuth(store)(
			auth.RequireRole(store, "Admin Prodi")(adminMux),
		),
	)

	// Dashboard role lain sementara belum tersedia
	placeholder := func(role string) http.Handler {
		return auth.RequireAuth(store)(
			auth.RequireRole(store, role)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(
						w,
						"Dashboard "+role+" belum dibuat",
						http.StatusNotImplemented,
					)
				}),
			),
		)
	}

	rootMux.Handle("/mahasiswa", placeholder("Mahasiswa"))
	rootMux.Handle("/dosen", placeholder("Dosen Wali"))
	rootMux.Handle("/kaprodi", placeholder("Kaprodi"))

	// 6. Modul KRS (dari feature/krs) — didaftarkan ke rootMux
	krsRepo := krs.NewRepository(db)
	krsService := krs.NewService(krsRepo)
	krsHandler := krs.NewHandler(krsService)
	krsHandler.RegisterRoutes(rootMux)

	// Halaman utama
	rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	log.Println("================================")
	log.Println("SISTEM KRS - Server berjalan di http://localhost:8081")
	log.Println("================================")

	if err := http.ListenAndServe(":8081", rootMux); err != nil {
		log.Fatal(err)
	}
}