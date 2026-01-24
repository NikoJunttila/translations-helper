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
		INSERT INTO projects (id, name, is_locked, secret_key_hash, session_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, project.ID, project.Name, project.IsLocked, project.SecretKeyHash, project.SessionToken, project.CreatedAt, project.UpdatedAt)
	return err
}

// GetProject retrieves a project by ID
func (db *DB) GetProject(id string) (*models.Project, error) {
	query := `SELECT id, name, is_locked, secret_key_hash, session_token, created_at, updated_at FROM projects WHERE id = ?`
	row := db.conn.QueryRow(query, id)

	var project models.Project
	var secretKeyHash sql.NullString
	var sessionToken sql.NullString
	err := row.Scan(&project.ID, &project.Name, &project.IsLocked, &secretKeyHash, &sessionToken, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if secretKeyHash.Valid {
		project.SecretKeyHash = secretKeyHash.String
	}
	if sessionToken.Valid {
		project.SessionToken = sessionToken.String
	}

	return &project, nil
}

// ListProjects retrieves all projects
func (db *DB) ListProjects(limit int) ([]models.Project, error) {
	query := `SELECT id, name, is_locked, secret_key_hash, session_token, created_at, updated_at FROM projects ORDER BY created_at DESC LIMIT ?`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		var secretKeyHash sql.NullString
		var sessionToken sql.NullString
		if err := rows.Scan(&project.ID, &project.Name, &project.IsLocked, &secretKeyHash, &sessionToken, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		if secretKeyHash.Valid {
			project.SecretKeyHash = secretKeyHash.String
		}
		if sessionToken.Valid {
			project.SessionToken = sessionToken.String
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

// GetProjectsBySession retrieves all projects for a session token
func (db *DB) GetProjectsBySession(sessionToken string, limit int) ([]models.Project, error) {
	query := `SELECT id, name, is_locked, secret_key_hash, session_token, created_at, updated_at 
	          FROM projects 
	          WHERE session_token = ? 
	          ORDER BY created_at DESC 
	          LIMIT ?`
	rows, err := db.conn.Query(query, sessionToken, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		var secretKeyHash sql.NullString
		var sessionTokenVal sql.NullString
		if err := rows.Scan(&project.ID, &project.Name, &project.IsLocked, &secretKeyHash, &sessionTokenVal, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		if secretKeyHash.Valid {
			project.SecretKeyHash = secretKeyHash.String
		}
		if sessionTokenVal.Valid {
			project.SessionToken = sessionTokenVal.String
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// CreateBaseTemplate creates a new base template
func (db *DB) CreateBaseTemplate(template *models.BaseTemplate) error {
	query := `
		INSERT INTO base_templates (id, session_token, name, language_code, content, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, template.ID, template.SessionToken, template.Name, template.LanguageCode, template.Content, template.CreatedAt, template.LastUsedAt)
	return err
}

// GetBaseTemplatesBySession retrieves all base templates for a session
func (db *DB) GetBaseTemplatesBySession(sessionToken string) ([]models.BaseTemplate, error) {
	query := `SELECT id, session_token, name, language_code, content, created_at, last_used_at 
	          FROM base_templates 
	          WHERE session_token = ? 
	          ORDER BY last_used_at DESC, created_at DESC`
	rows, err := db.conn.Query(query, sessionToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.BaseTemplate
	for rows.Next() {
		var template models.BaseTemplate
		var lastUsedAt sql.NullTime
		if err := rows.Scan(&template.ID, &template.SessionToken, &template.Name, &template.LanguageCode, &template.Content, &template.CreatedAt, &lastUsedAt); err != nil {
			return nil, err
		}
		if lastUsedAt.Valid {
			template.LastUsedAt = &lastUsedAt.Time
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// UpdateTemplateLastUsed updates the last used timestamp for a template
func (db *DB) UpdateTemplateLastUsed(id string) error {
	query := `UPDATE base_templates SET last_used_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, time.Now(), id)
	return err
}
