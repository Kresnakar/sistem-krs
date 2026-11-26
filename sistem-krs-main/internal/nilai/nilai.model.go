package nilai

type NilaiMahasiswa struct {
	ID         int     `json:"id"`
	NIM        string  `json:"nim"`
	NamaMhs    string  `json:"nama_mhs,omitempty"`
	KodeMK     string  `json:"kode_mk"`
	NamaMK     string  `json:"nama_mk,omitempty"`
	SemesterID int     `json:"semester_id"`
	NilaiHuruf string  `json:"nilai_huruf"`
	Bobot      float64 `json:"bobot"`
	SKS        int     `json:"sks"`
}

// TranskripEntry merepresentasikan hasil terbaik per MK untuk seorang mahasiswa.
type TranskripEntry struct {
	NIM        string  `json:"nim"`
	KodeMK     string  `json:"kode_mk"`
	NamaMK     string  `json:"nama_mk"`
	SKS        int     `json:"sks"`
	NilaiHuruf string  `json:"nilai_huruf"`
	Bobot      float64 `json:"bobot"`
}

// RingkasanSemester merepresentasikan IPS dan total SKS pada semester tertentu.
type RingkasanSemester struct {
	SemesterID    int     `json:"semester_id"`
	TahunAkademik string  `json:"tahun_akademik"`
	JenisSemester string  `json:"jenis_semester"`
	TotalSKS      int     `json:"total_sks"`
	TotalKualitas float64 `json:"total_kualitas"` // SKS * Bobot
	IPS           float64 `json:"ips"`
}
