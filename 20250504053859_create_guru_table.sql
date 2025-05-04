-- +goose Up
CREATE TABLE gurus (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    nip VARCHAR(30) UNIQUE NOT NULL,
    nik VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    telepon VARCHAR(20),
    alamat TEXT,
    jenis_kelamin ENUM('L', 'P') NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_wali_kelas BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE gurus;
