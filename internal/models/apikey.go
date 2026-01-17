package models

import "time"

type APIKey struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	KeyHash     string     `json:"-"` // Never expose hash
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
