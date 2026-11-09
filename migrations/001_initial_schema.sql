-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Generation Time: Nov 08, 2026 at 06:20 PM
-- Server version: 8.0.30
-- PHP Version: 8.1.10

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `sistem_krs`
--

-- --------------------------------------------------------

--
-- Table structure for table `daftar_tunggu`
--

CREATE TABLE `daftar_tunggu` (
  `id` int NOT NULL,
  `nim` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `kelas_id` int NOT NULL,
  `nomor_antrean` int NOT NULL,
  `status` enum('Menunggu','Ditawarkan','Diterima','Dibatalkan','Kedaluwarsa') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Menunggu',
  `waktu_daftar` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `dosen_wali`
--

CREATE TABLE `dosen_wali` (
  `id` int NOT NULL,
  `nidn` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `nama` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `jadwal_kelas`
--

CREATE TABLE `jadwal_kelas` (
  `id` int NOT NULL,
  `kelas_id` int NOT NULL,
  `hari` enum('Senin','Selasa','Rabu','Kamis','Jumat') COLLATE utf8mb4_unicode_ci NOT NULL,
  `jam_mulai` time NOT NULL,
  `jam_selesai` time NOT NULL,
  `ruang` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL
) ;

-- --------------------------------------------------------

--
-- Table structure for table `kelas`
--

CREATE TABLE `kelas` (
  `id` int NOT NULL,
  `kode_mk` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `semester_id` int NOT NULL,
  `kode_paralel` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `dosen_pengampu` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `kuota` int NOT NULL DEFAULT '0',
  `terisi` int NOT NULL DEFAULT '0',
  `status` enum('Dibuka','Ditutup','Dibatalkan') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Dibuka'
) ;

-- --------------------------------------------------------

--
-- Table structure for table `kesetaraan_mk`
--

CREATE TABLE `kesetaraan_mk` (
  `mk_lama` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `mk_baru` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `keterangan` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `krs`
--

CREATE TABLE `krs` (
  `id` int NOT NULL,
  `nim` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `semester_id` int NOT NULL,
  `status` enum('Draft','Diajukan','Disetujui','Ditolak') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Draft',
  `dosen_wali_id` int DEFAULT NULL,
  `ips_acuan` decimal(3,2) DEFAULT '0.00',
  `batas_sks` int NOT NULL DEFAULT '21',
  `total_sks` int NOT NULL DEFAULT '0',
  `diajukan_pada` datetime DEFAULT NULL,
  `disetujui_pada` datetime DEFAULT NULL,
  `catatan_penolakan` text COLLATE utf8mb4_unicode_ci
) ;

-- --------------------------------------------------------

--
-- Table structure for table `krs_detail`
--

CREATE TABLE `krs_detail` (
  `id` int NOT NULL,
  `krs_id` int NOT NULL,
  `kelas_id` int NOT NULL,
  `status` enum('Terdaftar','Dibatalkan') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Terdaftar',
  `jenis_perubahan` enum('Awal','Tambah','Batal') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Awal',
  `status_persetujuan` enum('Final','Menunggu Persetujuan') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Menunggu Persetujuan'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `log_audit`
--

CREATE TABLE `log_audit` (
  `id` bigint NOT NULL,
  `aktor_tipe` enum('Mahasiswa','Dosen Wali','Admin Prodi','Kaprodi') COLLATE utf8mb4_unicode_ci NOT NULL,
  `aktor_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `aktivitas` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `nama_tabel` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `id_data` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `data_sebelum` json DEFAULT NULL,
  `data_sesudah` json DEFAULT NULL,
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `waktu_aktivitas` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `mahasiswa`
--

CREATE TABLE `mahasiswa` (
  `nim` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `nama` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `angkatan` year NOT NULL,
  `kurikulum` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `dosen_wali_id` int DEFAULT NULL,
  `status_aktif` tinyint(1) NOT NULL DEFAULT '1'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `mata_kuliah`
--

CREATE TABLE `mata_kuliah` (
  `kode_mk` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `nama` varchar(150) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sks` int NOT NULL,
  `kurikulum` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `aktif` tinyint(1) NOT NULL DEFAULT '1'
) ;

-- --------------------------------------------------------

--
-- Table structure for table `notifikasi`
--

CREATE TABLE `notifikasi` (
  `id` bigint NOT NULL,
  `penerima_tipe` enum('Mahasiswa','Dosen Wali','Admin Prodi','Kaprodi') COLLATE utf8mb4_unicode_ci NOT NULL,
  `penerima_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `judul` varchar(150) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pesan` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `jenis` enum('Informasi','Peringatan','Persetujuan','Penolakan') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Informasi',
  `status_baca` tinyint(1) NOT NULL DEFAULT '0',
  `dibuat_pada` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `dibaca_pada` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `periode_krs`
--

CREATE TABLE `periode_krs` (
  `id` int NOT NULL,
  `semester_id` int NOT NULL,
  `mulai_krs` datetime NOT NULL,
  `selesai_krs` datetime NOT NULL,
  `mulai_perubahan` datetime NOT NULL,
  `selesai_perubahan` datetime NOT NULL,
  `tanggal_mulai_kuliah` date NOT NULL
) ;

-- --------------------------------------------------------

--
-- Table structure for table `permintaan_perubahan`
--

CREATE TABLE `permintaan_perubahan` (
  `id` int NOT NULL,
  `krs_id` int NOT NULL,
  `jenis_perubahan` enum('Tambah Kelas','Batal Kelas') COLLATE utf8mb4_unicode_ci NOT NULL,
  `alasan` text COLLATE utf8mb4_unicode_ci,
  `status` enum('Menunggu','Disetujui','Ditolak') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Menunggu',
  `diajukan_pada` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `diproses_oleh` int DEFAULT NULL,
  `diproses_pada` datetime DEFAULT NULL,
  `catatan` text COLLATE utf8mb4_unicode_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `permintaan_perubahan_detail`
--

CREATE TABLE `permintaan_perubahan_detail` (
  `id` int NOT NULL,
  `permintaan_id` int NOT NULL,
  `kelas_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `prasyarat_node`
--

CREATE TABLE `prasyarat_node` (
  `id` int NOT NULL,
  `kode_mk` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `parent_id` int DEFAULT NULL,
  `tipe` enum('AND','OR','LEAF') COLLATE utf8mb4_unicode_ci NOT NULL,
  `kode_mk_syarat` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `nilai_minimal` varchar(2) COLLATE utf8mb4_unicode_ci DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `riwayat_nilai`
--

CREATE TABLE `riwayat_nilai` (
  `nim` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `kode_mk` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `semester_id` int NOT NULL,
  `nilai_huruf` varchar(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `bobot` decimal(3,2) NOT NULL,
  `lulus` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `semester`
--

CREATE TABLE `semester` (
  `id` int NOT NULL,
  `tahun_akademik` varchar(9) COLLATE utf8mb4_unicode_ci NOT NULL,
  `jenis` enum('Ganjil','Genap','Antara') COLLATE utf8mb4_unicode_ci NOT NULL,
  `status_aktif` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `skala_nilai`
--

CREATE TABLE `skala_nilai` (
  `nilai_huruf` varchar(2) COLLATE utf8mb4_unicode_ci NOT NULL,
  `bobot` decimal(3,2) NOT NULL,
  `lulus` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `daftar_tunggu`
--
ALTER TABLE `daftar_tunggu`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_antrean` (`kelas_id`,`nomor_antrean`),
  ADD UNIQUE KEY `uq_tunggu_mahasiswa` (`nim`,`kelas_id`),
  ADD KEY `idx_tunggu_status` (`kelas_id`,`status`);

--
-- Indexes for table `dosen_wali`
--
ALTER TABLE `dosen_wali`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `nidn` (`nidn`);

--
-- Indexes for table `jadwal_kelas`
--
ALTER TABLE `jadwal_kelas`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_jadwal_kelas` (`kelas_id`),
  ADD KEY `idx_jadwal_hari` (`hari`);

--
-- Indexes for table `kelas`
--
ALTER TABLE `kelas`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_kelas_paralel` (`kode_mk`,`semester_id`,`kode_paralel`),
  ADD KEY `fk_kelas_semester` (`semester_id`);

--
-- Indexes for table `kesetaraan_mk`
--
ALTER TABLE `kesetaraan_mk`
  ADD PRIMARY KEY (`mk_lama`,`mk_baru`),
  ADD KEY `mk_baru` (`mk_baru`);

--
-- Indexes for table `krs`
--
ALTER TABLE `krs`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_krs_mahasiswa_semester` (`nim`,`semester_id`),
  ADD KEY `fk_krs_semester` (`semester_id`),
  ADD KEY `fk_krs_dosen` (`dosen_wali_id`);

--
-- Indexes for table `krs_detail`
--
ALTER TABLE `krs_detail`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_krs_kelas` (`krs_id`,`kelas_id`),
  ADD KEY `fk_detail_kelas` (`kelas_id`);

--
-- Indexes for table `log_audit`
--
ALTER TABLE `log_audit`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_audit_aktor` (`aktor_tipe`,`aktor_id`),
  ADD KEY `idx_audit_waktu` (`waktu_aktivitas`),
  ADD KEY `idx_audit_tabel` (`nama_tabel`);

--
-- Indexes for table `mahasiswa`
--
ALTER TABLE `mahasiswa`
  ADD PRIMARY KEY (`nim`),
  ADD KEY `fk_mahasiswa_dosen` (`dosen_wali_id`);

--
-- Indexes for table `mata_kuliah`
--
ALTER TABLE `mata_kuliah`
  ADD PRIMARY KEY (`kode_mk`);

--
-- Indexes for table `notifikasi`
--
ALTER TABLE `notifikasi`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_notifikasi_penerima` (`penerima_tipe`,`penerima_id`,`status_baca`);

--
-- Indexes for table `periode_krs`
--
ALTER TABLE `periode_krs`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_periode_semester` (`semester_id`);

--
-- Indexes for table `permintaan_perubahan`
--
ALTER TABLE `permintaan_perubahan`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_perubahan_krs` (`krs_id`),
  ADD KEY `fk_perubahan_dosen` (`diproses_oleh`);

--
-- Indexes for table `permintaan_perubahan_detail`
--
ALTER TABLE `permintaan_perubahan_detail`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_permintaan_kelas` (`permintaan_id`,`kelas_id`),
  ADD KEY `fk_detail_permintaan_kelas` (`kelas_id`);

--
-- Indexes for table `prasyarat_node`
--
ALTER TABLE `prasyarat_node`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_prasyarat_mk` (`kode_mk`),
  ADD KEY `fk_prasyarat_syarat` (`kode_mk_syarat`),
  ADD KEY `idx_prasyarat_parent` (`parent_id`);

--
-- Indexes for table `riwayat_nilai`
--
ALTER TABLE `riwayat_nilai`
  ADD PRIMARY KEY (`nim`,`kode_mk`,`semester_id`),
  ADD KEY `fk_nilai_matkul` (`kode_mk`),
  ADD KEY `fk_nilai_semester` (`semester_id`),
  ADD KEY `fk_nilai_skala` (`nilai_huruf`);

--
-- Indexes for table `semester`
--
ALTER TABLE `semester`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_semester` (`tahun_akademik`,`jenis`);

--
-- Indexes for table `skala_nilai`
--
ALTER TABLE `skala_nilai`
  ADD PRIMARY KEY (`nilai_huruf`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `daftar_tunggu`
--
ALTER TABLE `daftar_tunggu`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `dosen_wali`
--
ALTER TABLE `dosen_wali`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `jadwal_kelas`
--
ALTER TABLE `jadwal_kelas`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `kelas`
--
ALTER TABLE `kelas`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `krs`
--
ALTER TABLE `krs`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `krs_detail`
--
ALTER TABLE `krs_detail`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `log_audit`
--
ALTER TABLE `log_audit`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `notifikasi`
--
ALTER TABLE `notifikasi`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `periode_krs`
--
ALTER TABLE `periode_krs`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `permintaan_perubahan`
--
ALTER TABLE `permintaan_perubahan`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `permintaan_perubahan_detail`
--
ALTER TABLE `permintaan_perubahan_detail`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `prasyarat_node`
--
ALTER TABLE `prasyarat_node`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `semester`
--
ALTER TABLE `semester`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `daftar_tunggu`
--
ALTER TABLE `daftar_tunggu`
  ADD CONSTRAINT `fk_tunggu_kelas` FOREIGN KEY (`kelas_id`) REFERENCES `kelas` (`id`),
  ADD CONSTRAINT `fk_tunggu_mahasiswa` FOREIGN KEY (`nim`) REFERENCES `mahasiswa` (`nim`);

--
-- Constraints for table `jadwal_kelas`
--
ALTER TABLE `jadwal_kelas`
  ADD CONSTRAINT `fk_jadwal_kelas` FOREIGN KEY (`kelas_id`) REFERENCES `kelas` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `kelas`
--
ALTER TABLE `kelas`
  ADD CONSTRAINT `fk_kelas_mk` FOREIGN KEY (`kode_mk`) REFERENCES `mata_kuliah` (`kode_mk`),
  ADD CONSTRAINT `fk_kelas_semester` FOREIGN KEY (`semester_id`) REFERENCES `semester` (`id`);

--
-- Constraints for table `kesetaraan_mk`
--
ALTER TABLE `kesetaraan_mk`
  ADD CONSTRAINT `kesetaraan_mk_ibfk_1` FOREIGN KEY (`mk_lama`) REFERENCES `mata_kuliah` (`kode_mk`),
  ADD CONSTRAINT `kesetaraan_mk_ibfk_2` FOREIGN KEY (`mk_baru`) REFERENCES `mata_kuliah` (`kode_mk`);

--
-- Constraints for table `krs`
--
ALTER TABLE `krs`
  ADD CONSTRAINT `fk_krs_dosen` FOREIGN KEY (`dosen_wali_id`) REFERENCES `dosen_wali` (`id`),
  ADD CONSTRAINT `fk_krs_mahasiswa` FOREIGN KEY (`nim`) REFERENCES `mahasiswa` (`nim`),
  ADD CONSTRAINT `fk_krs_semester` FOREIGN KEY (`semester_id`) REFERENCES `semester` (`id`);

--
-- Constraints for table `krs_detail`
--
ALTER TABLE `krs_detail`
  ADD CONSTRAINT `fk_detail_kelas` FOREIGN KEY (`kelas_id`) REFERENCES `kelas` (`id`),
  ADD CONSTRAINT `fk_detail_krs` FOREIGN KEY (`krs_id`) REFERENCES `krs` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `mahasiswa`
--
ALTER TABLE `mahasiswa`
  ADD CONSTRAINT `fk_mahasiswa_dosen` FOREIGN KEY (`dosen_wali_id`) REFERENCES `dosen_wali` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

--
-- Constraints for table `periode_krs`
--
ALTER TABLE `periode_krs`
  ADD CONSTRAINT `fk_periode_semester` FOREIGN KEY (`semester_id`) REFERENCES `semester` (`id`);

--
-- Constraints for table `permintaan_perubahan`
--
ALTER TABLE `permintaan_perubahan`
  ADD CONSTRAINT `fk_perubahan_dosen` FOREIGN KEY (`diproses_oleh`) REFERENCES `dosen_wali` (`id`),
  ADD CONSTRAINT `fk_perubahan_krs` FOREIGN KEY (`krs_id`) REFERENCES `krs` (`id`);

--
-- Constraints for table `permintaan_perubahan_detail`
--
ALTER TABLE `permintaan_perubahan_detail`
  ADD CONSTRAINT `fk_detail_permintaan` FOREIGN KEY (`permintaan_id`) REFERENCES `permintaan_perubahan` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_detail_permintaan_kelas` FOREIGN KEY (`kelas_id`) REFERENCES `kelas` (`id`);

--
-- Constraints for table `prasyarat_node`
--
ALTER TABLE `prasyarat_node`
  ADD CONSTRAINT `fk_prasyarat_mk` FOREIGN KEY (`kode_mk`) REFERENCES `mata_kuliah` (`kode_mk`),
  ADD CONSTRAINT `fk_prasyarat_parent` FOREIGN KEY (`parent_id`) REFERENCES `prasyarat_node` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_prasyarat_syarat` FOREIGN KEY (`kode_mk_syarat`) REFERENCES `mata_kuliah` (`kode_mk`);

--
-- Constraints for table `riwayat_nilai`
--
ALTER TABLE `riwayat_nilai`
  ADD CONSTRAINT `fk_nilai_mahasiswa` FOREIGN KEY (`nim`) REFERENCES `mahasiswa` (`nim`),
  ADD CONSTRAINT `fk_nilai_matkul` FOREIGN KEY (`kode_mk`) REFERENCES `mata_kuliah` (`kode_mk`),
  ADD CONSTRAINT `fk_nilai_semester` FOREIGN KEY (`semester_id`) REFERENCES `semester` (`id`),
  ADD CONSTRAINT `fk_nilai_skala` FOREIGN KEY (`nilai_huruf`) REFERENCES `skala_nilai` (`nilai_huruf`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
