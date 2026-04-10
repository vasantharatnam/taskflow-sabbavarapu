package models

import "time"

type Project struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}