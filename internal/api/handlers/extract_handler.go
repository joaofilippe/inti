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

// ListarResumo retorna uma lista com nome, mandado e mandado abreviado dos mandados extraidos.
//
//	@Summary      Listar resumo de mandados
//	@Description  Retorna uma lista resumida dos mandados extraídos, com a opção de filtrar por lote.
//	@Tags         mandados
//	@Produce      json
//	@Param        lote  query     string  false  "Nome do lote para filtrar"
//	@Success      200   {array}   dto.MandadoResumoDTO
//	@Failure      500   {object}  map[string]string
//	@Router       /api/mandados/resumo [get]
func (h *ExtractHandler) ListarResumo(c echo.Context) error {
	lote := c.QueryParam("lote")
	resumo, err := h.svc.ListarResumo(c.Request().Context(), lote)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resumo)
}

// ExtrairDeExcel processa um arquivo Excel para pré-cadastro.
//
//	@Summary      Pré-cadastro via Excel
//	@Description  Recebe um arquivo Excel (.xlsx) e cadastra os mandados.
//	@Tags         extração
//	@Accept       multipart/form-data
//	@Produce      json
//	@Param        file  formData  file                      true  "Arquivo XLSX"
//	@Success      200   {array}   dto.MandadoExtraido       "Lista de mandados pré-cadastrados"
//	@Failure      400   {object}  map[string]string         "Arquivo inválido ou não enviado"
//	@Failure      500   {object}  map[string]string         "Erro interno"
//	@Router       /api/batch/pre-cadastro/excel [post]
func (h *ExtractHandler) ExtrairDeExcel(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Arquivo inválido ou não enviado"})
	}

	data, err := lerArquivo(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	dados, err := h.svc.ExtrairDeExcel(c.Request().Context(), data, file.Filename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, dados)
}
