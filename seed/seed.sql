-- ============================================================
-- SEED DATA SISTEM KRS
-- Berdasarkan struktur migrations/001_initial_schema.sql
-- Data berikut adalah data dummy untuk pengujian lokal.
-- Jalankan setelah migration/schema berhasil dibuat.
-- ============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 1;
START TRANSACTION;

-- 1. DOSEN WALI
INSERT INTO dosen_wali (id, nidn, nama) VALUES
(1, '0012088501', 'Dr. Andi Pratama, S.Kom., M.Kom.'),
(2, '0021098702', 'Rina Wulandari, S.Kom., M.Cs.')
ON DUPLICATE KEY UPDATE
nidn = VALUES(nidn), nama = VALUES(nama);

-- 2. MAHASISWA
INSERT INTO mahasiswa (nim, nama, angkatan, kurikulum, dosen_wali_id, status_aktif) VALUES
('230001', 'Aditya Ramadhan', 2023, '2023', 1, 1),
('230002', 'Nadia Putri Lestari', 2023, '2023', 1, 1),
('230003', 'Bima Saputra', 2023, '2023', 2, 1),
('240001', 'Citra Ayu Maharani', 2024, '2024', 2, 1)
ON DUPLICATE KEY UPDATE
nama = VALUES(nama), angkatan = VALUES(angkatan),
kurikulum = VALUES(kurikulum), dosen_wali_id = VALUES(dosen_wali_id),
status_aktif = VALUES(status_aktif);

-- 3. MATA KULIAH
INSERT INTO mata_kuliah (kode_mk, nama, sks, kurikulum, aktif) VALUES
('IF101', 'Algoritma dan Pemrograman', 3, '2023', 1),
('IF102', 'Matematika Diskrit', 3, '2023', 1),
('IF103', 'Pengantar Teknologi Informasi', 2, '2023', 1),
('IF201', 'Struktur Data', 3, '2023', 1),
('IF202', 'Basis Data', 3, '2023', 1),
('IF203', 'Jaringan Komputer', 3, '2023', 1),
('IF301', 'Kecerdasan Buatan', 3, '2023', 1),
('IF302', 'Pemrograman Web', 3, '2023', 1),
('IF001', 'Dasar Pemrograman (Kurikulum Lama)', 3, '2020', 1)
ON DUPLICATE KEY UPDATE
nama = VALUES(nama), sks = VALUES(sks),
kurikulum = VALUES(kurikulum), aktif = VALUES(aktif);

-- 4. SEMESTER
INSERT INTO semester (id, tahun_akademik, jenis, status_aktif) VALUES
(1, '2025/2026', 'Genap', 0),
(2, '2026/2027', 'Ganjil', 1)
ON DUPLICATE KEY UPDATE
tahun_akademik = VALUES(tahun_akademik),
jenis = VALUES(jenis), status_aktif = VALUES(status_aktif);

-- 5. SKALA NILAI
INSERT INTO skala_nilai (nilai_huruf, bobot, lulus) VALUES
('A', 4.00, 1),
('AB', 3.50, 1),
('B', 3.00, 1),
('BC', 2.50, 1),
('C', 2.00, 1),
('D', 1.00, 0),
('E', 0.00, 0)
ON DUPLICATE KEY UPDATE
bobot = VALUES(bobot), lulus = VALUES(lulus);

-- 6. KESETARAAN MATA KULIAH
INSERT INTO kesetaraan_mk (mk_lama, mk_baru, keterangan) VALUES
('IF001', 'IF101', 'Penyetaraan mata kuliah Dasar Pemrograman ke Algoritma dan Pemrograman')
ON DUPLICATE KEY UPDATE
keterangan = VALUES(keterangan);

