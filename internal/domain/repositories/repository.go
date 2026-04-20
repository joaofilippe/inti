package repositories

import (
	"context"

	"github.com/joaofilippe/inti/internal/api/dto"
	"github.com/joaofilippe/inti/internal/domain/entities"
)

type MandadoRepository interface {
	SalvarExtraido(ctx context.Context, m *dto.MandadoExtraido) error
	SalvarLoteExtraido(ctx context.Context, lista []dto.MandadoExtraido) error
	SalvarMandado(ctx context.Context, m *entities.Mandado) error
}

type TipoAtoRepository interface {
	CarregarTodos(ctx context.Context) (map[int]entities.TipoAto, error)
}
