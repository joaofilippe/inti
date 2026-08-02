package document

import (
	"fmt"
	"io"
	"strings"

	"github.com/joaofilippe/inti/internal/api/dto"
	"github.com/xuri/excelize/v2"
)

// ParseExcel reads an excel file and maps it to a slice of MandadoExtraido.
func ParseExcel(r io.Reader, filename string) ([]dto.MandadoExtraido, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir excel: %w", err)
	}
	defer f.Close()

	// Pega a primeira aba
	sheetMap := f.GetSheetMap()
	var firstSheet string
	for _, name := range sheetMap {
		firstSheet = name
		break
	}

	rows, err := f.GetRows(firstSheet)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler linhas da aba %s: %w", firstSheet, err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("a planilha deve ter pelo menos uma linha de cabeçalho e uma linha de dados")
	}

	// Mapear colunas (insensitive)
	header := rows[0]
	colIdx := map[string]int{}
	for i, h := range header {
		cleanH := strings.TrimSpace(strings.ToUpper(h))
		colIdx[cleanH] = i
	}

	// Helpers para achar as colunas
	getCol := func(row []string, names ...string) string {
		for _, name := range names {
			for headerName, idx := range colIdx {
				if strings.Contains(headerName, name) && idx < len(row) {
					return strings.TrimSpace(row[idx])
				}
			}
		}
		return ""
	}

	// Lote será derivado do nome do arquivo (removendo a extensão)
	lote := filename
	if idx := strings.LastIndex(lote, "."); idx != -1 {
		lote = lote[:idx]
	}

	var mandados []dto.MandadoExtraido
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		mandado := getCol(row, "MANDADO")
		// Se não tiver mandado, pode ser uma linha vazia, mas vamos pegar mesmo assim se tiver processo
		processo := getCol(row, "PROCESSO")
		
		if mandado == "" && processo == "" {
			continue
		}

		m := dto.MandadoExtraido{
			Mandado:        mandado,
			NumeroProcesso: processo,
			Lote:           lote,
			Nome:           getCol(row, "NOME", "PARTE", "DESTINATARIO"),
			Endereco:       getCol(row, "ENDERE", "ENDERECO", "LOCAL"),
			Documento:      getCol(row, "DOC", "CPF", "CNPJ"),
			Cidade:         getCol(row, "CIDADE", "MUNICIPIO"),
			// Pode tentar pegar Sexo e Posicao, mas muitas planilhas não têm
			Sexo:           getCol(row, "SEXO"),
			Posicao:        getCol(row, "POSICAO", "POLO"),
			TipoDocumento:  getCol(row, "TIPO DOC"), // Opcional, o parser de DOC depois pode inferir pelo tamanho
		}
		
		mandados = append(mandados, m)
	}

	return mandados, nil
}
