package database

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"templui/internal/models"
)

// CreateProject creates a new project in the database
func (db *DB) CreateProject(project *models.Project) error {
	query := `
		INSERT INTO projects (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, project.ID, project.Name, project.CreatedAt, project.UpdatedAt)
	return err
}

// GetProject retrieves a project by ID
func (db *DB) GetProject(id string) (*models.Project, error) {
	query := `SELECT id, name, created_at, updated_at FROM projects WHERE id = ?`
	row := db.conn.QueryRow(query, id)

	var project models.Project
	err := row.Scan(&project.ID, &project.Name, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// ListProjects retrieves all projects
func (db *DB) ListProjects(limit int) ([]models.Project, error) {
	query := `SELECT id, name, created_at, updated_at FROM projects ORDER BY created_at DESC LIMIT ?`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// CreateFile creates a new translation file
func (db *DB) CreateFile(file *models.TranslationFile) error {
	query := `
		INSERT INTO files (id, project_id, file_type, language_code, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, file.ID, file.ProjectID, file.FileType, file.LanguageCode, file.Content, file.CreatedAt, file.UpdatedAt)
	return err
}

// GetFile retrieves a file by ID
func (db *DB) GetFile(id string) (*models.TranslationFile, error) {
	query := `SELECT id, project_id, file_type, language_code, content, created_at, updated_at FROM files WHERE id = ?`
	row := db.conn.QueryRow(query, id)

	var file models.TranslationFile
	err := row.Scan(&file.ID, &file.ProjectID, &file.FileType, &file.LanguageCode, &file.Content, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Parse content into ParsedData
	if err := json.Unmarshal([]byte(file.Content), &file.ParsedData); err == nil {
		// Successfully parsed
	}

	return &file, nil
}

// GetFilesByProject retrieves all files for a project
func (db *DB) GetFilesByProject(projectID string) ([]models.TranslationFile, error) {
	query := `SELECT id, project_id, file_type, language_code, content, created_at, updated_at FROM files WHERE project_id = ? ORDER BY file_type`
	rows, err := db.conn.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.TranslationFile
	for rows.Next() {
		var file models.TranslationFile
		if err := rows.Scan(&file.ID, &file.ProjectID, &file.FileType, &file.LanguageCode, &file.Content, &file.CreatedAt, &file.UpdatedAt); err != nil {
			return nil, err
		}
		// Parse content
		json.Unmarshal([]byte(file.Content), &file.ParsedData)
		files = append(files, file)
	}

	return files, nil
}

// UpdateFile updates a file's content
func (db *DB) UpdateFile(id, content string) error {
	query := `UPDATE files SET content = ?, updated_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, content, time.Now(), id)
	return err
}

// CreateAPIKey creates a new API key
func (db *DB) CreateAPIKey(apiKey *models.APIKey) error {
	perms := strings.Join(apiKey.Permissions, ",")
	query := `
		INSERT INTO api_keys (id, project_id, key_hash, permissions, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, apiKey.ID, apiKey.ProjectID, apiKey.KeyHash, perms, apiKey.ExpiresAt, apiKey.CreatedAt)
	return err
}

// GetAPIKeyByHash retrieves an API key by its hash
func (db *DB) GetAPIKeyByHash(keyHash string) (*models.APIKey, error) {
	query := `SELECT id, project_id, key_hash, permissions, expires_at, created_at FROM api_keys WHERE key_hash = ?`
	row := db.conn.QueryRow(query, keyHash)

	var apiKey models.APIKey
	var perms string
	var expiresAt sql.NullTime

	err := row.Scan(&apiKey.ID, &apiKey.ProjectID, &apiKey.KeyHash, &perms, &expiresAt, &apiKey.CreatedAt)
	if err != nil {
		return nil, err
	}

	apiKey.Permissions = strings.Split(perms, ",")
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}

	return &apiKey, nil
}
