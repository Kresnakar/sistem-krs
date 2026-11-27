package krs

import "time"

// KRSRingkas dipakai dosen wali untuk melihat daftar KRS bimbingan.
type KRSRingkas struct {
	ID               int        `json:"id"`
	NIM              string     `json:"nim"`
	NamaMhs          string     `json:"nama_mhs"`
	SemesterID       int        `json:"semester_id"`
	Status           string     `json:"status"`
	TotalSKS         int        `json:"total_sks"`
	BatasSKS         int        `json:"batas_sks"`
	DiajukanPada     *time.Time `json:"diajukan_pada"`
	DisetujuiPada    *time.Time `json:"disetujui_pada"`
	CatatanPenolakan *string    `json:"catatan_penolakan"`
}

// KRSDetailLengkap memperkaya detail KRS dengan info mata kuliah dan kelas.
type KRSDetailLengkap struct {
	ID                int    `json:"id"`
	KRSID             int    `json:"krs_id"`
	KelasID           int    `json:"kelas_id"`
	KodeMK            string `json:"kode_mk"`
	NamaMK            string `json:"nama_mk"`
	SKS               int    `json:"sks"`
	KodeParalel       string `json:"kode_paralel"`
	Status            string `json:"status"`
	JenisPerubahan    string `json:"jenis_perubahan"`
	StatusPersetujuan string `json:"status_persetujuan"`
}

// KRSLengkap adalah KRS beserta detail kelas yang sudah diperkaya.
type KRSLengkap struct {
	KRS
	DetailLengkap []KRSDetailLengkap `json:"detail_lengkap"`
}

type PermintaanPerubahan struct {
	ID             int                         `json:"id"`
	KRSID          int                         `json:"krs_id"`
	NIM            string                      `json:"nim"`
	NamaMhs        string                      `json:"nama_mhs,omitempty"`
	JenisPerubahan string                      `json:"jenis_perubahan"`
	Alasan         string                      `json:"alasan"`
	Status         string                      `json:"status"`
	DiajukanPada   time.Time                   `json:"diajukan_pada"`
	DiprosesOleh   *int                        `json:"diproses_oleh"`
	DiprosesPada   *time.Time                  `json:"diproses_pada"`
	Catatan        *string                     `json:"catatan"`
	Detail         []PermintaanPerubahanDetail `json:"detail,omitempty"`
}

type PermintaanPerubahanDetail struct {
	ID           int    `json:"id"`
	PermintaanID int    `json:"permintaan_id"`
	KelasID      int    `json:"kelas_id"`
	KodeMK       string `json:"kode_mk,omitempty"`
	NamaMK       string `json:"nama_mk,omitempty"`
	KodeParalel  string `json:"kode_paralel,omitempty"`
}

type AjukanPerubahanRequest struct {
	JenisPerubahan string `json:"jenis_perubahan"`
	Alasan         string `json:"alasan"`
	KelasIDs       []int  `json:"kelas_ids"`
	KelasID        int    `json:"kelas_id"`
}
