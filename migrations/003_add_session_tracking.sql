-- +goose Up
-- Add session tracking to projects
ALTER TABLE projects ADD COLUMN session_token TEXT;
CREATE INDEX idx_projects_session ON projects(session_token);

-- Create base_templates table for reusable base files
CREATE TABLE base_templates (
    id TEXT PRIMARY KEY,
    session_token TEXT NOT NULL,
    name TEXT NOT NULL,
    language_code TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP
);
CREATE INDEX idx_templates_session ON base_templates(session_token);

-- +goose Down
DROP INDEX idx_templates_session;
DROP TABLE base_templates;
DROP INDEX idx_projects_session;
ALTER TABLE projects DROP COLUMN session_token;
