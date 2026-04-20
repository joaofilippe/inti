package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/joaofilippe/inti/internal/domain/entities"
)

type TipoAtoRepository struct {
	db *sqlx.DB
}

func NewTipoAtoRepository(db *sqlx.DB) *TipoAtoRepository {
	return &TipoAtoRepository{db: db}
}

type tipoAtoDB struct {
	Codigo    int            `db:"codigo"`
	Descricao string         `db:"descricao"`
	Positivo  sql.NullString `db:"positivo"`
	Negativo  sql.NullString `db:"negativo"`
}

func (r *TipoAtoRepository) CarregarTodos(ctx context.Context) (map[int]entities.TipoAto, error) {
	var rows []tipoAtoDB
	if err := r.db.SelectContext(ctx, &rows, `SELECT codigo, descricao, positivo, negativo FROM tipos_ato`); err != nil {
		return nil, fmt.Errorf("erro ao carregar tipos de ato: %w", err)
	}

	m := make(map[int]entities.TipoAto, len(rows))
	for _, row := range rows {
		m[row.Codigo] = entities.TipoAto{
			Codigo:    row.Codigo,
			Descricao: row.Descricao,
			Positivo:  row.Positivo.String,
			Negativo:  row.Negativo.String,
		}
	}
	return m, nil
}
