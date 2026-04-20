package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/joaofilippe/inti/internal/application/repository"
	"github.com/joaofilippe/inti/internal/domain/entities"
	"github.com/joaofilippe/inti/internal/infra/cache"
)

type TipoAtoHandler struct {
	cache *cache.Cache
	repo  *repository.TipoAtoRepository
}

func NewTipoAtoHandler(c *cache.Cache, repo *repository.TipoAtoRepository) *TipoAtoHandler {
	return &TipoAtoHandler{cache: c, repo: repo}
}

type tipoAtoResponse struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
	Positivo  string `json:"positivo,omitempty"`
	Negativo  string `json:"negativo,omitempty"`
}

// ListarTiposAto retorna todos os tipos de ato disponíveis.
//
//	@Summary      Lista tipos de ato
//	@Description  Retorna a lista de tipos de ato. Os dados são servidos a partir do Redis; caso não estejam em cache, são buscados no banco de dados.
//	@Tags         tipos-ato
//	@Produce      json
//	@Success      200  {array}   tipoAtoResponse
//	@Failure      500  {object}  map[string]string  "Erro interno"
//	@Router       /api/tipos-ato [get]
func (h *TipoAtoHandler) ListarTiposAto(c echo.Context) error {
	ctx := c.Request().Context()

	if cached, err := h.cache.GetTiposAto(ctx); err == nil {
		return c.JSONBlob(http.StatusOK, []byte(cached))
	}

	tiposAto, err := h.repo.CarregarTodos(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar tipos de ato"})
	}

	resp := tiposAtoToResponse(tiposAto)

	data, err := json.Marshal(resp)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao serializar tipos de ato"})
	}

	_ = h.cache.SetTiposAto(ctx, string(data))

	return c.JSONBlob(http.StatusOK, data)
}

func tiposAtoToResponse(m map[int]entities.TipoAto) []tipoAtoResponse {
	resp := make([]tipoAtoResponse, 0, len(m))
	for _, t := range m {
		resp = append(resp, tipoAtoResponse{
			Codigo:    t.Codigo,
			Descricao: t.Descricao,
			Positivo:  t.Positivo,
			Negativo:  t.Negativo,
		})
	}
	return resp
}
