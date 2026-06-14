package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/joaofilippe/inti/internal/domain/entities"
)

type MotivoNaoRealizacaoRepository struct {
	db *sqlx.DB
}

func NewMotivoNaoRealizacaoRepository(db *sqlx.DB) *MotivoNaoRealizacaoRepository {
	return &MotivoNaoRealizacaoRepository{db: db}
}

type motivoNaoRealizacaoDB struct {
	ID         int    `db:"id"`
	Codigo     string `db:"codigo"`
	Explicacao string `db:"explicacao"`
}

func (r *MotivoNaoRealizacaoRepository) CarregarTodos(ctx context.Context) ([]entities.MotivoNaoRealizacao, error) {
	var rows []motivoNaoRealizacaoDB
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, codigo, explicacao FROM motivos_nao_realizacao ORDER BY explicacao ASC`); err != nil {
		return nil, fmt.Errorf("erro ao carregar motivos de nao realizacao: %w", err)
	}

	result := make([]entities.MotivoNaoRealizacao, 0, len(rows))
	for _, row := range rows {
		result = append(result, entities.MotivoNaoRealizacao{
			ID:         row.ID,
			Codigo:     row.Codigo,
			Explicacao: row.Explicacao,
		})
	}
	return result, nil
}
