package prasyarat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kresnakar/sistem-krs/internal/kesetaraan"
)

var ErrSiklusTerdeteksi = errors.New("siklus prasyarat terdeteksi")

type Service struct {
	repo   *Repository
	ksRepo *kesetaraan.Repository
}

func NewService(repo *Repository, ksRepo *kesetaraan.Repository) *Service {
	return &Service{
		repo:   repo,
		ksRepo: ksRepo,
	}
}

// BangunPohon menyusun daftar flat prasyarat_node menjadi struktur pohon hierarkis.
func (s *Service) BangunPohon(ctx context.Context, kodeMK string) ([]PrasyaratNode, error) {
	nodes, err := s.repo.GetNodesByKodeMK(ctx, kodeMK)
	if err != nil {
		return nil, err
	}
	return buildTree(nodes), nil
}

func buildTree(nodes []PrasyaratNode) []PrasyaratNode {
	childrenMap := make(map[int][]PrasyaratNode)
	var roots []PrasyaratNode

	for _, n := range nodes {
		if n.ParentID != nil {
			childrenMap[*n.ParentID] = append(childrenMap[*n.ParentID], n)
		}
	}

	var attach func(node PrasyaratNode) PrasyaratNode
	attach = func(node PrasyaratNode) PrasyaratNode {
		kids := childrenMap[node.ID]
		for _, k := range kids {
			node.Children = append(node.Children, attach(k))
		}
		return node
	}

	for _, n := range nodes {
		if n.ParentID == nil {
			roots = append(roots, attach(n))
		}
	}
	return roots
}

// CekSiklus mendeteksi apakah penambahan syarat baru (kodeMK membutuhkan syaratMK)
// menimbulkan ketergantungan melingkar (siklus), termasuk melalui mata kuliah setara.
func (s *Service) CekSiklus(ctx context.Context, kodeMK, syaratMK string) ([]string, error) {
	setaraTarget, err := s.ksRepo.GetSetaraCodes(ctx, kodeMK)
	if err != nil {
		return nil, err
	}
	targetSet := make(map[string]bool)
	for _, c := range setaraTarget {
		targetSet[c] = true
	}

	if targetSet[syaratMK] {
		return []string{kodeMK, syaratMK}, fmt.Errorf("%w: %s -> %s", ErrSiklusTerdeteksi, kodeMK, syaratMK)
	}

	semuaNode, err := s.repo.GetAllNodes(ctx)
	if err != nil {
		return nil, err
	}

	// Graf berarah: MK -> daftar MK prasyaratnya
	adj := make(map[string][]string)
	for _, n := range semuaNode {
		if n.Tipe == "LEAF" && n.KodeMKSyarat != nil && *n.KodeMKSyarat != "" {
			adj[n.KodeMK] = append(adj[n.KodeMK], *n.KodeMKSyarat)
		}
	}

	// DFS untuk mencari lintasan dari syaratMK kembali ke kodeMK (atau yang setara)
	visited := make(map[string]bool)
	var path []string

	var dfs func(curr string) bool
	dfs = func(curr string) bool {
		visited[curr] = true
		path = append(path, curr)

		if targetSet[curr] {
			return true
		}

		// Periksa prasyarat langsung maupun melalui mata kuliah yang setara dengan curr
		equivCurr, _ := s.ksRepo.GetSetaraCodes(ctx, curr)
		for _, eq := range equivCurr {
			for _, next := range adj[eq] {
				if !visited[next] {
					if dfs(next) {
						return true
					}
				}
			}
		}

		path = path[:len(path)-1]
		return false
	}

	if dfs(syaratMK) {
		fullCycle := append([]string{kodeMK}, path...)
		return fullCycle, fmt.Errorf("%w: %s", ErrSiklusTerdeteksi, strings.Join(fullCycle, " -> "))
	}

	return nil, nil
}