-- 7. PRASYARAT MATA KULIAH
-- IF301 membutuhkan IF201 (AND dengan satu syarat).
-- IF302 membutuhkan IF101 (AND dengan satu syarat).
INSERT INTO prasyarat_node
(id, kode_mk, parent_id, tipe, kode_mk_syarat, nilai_minimal) VALUES
(1, 'IF301', NULL, 'AND', NULL, NULL),
(2, 'IF301', 1, 'LEAF', 'IF201', 'C'),
(3, 'IF302', NULL, 'AND', NULL, NULL),
(4, 'IF302', 3, 'LEAF', 'IF101', 'C')
ON DUPLICATE KEY UPDATE
kode_mk = VALUES(kode_mk), parent_id = VALUES(parent_id),
tipe = VALUES(tipe), kode_mk_syarat = VALUES(kode_mk_syarat),
nilai_minimal = VALUES(nilai_minimal);

-- 8. KELAS
INSERT INTO kelas
(id, kode_mk, semester_id, kode_paralel, dosen_pengampu, kuota, terisi, status) VALUES
(1, 'IF201', 2, 'A', 'Dr. Andi Pratama, S.Kom., M.Kom.', 35, 2, 'Dibuka'),
(2, 'IF202', 2, 'A', 'Rina Wulandari, S.Kom., M.Cs.', 35, 2, 'Dibuka'),
(3, 'IF203', 2, 'A', 'Dr. Andi Pratama, S.Kom., M.Kom.', 30, 1, 'Dibuka'),
(4, 'IF301', 2, 'A', 'Rina Wulandari, S.Kom., M.Cs.', 25, 25, 'Ditutup'),
(5, 'IF302', 2, 'A', 'Dr. Andi Pratama, S.Kom., M.Kom.', 30, 1, 'Dibuka'),
(6, 'IF101', 2, 'A', 'Rina Wulandari, S.Kom., M.Cs.', 40, 1, 'Dibuka'),
(7, 'IF102', 2, 'A', 'Dr. Andi Pratama, S.Kom., M.Kom.', 40, 1, 'Dibuka'),
(8, 'IF103', 2, 'A', 'Rina Wulandari, S.Kom., M.Cs.', 40, 1, 'Dibuka')
ON DUPLICATE KEY UPDATE
kode_mk = VALUES(kode_mk), semester_id = VALUES(semester_id),
kode_paralel = VALUES(kode_paralel), dosen_pengampu = VALUES(dosen_pengampu),
kuota = VALUES(kuota), terisi = VALUES(terisi), status = VALUES(status);

-- 9. JADWAL KELAS
INSERT INTO jadwal_kelas
(id, kelas_id, hari, jam_mulai, jam_selesai, ruang) VALUES
(1, 1, 'Senin', '08:00:00', '10:30:00', 'D2-101'),
(2, 2, 'Senin', '13:00:00', '15:30:00', 'D2-102'),
(3, 3, 'Selasa', '08:00:00', '10:30:00', 'D2-201'),
(4, 4, 'Selasa', '13:00:00', '15:30:00', 'D2-202'),
(5, 5, 'Rabu', '08:00:00', '10:30:00', 'D3-101'),
(6, 6, 'Rabu', '13:00:00', '14:40:00', 'D3-102'),
(7, 7, 'Kamis', '08:00:00', '10:30:00', 'D3-201'),
(8, 8, 'Jumat', '08:00:00', '09:40:00', 'D3-202')
ON DUPLICATE KEY UPDATE
kelas_id = VALUES(kelas_id), hari = VALUES(hari),
jam_mulai = VALUES(jam_mulai), jam_selesai = VALUES(jam_selesai),
ruang = VALUES(ruang);

-- 10. PERIODE KRS
INSERT INTO periode_krs
(id, semester_id, mulai_krs, selesai_krs, mulai_perubahan, selesai_perubahan, tanggal_mulai_kuliah) VALUES
(1, 2, '2026-08-10 08:00:00', '2026-08-21 23:59:00',
 '2026-08-24 08:00:00', '2026-08-28 23:59:00', '2026-08-31')
ON DUPLICATE KEY UPDATE
semester_id = VALUES(semester_id), mulai_krs = VALUES(mulai_krs),
selesai_krs = VALUES(selesai_krs), mulai_perubahan = VALUES(mulai_perubahan),
selesai_perubahan = VALUES(selesai_perubahan),
tanggal_mulai_kuliah = VALUES(tanggal_mulai_kuliah);

