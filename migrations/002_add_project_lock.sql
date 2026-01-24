-- +goose Up
ALTER TABLE projects ADD COLUMN is_locked BOOLEAN DEFAULT FALSE NOT NULL;
ALTER TABLE projects ADD COLUMN secret_key_hash TEXT;

-- +goose Down
ALTER TABLE projects DROP COLUMN secret_key_hash;
ALTER TABLE projects DROP COLUMN is_locked;
