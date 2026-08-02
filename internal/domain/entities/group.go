package entities

import "time"

type Group struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserGroup struct {
	UserID  string `json:"user_id" db:"user_id"`
	GroupID string `json:"group_id" db:"group_id"`
	Role    string `json:"role" db:"role"`
}
