package kelas

// JadwalSesi merepresentasikan satu sesi jadwal untuk sebuah kelas.
type JadwalSesi struct {
	ID         int    `json:"id"`
	KelasID    int    `json:"kelas_id"`
	Hari       string `json:"hari"`
	JamMulai   string `json:"jam_mulai"`   // Format HH:MM
	JamSelesai string `json:"jam_selesai"` // Format HH:MM
	Ruang      string `json:"ruang"`
}

// KelasParalel merepresentasikan kelas paralel untuk suatu MK pada semester tertentu.
type KelasParalel struct {
	ID            int          `json:"id"`
	KodeMK        string       `json:"kode_mk"`
	NamaMK        string       `json:"nama_mk,omitempty"` // Join dari tabel MK
	SemesterID    int          `json:"semester_id"`
	TahunSemester string       `json:"tahun_semester,omitempty"` // Gabungan Tahun/Jenis
	KodeParalel   string       `json:"kode_paralel"`
	DosenPengampu string       `json:"dosen_pengampu"`
	Kuota         int          `json:"kuota"`
	Terisi        int          `json:"terisi"`
	Status        string       `json:"status"` // Dibuka, Ditutup, Dibatalkan
	Jadwal        []JadwalSesi `json:"jadwal,omitempty"`
}