-- 11. RIWAYAT NILAI
INSERT INTO riwayat_nilai
(nim, kode_mk, semester_id, nilai_huruf, bobot, lulus) VALUES
('230001', 'IF101', 1, 'A', 4.00, 1),
('230001', 'IF102', 1, 'B', 3.00, 1),
('230001', 'IF103', 1, 'AB', 3.50, 1),
('230002', 'IF101', 1, 'B', 3.00, 1),
('230002', 'IF102', 1, 'AB', 3.50, 1),
('230003', 'IF101', 1, 'C', 2.00, 1),
('240001', 'IF101', 1, 'A', 4.00, 1)
ON DUPLICATE KEY UPDATE
nilai_huruf = VALUES(nilai_huruf), bobot = VALUES(bobot),
lulus = VALUES(lulus);

-- 12. KRS
INSERT INTO krs
(id, nim, semester_id, status, dosen_wali_id, ips_acuan, batas_sks,
 total_sks, diajukan_pada, disetujui_pada, catatan_penolakan) VALUES
(1, '230001', 2, 'Disetujui', 1, 3.50, 24, 9,
 '2026-08-12 09:00:00', '2026-08-12 13:00:00', NULL),
(2, '230002', 2, 'Diajukan', 1, 3.25, 24, 6,
 '2026-08-13 10:00:00', NULL, NULL),
(3, '230003', 2, 'Draft', 2, 2.00, 21, 3,
 NULL, NULL, NULL),
(4, '240001', 2, 'Draft', 2, 4.00, 24, 2,
 NULL, NULL, NULL)
ON DUPLICATE KEY UPDATE
nim = VALUES(nim), semester_id = VALUES(semester_id),
status = VALUES(status), dosen_wali_id = VALUES(dosen_wali_id),
ips_acuan = VALUES(ips_acuan), batas_sks = VALUES(batas_sks),
total_sks = VALUES(total_sks), diajukan_pada = VALUES(diajukan_pada),
disetujui_pada = VALUES(disetujui_pada),
catatan_penolakan = VALUES(catatan_penolakan);

-- 13. DETAIL KRS
INSERT INTO krs_detail
(id, krs_id, kelas_id, status, jenis_perubahan, status_persetujuan) VALUES
(1, 1, 1, 'Terdaftar', 'Awal', 'Final'),
(2, 1, 2, 'Terdaftar', 'Awal', 'Final'),
(3, 1, 3, 'Terdaftar', 'Awal', 'Final'),
(4, 2, 1, 'Terdaftar', 'Awal', 'Menunggu Persetujuan'),
(5, 2, 2, 'Terdaftar', 'Awal', 'Menunggu Persetujuan'),
(6, 3, 5, 'Terdaftar', 'Awal', 'Menunggu Persetujuan'),
(7, 4, 8, 'Terdaftar', 'Awal', 'Menunggu Persetujuan')
ON DUPLICATE KEY UPDATE
krs_id = VALUES(krs_id), kelas_id = VALUES(kelas_id),
status = VALUES(status), jenis_perubahan = VALUES(jenis_perubahan),
status_persetujuan = VALUES(status_persetujuan);

-- 14. DAFTAR TUNGGU
INSERT INTO daftar_tunggu
(id, nim, kelas_id, nomor_antrean, status, waktu_daftar) VALUES
(1, '230003', 4, 1, 'Menunggu', '2026-08-14 09:15:00'),
(2, '240001', 4, 2, 'Menunggu', '2026-08-14 09:30:00')
ON DUPLICATE KEY UPDATE
nim = VALUES(nim), kelas_id = VALUES(kelas_id),
nomor_antrean = VALUES(nomor_antrean), status = VALUES(status),
waktu_daftar = VALUES(waktu_daftar);

-- 15. PERMINTAAN PERUBAHAN KRS
INSERT INTO permintaan_perubahan
(id, krs_id, jenis_perubahan, alasan, status, diajukan_pada,
 diproses_oleh, diproses_pada, catatan) VALUES
(1, 2, 'Tambah Kelas', 'Mahasiswa ingin menambah mata kuliah pilihan.', 'Menunggu',
 '2026-08-25 10:00:00', NULL, NULL, NULL),
