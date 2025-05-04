-- +goose Up
CREATE TABLE guru_roles (
    id SERIAL PRIMARY KEY,
    guru_id INTEGER NOT NULL REFERENCES gurus(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('wali_kelas', 'guru_mapel')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE guru_roles;
