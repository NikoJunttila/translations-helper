-- +goose Up
-- Add secret_key column to projects table to allow owner recovery
ALTER TABLE projects ADD COLUMN secret_key TEXT;

-- +goose Down
ALTER TABLE projects DROP COLUMN secret_key;
