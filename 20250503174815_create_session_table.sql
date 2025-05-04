-- +goose Up
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    guru_id INTEGER REFERENCES gurus(id) ON DELETE CASCADE,
    admin_id INTEGER REFERENCES admins(id) ON DELETE CASCADE,
    wali_id INTEGER REFERENCES gurus(id) ON DELETE CASCADE, 
    token VARCHAR(512) UNIQUE NOT NULL,
    role VARCHAR(20) CHECK (role IN ('guru', 'wali_kelas', 'admin')) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- +goose Down
DROP TABLE sessions;
