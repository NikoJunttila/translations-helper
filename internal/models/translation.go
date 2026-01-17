package models

import "time"

type TranslationFile struct {
	ID           string                 `json:"id"`
	ProjectID    string                 `json:"project_id"`
	FileType     string                 `json:"file_type"` // "base" or "target"
	LanguageCode string                 `json:"language_code"`
	Content      string                 `json:"content"` // JSON as string
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ParsedData   map[string]interface{} `json:"-"` // In-memory only
}
