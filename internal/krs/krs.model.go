package krs

import "time"

type KRS struct {
	ID               int         `json:"id"`
	NIM              string      `json:"nim"`
	SemesterID       int         `json:"semester_id"`
	Status           string      `json:"status"`
	DosenWaliID      *int        `json:"dosen_wali_id"`
	IPSAcuan         float64     `json:"ips_acuan"`
	BatasSKS         int         `json:"batas_sks"`
	TotalSKS         int         `json:"total_sks"`
	DiajukanPada     *time.Time  `json:"diajukan_pada"`
	DisetujuiPada    *time.Time  `json:"disetujui_pada"`
	CatatanPenolakan *string     `json:"catatan_penolakan"`
	Detail           []KRSDetail `json:"detail,omitempty"`
}

type KRSDetail struct {
	ID                int    `json:"id"`
	KRSID             int    `json:"krs_id"`
	KelasID           int    `json:"kelas_id"`
	Status            string `json:"status"`
	JenisPerubahan    string `json:"jenis_perubahan"`
	StatusPersetujuan string `json:"status_persetujuan"`
}
