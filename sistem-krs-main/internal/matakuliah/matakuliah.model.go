package matakuliah

type MataKuliah struct {
	KodeMK    string `json:"kode_mk"`
	Nama      string `json:"nama"`
	SKS       int    `json:"sks"`
	Kurikulum string `json:"kurikulum"`
	Aktif     bool   `json:"aktif"`
}