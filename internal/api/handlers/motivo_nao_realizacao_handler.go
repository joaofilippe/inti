package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/joaofilippe/inti/internal/application/repository"
	"github.com/joaofilippe/inti/internal/domain/entities"
	"github.com/joaofilippe/inti/internal/infra/cache"
)

type MotivoNaoRealizacaoHandler struct {
	cache *cache.Cache
	repo  *repository.MotivoNaoRealizacaoRepository
}

func NewMotivoNaoRealizacaoHandler(c *cache.Cache, repo *repository.MotivoNaoRealizacaoRepository) *MotivoNaoRealizacaoHandler {
	return &MotivoNaoRealizacaoHandler{cache: c, repo: repo}
}

type motivoNaoRealizacaoResponse struct {
	ID         int    `json:"id"`
	Codigo     string `json:"codigo"`
	Explicacao string `json:"explicacao"`
}

func (h *MotivoNaoRealizacaoHandler) ListarMotivosNaoRealizacao(c echo.Context) error {
	ctx := c.Request().Context()

	if cached, err := h.cache.GetMotivosNaoRealizacao(ctx); err == nil {
		return c.JSONBlob(http.StatusOK, []byte(cached))
	}

	motivos, err := h.repo.CarregarTodos(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar motivos de não realização"})
	}

	resp := motivosToResponse(motivos)

	data, err := json.Marshal(resp)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao serializar motivos de não realização"})
	}

	_ = h.cache.SetMotivosNaoRealizacao(ctx, string(data))

	return c.JSONBlob(http.StatusOK, data)
}

func motivosToResponse(m []entities.MotivoNaoRealizacao) []motivoNaoRealizacaoResponse {
	resp := make([]motivoNaoRealizacaoResponse, 0, len(m))
	for _, t := range m {
		resp = append(resp, motivoNaoRealizacaoResponse{
			ID:         t.ID,
			Codigo:     t.Codigo,
			Explicacao: t.Explicacao,
		})
	}
	return resp
}
