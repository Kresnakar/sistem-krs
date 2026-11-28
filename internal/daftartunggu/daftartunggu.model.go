package daftartunggu

import "time"

// StatusDaftarTunggu merepresentasikan kondisi antrean mahasiswa.
type StatusDaftarTunggu string

const (
	StatusMenunggu           StatusDaftarTunggu = "Menunggu"
	StatusMenungguKonfirmasi StatusDaftarTunggu = "MenungguKonfirmasi"
	StatusDilewati           StatusDaftarTunggu = "Dilewati"
	StatusDiterima           StatusDaftarTunggu = "Diterima"
	StatusDitolak            StatusDaftarTunggu = "Ditolak"
	StatusKedaluwarsa        StatusDaftarTunggu = "Kedaluwarsa"
	StatusKeluar             StatusDaftarTunggu = "Keluar"
)

// DaftarTunggu merepresentasikan satu antrean mahasiswa pada suatu kelas.
type DaftarTunggu struct {
	ID              int                `json:"id"`
	NIM             string             `json:"nim"`
	KelasID         int                `json:"kelas_id"`
	NomorAntrean    int                `json:"nomor_antrean"`
	PosisiAntrean   int                `json:"posisi_antrean"`
	Status          StatusDaftarTunggu `json:"status"`
	WaktuDaftar     time.Time          `json:"waktu_daftar"`
	AlasanDilewati  *string            `json:"alasan_dilewati,omitempty"`
	BatasKonfirmasi *time.Time         `json:"batas_konfirmasi,omitempty"`
}

// KonfirmasiRequest merupakan data yang dikirim mahasiswa
// ketika menerima atau menolak tawaran kursi.
type KonfirmasiRequest struct {
	Aksi string `json:"aksi"`
}

// KapasitasKelas menyimpan informasi kapasitas kelas
// dan jumlah kursi yang sedang ditahan.
type KapasitasKelas struct {
	Kuota   int
	Terisi  int
	Ditahan int
}

// KursiTersedia menghitung kursi yang benar-benar tersedia.
func (k KapasitasKelas) KursiTersedia() int {
	tersedia := k.Kuota - k.Terisi - k.Ditahan

	if tersedia < 0 {
		return 0
	}

	return tersedia
}

// HasilPromosi menyimpan hasil satu proses promosi antrean.
type HasilPromosi struct {
	KelasID       int            `json:"kelas_id"`
	KursiTersedia int            `json:"kursi_tersedia"`
	Dipromosikan  []DaftarTunggu `json:"dipromosikan"`
	Dilewati      []DaftarTunggu `json:"dilewati"`
}