// TambahNode memvalidasi node baru dan mencegah siklus sebelum menyimpannya ke database.
func (s *Service) TambahNode(ctx context.Context, req PrasyaratNode) (*PrasyaratNode, error) {
	req.KodeMK = strings.TrimSpace(req.KodeMK)
	req.Tipe = strings.ToUpper(strings.TrimSpace(req.Tipe))

	if req.KodeMK == "" {
		return nil, errors.New("kode_mk wajib diisi")
	}
	if req.Tipe != "AND" && req.Tipe != "OR" && req.Tipe != "LEAF" {
		return nil, errors.New("tipe harus bernilai AND, OR, atau LEAF")
	}

	if req.Tipe == "LEAF" {
		if req.KodeMKSyarat == nil || strings.TrimSpace(*req.KodeMKSyarat) == "" {
			return nil, errors.New("kode_mk_syarat wajib diisi untuk tipe LEAF")
		}
		syarat := strings.TrimSpace(*req.KodeMKSyarat)
		req.KodeMKSyarat = &syarat

		if _, err := s.CekSiklus(ctx, req.KodeMK, syarat); err != nil {
			return nil, err
		}

		if req.NilaiMinimal == nil || strings.TrimSpace(*req.NilaiMinimal) == "" {
			def := "C"
			req.NilaiMinimal = &def
		} else {
			nm := strings.ToUpper(strings.TrimSpace(*req.NilaiMinimal))
			req.NilaiMinimal = &nm
		}
	} else {
		req.KodeMKSyarat = nil
		req.NilaiMinimal = nil
	}

	id, err := s.repo.CreateNode(ctx, req)
	if err != nil {
		return nil, err
	}
	req.ID = id
	return &req, nil
}

