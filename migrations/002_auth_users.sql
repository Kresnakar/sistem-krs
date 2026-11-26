CREATE TABLE IF NOT EXISTS users (
    id INT NOT NULL AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM(
        'Mahasiswa',
        'Dosen Wali',
        'Admin Prodi',
        'Kaprodi'
    ) NOT NULL,
    nim VARCHAR(20) DEFAULT NULL,
    dosen_wali_id INT DEFAULT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_users_username (username),
    UNIQUE KEY uq_users_nim (nim),
    UNIQUE KEY uq_users_dosen_wali (dosen_wali_id),

    CONSTRAINT fk_users_mahasiswa
        FOREIGN KEY (nim) REFERENCES mahasiswa (nim),

    CONSTRAINT fk_users_dosen_wali
        FOREIGN KEY (dosen_wali_id) REFERENCES dosen_wali (id),

    CONSTRAINT chk_users_role_identity CHECK (
        (role = 'Mahasiswa' AND nim IS NOT NULL AND dosen_wali_id IS NULL)
        OR
        (role = 'Dosen Wali' AND nim IS NULL AND dosen_wali_id IS NOT NULL)
        OR
        (role IN ('Admin Prodi', 'Kaprodi')
            AND nim IS NULL AND dosen_wali_id IS NULL)
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;