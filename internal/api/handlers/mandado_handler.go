package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/joaofilippe/inti/config"
	"github.com/joaofilippe/inti/internal/api/dto"
	"github.com/joaofilippe/inti/internal/application/service"
	"github.com/joaofilippe/inti/internal/document"
)

type MandadoHandler struct {
	cfg config.Config
	svc *service.MandadoService
}

func NewMandadoHandler(cfg config.Config, svc *service.MandadoService) *MandadoHandler {
	return &MandadoHandler{cfg: cfg, svc: svc}
}

func (h *MandadoHandler) templatePath(situacao string) string {
	upper := strings.ToUpper(situacao)
	if strings.Contains(upper, "NEG") || situacao == "6" {
		return h.cfg.DocxNegTemplatePath
	}
	return h.cfg.DocxTemplatePath
}

// GerarLote gera um lote de mandados em formato ZIP.
//
//	@Summary      Gera lote de mandados
//	@Description  Recebe uma lista de mandados e retorna um arquivo ZIP contendo os documentos DOCX gerados.
//	@Tags         mandados
//	@Accept       json
//	@Produce      application/zip
//	@Param        mandados  body      []dto.MandadoPositivoDTO  true  "Lista de mandados"
//	@Success      200       {file}    binary                    "Arquivo ZIP com os mandados"
//	@Failure      400       {object}  map[string]string         "Payload inválido"
//	@Failure      500       {object}  map[string]string         "Erro interno"
//	@Router       /api/batch/mandados [post]
func (h *MandadoHandler) GerarLote(c echo.Context) error {
	var payloads []dto.MandadoPositivoDTO
	if err := c.Bind(&payloads); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Payload inválido"})
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, payload := range payloads {
		mandado := h.svc.MapDTOToDomain(payload)
		replaces := h.svc.BuildReplaces(mandado)

		docBuf, err := document.ReplaceInDocx(h.templatePath(mandado.Situacao), replaces)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao compilar documento: " + err.Error()})
		}

		fileName := fmt.Sprintf("%s-%s.docx", payload.Nome, payload.Mandado)
		fw, err := zw.Create(fileName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao criar arquivo no zip"})
		}
		if _, err := fw.Write(docBuf.Bytes()); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao escrever no zip"})
		}
	}

	if err := zw.Close(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao fechar zip"})
	}

	c.Response().Header().Set("Content-Disposition", `attachment; filename="mandados_lote.zip"`)
	return c.Blob(http.StatusOK, "application/zip", buf.Bytes())
}

// GerarMandado gera um único mandado em formato DOCX.
//
//	@Summary      Gera mandado
//	@Description  Recebe os dados de um mandado e retorna o documento DOCX gerado.
//	@Tags         mandados
//	@Accept       json
//	@Produce      application/vnd.openxmlformats-officedocument.wordprocessingml.document
//	@Param        mandado  body      dto.MandadoPositivoDTO  true  "Dados do mandado"
//	@Success      200      {file}    binary                  "Arquivo DOCX do mandado"
//	@Failure      400      {object}  map[string]string       "Payload inválido"
//	@Failure      500      {object}  map[string]string       "Erro interno"
//	@Router       /api/mandados [post]
func (h *MandadoHandler) GerarMandado(c echo.Context) error {
	var payload dto.MandadoPositivoDTO
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Payload inválido"})
	}

	mandado := h.svc.MapDTOToDomain(payload)
	replaces := h.svc.BuildReplaces(mandado)

	buf, err := document.ReplaceInDocx(h.templatePath(mandado.Situacao), replaces)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao compilar documento: " + err.Error()})
	}

	fileName := fmt.Sprintf("%s-%s.docx", payload.Nome, payload.Mandado)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", buf.Bytes())
}