// Evaluasi memeriksa apakah mahasiswa (nim) memenuhi aturan prasyarat AND/OR
// untuk mengambil kodeMK, dengan memperhitungkan kesetaraan kurikulum.
func (s *Service) Evaluasi(ctx context.Context, nim, kodeMK string) (*HasilEvaluasiPrasyarat, error) {
	roots, err := s.BangunPohon(ctx, kodeMK)
	if err != nil {
		return nil, err
	}

	if len(roots) == 0 {
		return &HasilEvaluasiPrasyarat{
			NIM:       nim,
			KodeMK:    kodeMK,
			Memenuhi:  true,
			Ringkasan: "Mata kuliah tidak memiliki prasyarat",
		}, nil
	}

	skala, err := s.repo.GetSkalaNilai(ctx)
	if err != nil {
		return nil, err
	}

	riwayat, err := s.repo.GetRiwayatNilai(ctx, nim)
	if err != nil {
		return nil, err
	}

	var evalNode func(node PrasyaratNode) HasilEvaluasiNode
	evalNode = func(node PrasyaratNode) HasilEvaluasiNode {
		switch node.Tipe {
		case "LEAF":
			syaratMK := ""
			if node.KodeMKSyarat != nil {
				syaratMK = *node.KodeMKSyarat
			}
			nilaiMin := "C"
			if node.NilaiMinimal != nil && *node.NilaiMinimal != "" {
				nilaiMin = *node.NilaiMinimal
			}
			bobotMin := skala[nilaiMin]

			// Cari semua kode mata kuliah yang setara dengan syaratMK
			daftarSetara, _ := s.ksRepo.GetSetaraCodes(ctx, syaratMK)
			if len(daftarSetara) == 0 {
				daftarSetara = []string{syaratMK}
			}

			var nilaiTerbaik *NilaiMahasiswa
			for _, kodeCek := range daftarSetara {
				if nm, ada := riwayat[kodeCek]; ada {
					if nilaiTerbaik == nil || nm.Bobot > nilaiTerbaik.Bobot {
						copyNM := nm
						nilaiTerbaik = &copyNM
					}
				}
			}

			if nilaiTerbaik == nil {
				return HasilEvaluasiNode{
					NodeID:       node.ID,
					Tipe:         "LEAF",
					KodeMKSyarat: syaratMK,
					NilaiMinimal: nilaiMin,
					Terpenuhi:    false,
					Keterangan:   fmt.Sprintf("Belum pernah lulus %s (nilai minimal %s)", syaratMK, nilaiMin),
				}
			}

			infoSetara := nilaiTerbaik.KodeMK
			if nilaiTerbaik.KodeMK != syaratMK {
				infoSetara = fmt.Sprintf("%s setara %s", nilaiTerbaik.KodeMK, syaratMK)
			}

			terpenuhi := nilaiTerbaik.Bobot >= bobotMin
			statusTeks := "Memenuhi"
			if !terpenuhi {
				statusTeks = "Tidak memenuhi"
			}

			return HasilEvaluasiNode{
				NodeID:       node.ID,
				Tipe:         "LEAF",
				KodeMKSyarat: syaratMK,
				NilaiMinimal: nilaiMin,
				Terpenuhi:    terpenuhi,
				Keterangan: fmt.Sprintf("%s (%s): nilai %s (bobot %.2f, minimal %s/%.2f)",
					statusTeks, infoSetara, nilaiTerbaik.NilaiHuruf, nilaiTerbaik.Bobot, nilaiMin, bobotMin),
			}

		case "AND":
			if len(node.Children) == 0 {
				return HasilEvaluasiNode{
					NodeID:     node.ID,
					Tipe:       "AND",
					Terpenuhi:  true,
					Keterangan: "Grup AND kosong",
				}
			}
			semuaLulus := true
			var detail []HasilEvaluasiNode
			for _, child := range node.Children {
				res := evalNode(child)
				if !res.Terpenuhi {
					semuaLulus = false
				}
				detail = append(detail, res)
			}
			return HasilEvaluasiNode{
				NodeID:     node.ID,
				Tipe:       "AND",
				Terpenuhi:  semuaLulus,
				Keterangan: fmt.Sprintf("Evaluasi grup AND: %v", semuaLulus),
				Children:   detail,
			}

		case "OR":
			if len(node.Children) == 0 {
				return HasilEvaluasiNode{
					NodeID:     node.ID,
					Tipe:       "OR",
					Terpenuhi:  false,
					Keterangan: "Grup OR kosong",
				}
			}
			salahSatuLulus := false
			var detail []HasilEvaluasiNode
			for _, child := range node.Children {
				res := evalNode(child)
				if res.Terpenuhi {
					salahSatuLulus = true
				}
				detail = append(detail, res)
			}
			return HasilEvaluasiNode{
				NodeID:     node.ID,
				Tipe:       "OR",
				Terpenuhi:  salahSatuLulus,
				Keterangan: fmt.Sprintf("Evaluasi grup OR: %v", salahSatuLulus),
				Children:   detail,
			}

		default:
			return HasilEvaluasiNode{
				NodeID:     node.ID,
				Tipe:       node.Tipe,
				Terpenuhi:  false,
				Keterangan: "Tipe node tidak dikenali",
			}
		}
	}

	memenuhiSemua := true
	var detailPohon []HasilEvaluasiNode
	for _, root := range roots {
		hasilRoot := evalNode(root)
		if !hasilRoot.Terpenuhi {
			memenuhiSemua = false
		}
		detailPohon = append(detailPohon, hasilRoot)
	}

	ringkasan := fmt.Sprintf("Mahasiswa %s MEMENUHI prasyarat untuk mengambil %s", nim, kodeMK)
	if !memenuhiSemua {
		ringkasan = fmt.Sprintf("Mahasiswa %s BELUM MEMENUHI prasyarat untuk mengambil %s", nim, kodeMK)
	}

	return &HasilEvaluasiPrasyarat{
		NIM:         nim,
		KodeMK:      kodeMK,
		Memenuhi:    memenuhiSemua,
		Ringkasan:   ringkasan,
		DetailPohon: detailPohon,
	}, nil
}