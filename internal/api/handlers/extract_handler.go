package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/joaofilippe/inti/internal/application/service"
	"github.com/labstack/echo/v4"
)

type ExtractHandler struct {
	svc *service.ExtractService
}

func NewExtractHandler(svc *service.ExtractService) *ExtractHandler {
	return &ExtractHandler{svc: svc}
}

// ExtrairMandado extrai dados de um único documento PDF.
//
//	@Summary      Extrai mandado
//	@Description  Recebe um arquivo PDF e extrai os dados do mandado usando IA (Gemini).
//	@Tags         extração
//	@Accept       multipart/form-data
//	@Produce      json
//	@Param        file  formData  file                    true  "Arquivo PDF do mandado"
//	@Success      200   {object}  dto.MandadoExtraido     "Dados extraídos do mandado"
//	@Failure      400   {object}  map[string]string       "Arquivo inválido ou não enviado"
//	@Failure      500   {object}  map[string]string       "Erro interno"
//	@Router       /api/extract [post]
func (h *ExtractHandler) ExtrairMandado(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Arquivo inválido ou não enviado"})
	}

	data, err := lerArquivo(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	dados, err := h.svc.ExtrairMandado(c.Request().Context(), data, file.Filename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, dados)
}

// ExtrairLote extrai dados de um PDF com múltiplos mandados.
//
//	@Summary      Extrai lote de mandados
//	@Description  Recebe um arquivo PDF com múltiplos mandados e extrai os dados de cada um usando IA (Gemini).
//	@Tags         extração
//	@Accept       multipart/form-data
//	@Produce      json
//	@Param        file  formData  file                      true  "Arquivo PDF com múltiplos mandados"
//	@Success      200   {array}   dto.MandadoExtraido       "Lista de mandados extraídos"
//	@Failure      400   {object}  map[string]string         "Arquivo inválido ou não enviado"
//	@Failure      500   {object}  map[string]string         "Erro interno"
//	@Router       /api/batch/extract [post]
func (h *ExtractHandler) ExtrairLote(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Arquivo inválido ou não enviado"})
	}

	data, err := lerArquivo(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	dados, err := h.svc.ExtrairLote(c.Request().Context(), data, file.Filename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, dados)
}

func lerArquivo(fh *multipart.FileHeader) ([]byte, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer f.Close()
	return io.ReadAll(f)
}
