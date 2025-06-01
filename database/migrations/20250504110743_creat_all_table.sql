-- +goose Up
CREATE TABLE admins (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE gurus (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    nip VARCHAR(30) UNIQUE NOT NULL,
    nik VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    telepon VARCHAR(20),
    alamat TEXT,
    jenis_kelamin ENUM('L','P') NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE kelas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(50) NOT NULL,
    tingkat ENUM('SD','SMP','SMA') NOT NULL,
    tahun_ajaran VARCHAR(9) NOT NULL,
    wali_kelas_id INT, -- case jika wali_kelas_id = NULL, maka kelas tersebut tidak memiliki wali kelas
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_kelas (nama, tingkat, tahun_ajaran),
    FOREIGN KEY (wali_kelas_id) REFERENCES gurus(id)
);

CREATE TABLE mata_pelajarans (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    kode VARCHAR(20) UNIQUE NOT NULL,
    tingkat ENUM('SD','SMP','SMA') NOT NULL DEFAULT 'SMP',
    semester ENUM('ganjil','genap'),
    hari ENUM('Senin','Selasa','Rabu','Kamis','Jumat','Sabtu') DEFAULT 'Senin',
    jam_mulai TIME,
    jam_selesai TIME,
    is_active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE guru_roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    guru_id INT NOT NULL,
    role ENUM('wali_kelas','guru_mapel') NOT NULL,
    kelas_id INT,  -- case jika role = 'wali_kelas'
    mapel_id INT,  -- case jika role = 'guru_mapel'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (guru_id) REFERENCES gurus(id) ON DELETE CASCADE,
    FOREIGN KEY (kelas_id) REFERENCES kelas(id),
    FOREIGN KEY (mapel_id) REFERENCES mata_pelajarans(id),
    UNIQUE KEY uniq_wali_kelas (role, kelas_id),
    UNIQUE KEY uniq_guru_mapel (guru_id, role, mapel_id),
    INDEX idx_guru_roles_guru (guru_id),
    INDEX idx_guru_roles_mapel (mapel_id)
);

CREATE TABLE guru_mapel_kelas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    guru_id INT NOT NULL,
    mapel_id INT NOT NULL,
    kelas_id INT NOT NULL,
    tahun_ajaran VARCHAR(9) NOT NULL, 
    semester ENUM('ganjil','genap') NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_gmk (guru_id, mapel_id, kelas_id, tahun_ajaran, semester),
    FOREIGN KEY (guru_id) REFERENCES gurus(id),
    FOREIGN KEY (mapel_id) REFERENCES mata_pelajarans(id),
    FOREIGN KEY (kelas_id) REFERENCES kelas(id),
    INDEX idx_gmk_tahun (tahun_ajaran),
    INDEX idx_gmk_semester (semester)
);

CREATE TABLE siswas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    nisn VARCHAR(20) UNIQUE NOT NULL,
    tempat_lahir VARCHAR(100) NOT NULL,
    tanggal_lahir DATE NOT NULL,
    jenis_kelamin ENUM('L','P') NOT NULL,
    nama_ayah VARCHAR(100) NOT NULL,
    nama_ibu VARCHAR(100) NOT NULL,
    alamat TEXT NOT NULL,
    agama VARCHAR(20) NOT NULL,
    email VARCHAR(100) UNIQUE,
    telepon VARCHAR(20),
    asal_sekolah VARCHAR(100) NOT NULL,
    kelas_id INT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (kelas_id) REFERENCES kelas(id)
);

CREATE TABLE absensi_siswas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    siswa_id INT NOT NULL,
    kelas_id INT NOT NULL,
    mapel_id INT,
    guru_id INT NOT NULL,
    tipe_absensi ENUM('kelas','mapel') NOT NULL,
    tanggal DATE NOT NULL,
    status ENUM('masuk','izin','sakit','terlambat','alpa') NOT NULL,
    keterangan TEXT,
    tahun_ajaran VARCHAR(9) NOT NULL,
    semester ENUM('ganjil','genap') NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (siswa_id) REFERENCES siswas(id) ON DELETE CASCADE,
    FOREIGN KEY (kelas_id) REFERENCES kelas(id),
    FOREIGN KEY (mapel_id) REFERENCES mata_pelajarans(id),
    FOREIGN KEY (guru_id) REFERENCES gurus(id)
);

CREATE TABLE sessions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    token VARCHAR(512) UNIQUE NOT NULL,
    role ENUM('guru','admin','wali_kelas') NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

-- +goose Down
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS absensi_siswas;
DROP TABLE IF EXISTS siswas;
DROP TABLE IF EXISTS guru_mapel_kelas;
DROP TABLE IF EXISTS mata_pelajarans;
DROP TABLE IF EXISTS kelas;
DROP TABLE IF EXISTS guru_roles;
DROP TABLE IF EXISTS gurus;
DROP TABLE IF EXISTS admins;
