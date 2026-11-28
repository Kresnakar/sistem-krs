-- Akun dummy untuk pengujian Sistem KRS.
-- Jalankan setelah:
-- 1. migrations/001_initial_schema.sql
-- 2. migrations/002_auth_users.sql
-- 3. seed/seed.sql
--
-- Password disimpan sebagai bcrypt hash, bukan plaintext.
-- Mapping akun Mahasiswa menggunakan NIM 230001.
-- Mapping akun Dosen Wali menggunakan dosen_wali.id = 1.

INSERT INTO users (
    username,
    password_hash,
    role,
    nim,
    dosen_wali_id,
    is_active
) VALUES
(
    'mahasiswa01',
    '$2a$10$R2BJlJvp79Kr1l088pfeQunMv5FYnEFDpzmZ0EMnkM8WIOqhECiQ.',
    'Mahasiswa',
    '230001',
    NULL,
    1
),
(
    'dosen01',
    '$2a$10$rXvxfTI1aJSCM7YYwkRC/OvqTqHdx3PnfNqODZIrkyvyjGNov3ati',
    'Dosen Wali',
    NULL,
    1,
    1
),
(
    'admin01',
    '$2a$10$GbNuxblaErcCR9Z7OoD5AOKIPf3fMQPOANOwoNJFv9VtGKuW5/uRi',
    'Admin Prodi',
    NULL,
    NULL,
    1
),
(
    'kaprodi01',
    '$2a$10$UtQ7Hs2c38n8kiXvXGYyHO7KCEHMyhdrIzJ2CLfoZwLjhBxdPemQG',
    'Kaprodi',
    NULL,
    NULL,
    1
)
ON DUPLICATE KEY UPDATE
    password_hash = VALUES(password_hash),
    role = VALUES(role),
    nim = VALUES(nim),
    dosen_wali_id = VALUES(dosen_wali_id),
    is_active = VALUES(is_active);