(2, 1, 'Batal Kelas', 'Penyesuaian jadwal perkuliahan.', 'Disetujui',
 '2026-08-25 11:00:00', 1, '2026-08-25 14:00:00', 'Permintaan disetujui.')
ON DUPLICATE KEY UPDATE
krs_id = VALUES(krs_id), jenis_perubahan = VALUES(jenis_perubahan),
alasan = VALUES(alasan), status = VALUES(status),
diajukan_pada = VALUES(diajukan_pada), diproses_oleh = VALUES(diproses_oleh),
diproses_pada = VALUES(diproses_pada), catatan = VALUES(catatan);

-- 16. DETAIL PERMINTAAN PERUBAHAN
INSERT INTO permintaan_perubahan_detail (id, permintaan_id, kelas_id) VALUES
(1, 1, 5),
(2, 2, 3)
ON DUPLICATE KEY UPDATE
permintaan_id = VALUES(permintaan_id), kelas_id = VALUES(kelas_id);

-- 17. LOG AUDIT
INSERT INTO log_audit
(id, aktor_tipe, aktor_id, aktivitas, nama_tabel, id_data,
 data_sebelum, data_sesudah, ip_address, waktu_aktivitas) VALUES
(1, 'Mahasiswa', '230001', 'Mengajukan KRS', 'krs', '1',
 NULL, JSON_OBJECT('status', 'Diajukan', 'total_sks', 9),
 '127.0.0.1', '2026-08-12 09:00:00'),
(2, 'Dosen Wali', '1', 'Menyetujui KRS', 'krs', '1',
 JSON_OBJECT('status', 'Diajukan'), JSON_OBJECT('status', 'Disetujui'),
 '127.0.0.1', '2026-08-12 13:00:00')
ON DUPLICATE KEY UPDATE
aktor_tipe = VALUES(aktor_tipe), aktor_id = VALUES(aktor_id),
aktivitas = VALUES(aktivitas), nama_tabel = VALUES(nama_tabel),
id_data = VALUES(id_data), data_sebelum = VALUES(data_sebelum),
data_sesudah = VALUES(data_sesudah), ip_address = VALUES(ip_address),
waktu_aktivitas = VALUES(waktu_aktivitas);

-- 18. NOTIFIKASI
INSERT INTO notifikasi
(id, penerima_tipe, penerima_id, judul, pesan, jenis,
 status_baca, dibuat_pada, dibaca_pada) VALUES
(1, 'Mahasiswa', '230001', 'KRS Disetujui',
 'Pengisian KRS semester 2026/2027 Ganjil telah disetujui dosen wali.',
 'Persetujuan', 0, '2026-08-12 13:05:00', NULL),
(2, 'Mahasiswa', '230002', 'KRS Menunggu Persetujuan',
 'Pengajuan KRS kamu sedang menunggu persetujuan dosen wali.',
 'Informasi', 0, '2026-08-13 10:05:00', NULL),
(3, 'Mahasiswa', '230003', 'Daftar Tunggu Kelas',
 'Kamu masuk daftar tunggu untuk kelas IF301 paralel A.',
 'Peringatan', 0, '2026-08-14 09:20:00', NULL)
ON DUPLICATE KEY UPDATE
penerima_tipe = VALUES(penerima_tipe), penerima_id = VALUES(penerima_id),
judul = VALUES(judul), pesan = VALUES(pesan), jenis = VALUES(jenis),
status_baca = VALUES(status_baca), dibuat_pada = VALUES(dibuat_pada),
dibaca_pada = VALUES(dibaca_pada);

COMMIT;

-- ============================================================
-- SELESAI
-- Periksa jumlah data dengan:
-- SELECT 'mahasiswa' AS tabel, COUNT(*) AS jumlah FROM mahasiswa
-- UNION ALL SELECT 'mata_kuliah', COUNT(*) FROM mata_kuliah
-- UNION ALL SELECT 'kelas', COUNT(*) FROM kelas
-- UNION ALL SELECT 'krs', COUNT(*) FROM krs;
-- ============================================================
