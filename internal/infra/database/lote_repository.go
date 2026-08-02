package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/joaofilippe/inti/internal/domain/entities"
)

type LoteRepository struct {
	db *sqlx.DB
}

func NewLoteRepository(db *sqlx.DB) *LoteRepository {
	return &LoteRepository{db: db}
}

// FindOrCreateLote busca um lote pelo nome. Se não existir, cria-o com o admin_user_id fornecido.
func (r *LoteRepository) FindOrCreateLote(ctx context.Context, nomeLote string, adminUserID string) (*entities.Lote, error) {
	var lote entities.Lote
	query := `SELECT id, nome_lote, admin_user_id, group_id, created_at, updated_at FROM lotes WHERE nome_lote = $1`
	err := r.db.GetContext(ctx, &lote, query, nomeLote)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Não existe, vamos criar
			insertQuery := `
				INSERT INTO lotes (nome_lote, admin_user_id)
				VALUES ($1, $2)
				RETURNING id, nome_lote, admin_user_id, group_id, created_at, updated_at
			`
			err = r.db.QueryRowxContext(ctx, insertQuery, nomeLote, adminUserID).StructScan(&lote)
			if err != nil {
				return nil, err
			}
			return &lote, nil
		}
		return nil, err
	}

	// Verificar permissão
	if lote.AdminUserID != adminUserID {
		// Verificar se está no grupo (se houver group_id)
		if lote.GroupID != nil {
			var count int
			err = r.db.GetContext(ctx, &count, `SELECT count(*) FROM user_groups WHERE group_id = $1 AND user_id = $2`, *lote.GroupID, adminUserID)
			if err == nil && count > 0 {
				return &lote, nil
			}
		}
		return nil, errors.New("usuário não tem permissão para acessar este lote")
	}

	return &lote, nil
}
