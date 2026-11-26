package prasyarat

type PrasyaratNode struct {
	ID           int             `json:"id"`
	KodeMK       string          `json:"kode_mk"`
	ParentID     *int            `json:"parent_id"`
	Tipe         string          `json:"tipe"` // "AND", "OR", "LEAF"
	KodeMKSyarat *string         `json:"kode_mk_syarat"`
	NilaiMinimal *string         `json:"nilai_minimal"`
	Children     []PrasyaratNode `json:"children,omitempty"`
}

type HasilEvaluasiNode struct {
	NodeID       int                 `json:"node_id"`
	Tipe         string              `json:"tipe"`
	KodeMKSyarat string              `json:"kode_mk_syarat,omitempty"`
	NilaiMinimal string              `json:"nilai_minimal,omitempty"`
	Terpenuhi    bool                `json:"terpenuhi"`
	Keterangan   string              `json:"keterangan"`
	Children     []HasilEvaluasiNode `json:"children,omitempty"`
}

type HasilEvaluasiPrasyarat struct {
	NIM         string              `json:"nim"`
	KodeMK      string              `json:"kode_mk"`
	Memenuhi    bool                `json:"memenuhi"`
	Ringkasan   string              `json:"ringkasan"`
	DetailPohon []HasilEvaluasiNode `json:"detail_pohon,omitempty"`
}