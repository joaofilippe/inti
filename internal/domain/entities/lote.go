package entities

import "time"

type Lote struct {
	ID          string    `json:"id" db:"id"`
	NomeLote    string    `json:"nome_lote" db:"nome_lote"`
	AdminUserID string    `json:"admin_user_id" db:"admin_user_id"`
	GroupID     *string   `json:"group_id" db:"group_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
