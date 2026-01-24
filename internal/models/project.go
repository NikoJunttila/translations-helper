package models

import "time"

type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	IsLocked      bool      `json:"is_locked"`
	SecretKeyHash string    `json:"-"` // Never expose hash to client
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
