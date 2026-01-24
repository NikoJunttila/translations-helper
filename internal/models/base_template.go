package models

import "time"

type BaseTemplate struct {
	ID           string     `json:"id"`
	SessionToken string     `json:"-"` // Don't expose to client
	Name         string     `json:"name"`
	LanguageCode string     `json:"language_code"`
	Content      string     `json:"content"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}
