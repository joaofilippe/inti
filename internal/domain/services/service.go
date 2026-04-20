package services

import (
	"context"

	"github.com/joaofilippe/inti/internal/domain/entities"
	"github.com/joaofilippe/inti/internal/api/dto"
)

type ExtractService interface {
	ExtrairDadosMandado(ctx context.Context, data []byte, apiKey string) (dto.MandadoExtraido, error)
	ExtrairDadosLote(ctx context.Context, data []byte, apiKey string) ([]dto.MandadoExtraido, error)
	NormalizarExtraido(m *dto.MandadoExtraido)
	NormalizarLoteExtraido(items []dto.MandadoExtraido)
}

type MandadoService interface {
	MapDTOToDomain(p dto.MandadoPositivoDTO) entities.Mandado
	BuildReplaces(m entities.Mandado) map[string]string
}
