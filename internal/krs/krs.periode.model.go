package krs

import "time"

// Semester merepresentasikan data semester akademik.
type Semester struct {
	ID            int    `json:"id"`
	TahunAkademik string `json:"tahun_akademik"`
	Jenis         string `json:"jenis"`
	StatusAktif   bool   `json:"status_aktif"`
}

// PeriodeKRS merepresentasikan jendela waktu pengisian dan perubahan KRS.
type PeriodeKRS struct {
	ID                 int       `json:"id"`
	SemesterID         int       `json:"semester_id"`
	MulaiKRS           time.Time `json:"mulai_krs"`
	SelesaiKRS         time.Time `json:"selesai_krs"`
	MulaiPerubahan     time.Time `json:"mulai_perubahan"`
	SelesaiPerubahan   time.Time `json:"selesai_perubahan"`
	TanggalMulaiKuliah time.Time `json:"tanggal_mulai_kuliah"`

	// Field turunan untuk memudahkan tampilan UI.
	KRSDibuka       bool `json:"krs_dibuka"`
	PerubahanDibuka bool `json:"perubahan_dibuka"`
}

type SemesterRequest struct {
	TahunAkademik string `json:"tahun_akademik"`
	Jenis         string `json:"jenis"`
	StatusAktif   bool   `json:"status_aktif"`
}

type PeriodeRequest struct {
	SemesterID         int    `json:"semester_id"`
	MulaiKRS           string `json:"mulai_krs"`
	SelesaiKRS         string `json:"selesai_krs"`
	MulaiPerubahan     string `json:"mulai_perubahan"`
	SelesaiPerubahan   string `json:"selesai_perubahan"`
	TanggalMulaiKuliah string `json:"tanggal_mulai_kuliah"`
}

type BukaPeriodeRequest struct {
	Jenis   string `json:"jenis"` // "krs" atau "perubahan"
	Mulai   string `json:"mulai"`
	Selesai string `json:"selesai"`
}

type TutupPeriodeRequest struct {
	Jenis string `json:"jenis"` // "krs" atau "perubahan"
}

// KelasInfo adalah data ringkas kelas untuk dipilih mahasiswa.
type KelasInfo struct {
	ID            int    `json:"id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SemesterID    int    `json:"semester_id"`
	KodeParalel   string `json:"kode_paralel"`
	DosenPengampu string `json:"dosen_pengampu"`
	Kuota         int    `json:"kuota"`
	Terisi        int    `json:"terisi"`
	Status        string `json:"status"`
}
