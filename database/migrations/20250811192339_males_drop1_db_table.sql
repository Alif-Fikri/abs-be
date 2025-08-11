-- -- +goose Up
-- CREATE TABLE kelas_siswas (
--     siswa_id INT NOT NULL,
--     kelas_id INT NOT NULL,
--     PRIMARY KEY (siswa_id, kelas_id),
--     FOREIGN KEY (siswa_id) REFERENCES siswas(id) ON DELETE CASCADE,
--     FOREIGN KEY (kelas_id) REFERENCES kelas(id) ON DELETE CASCADE
-- );

-- CREATE TABLE mapel_siswas (
--     siswa_id INT NOT NULL,
--     mapel_id INT NOT NULL,
--     PRIMARY KEY (siswa_id, mapel_id),
--     FOREIGN KEY (siswa_id) REFERENCES siswas(id) ON DELETE CASCADE,
--     FOREIGN KEY (mapel_id) REFERENCES mata_pelajarans(id) ON DELETE CASCADE
-- );

-- -- +goose Down
-- ALTER TABLE siswas
--   DROP COLUMN password;
