package models

import "time"

type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	IsLocked      bool      `json:"is_locked"`
	SecretKeyHash string    `json:"-"` // Never expose hash to client
	SecretKey     string    `json:"-"` // Never expose raw key to client (except owner)
	SessionToken  string    `json:"-"` // Don't expose session token
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
