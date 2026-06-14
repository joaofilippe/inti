package service

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/joaofilippe/inti/internal/api/dto"
	"github.com/joaofilippe/inti/internal/domain/entities"
)

type MandadoService struct {
	tiposAto map[int]entities.TipoAto
	motivos  map[int]entities.MotivoNaoRealizacao
}

func NewMandadoService(tiposAto map[int]entities.TipoAto, motivos map[int]entities.MotivoNaoRealizacao) *MandadoService {
	return &MandadoService{tiposAto: tiposAto, motivos: motivos}
}

// MapDTOToDomain converte o DTO recebido do frontend para o tipo de domínio.
func (s *MandadoService) MapDTOToDomain(p dto.MandadoPositivoDTO) entities.Mandado {
	atos := make([]entities.Ato, len(p.Atos))
	for i, a := range p.Atos {
		atos[i] = entities.Ato{
			CodigoAto:             a.CodigoAto,
			DataCumprimento:       a.DataCumprimento,
			Horario:               a.Horario,
			Realizado:             a.Realizado,
			MotivoNaoRealizacaoID: a.MotivoNaoRealizacaoID,
		}
	}

	contatos := make([]entities.Contato, len(p.Contato))
	for i, c := range p.Contato {
		contatos[i] = entities.Contato{
			Tipo:            entities.TipoContato(c.Tipo),
			Valor:           c.Valor,
			IsTerceiro:      c.IsTerceiro,
			NomeTerceiro:    c.NomeTerceiro,
			RelacaoTerceiro: c.RelacaoTerceiro,
		}
	}

	return entities.Mandado{
		Mandado:           p.Mandado,
		Lote:              p.Lote,
		Situacao:          p.Situacao,
		DataCarga:         p.DataCarga,
		Atos:              atos,
		Diligencias:       p.Diligencias,
		IsPJ:              p.IsPJ,
		Contatos:          contatos,
		Cidade:            p.Cidade,
		Endereco:          normalizarMaiusculas(p.Endereco),
		Nome:              normalizarMaiusculas(p.Nome),
		Sexo:              p.Sexo,
		Posicao:           p.Posicao,
		Documento:             p.Documento,
		TipoDocumento:         strings.ToUpper(p.TipoDocumento),
		RepresentanteNome:     p.RepresentanteNome,
		RepresentanteDoc:      p.RepresentanteDoc,
		Obs:                   p.Obs,
		MotivoNaoRealizacaoID: p.MotivoNaoRealizacaoID,
	}
}

// BuildReplaces constrói o mapa de tags → valores para substituição no template Word.
func (s *MandadoService) BuildReplaces(m entities.Mandado) map[string]string {
	var contatosStr []string
	for _, ct := range m.Contatos {
		if ct.IsTerceiro {
			contatosStr = append(contatosStr, fmt.Sprintf("%s (%s - %s)", ct.Valor, ct.NomeTerceiro, ct.RelacaoTerceiro))
		} else {
			contatosStr = append(contatosStr, ct.Valor)
		}
	}

	positionText := strings.ToLower(m.Posicao)
	nomeText := m.Nome
	if m.Sexo == "F" {
		positionText = "a " + positionText
		nomeText = "a Sra. " + nomeText
	} else {
		positionText = "o " + positionText
		nomeText = "o Sr. " + nomeText
	}

	dataCargaCurta := m.DataCarga
	if len(dataCargaCurta) >= 5 {
		dataCargaCurta = dataCargaCurta[:5]
	} else {
		dataCargaCurta = "*"
	}

	type atoTexto struct {
		positivo  string
		negativo  string
		motivo    string
		realizado bool
	}

	var datasExtensas, horarios, dataHorarios []string
	var atosTextos []atoTexto

	for _, ato := range m.Atos {
		datasExtensas = append(datasExtensas, formatarData(ato.DataCumprimento))

		horarioF := strings.Replace(ato.Horario, ":", "h", 1)
		horarios = append(horarios, horarioF)

		dC := ato.DataCumprimento
		if len(dC) >= 5 {
			dC = dC[:5]
		}
		dhCombinado := strings.TrimSpace(fmt.Sprintf("%s %s", dC, horarioF))
		if dhCombinado == "" {
			dhCombinado = "*"
		}
		dataHorarios = append(dataHorarios, dhCombinado)

		if tipo, ok := s.tiposAto[ato.CodigoAto]; ok {
			isRealizado := m.Situacao != "6"
			if ato.Realizado != nil {
				isRealizado = *ato.Realizado
			}

			var motivoStr string
			if ato.MotivoNaoRealizacaoID != nil {
				if mtv, ok := s.motivos[*ato.MotivoNaoRealizacaoID]; ok {
					motivoStr = mtv.Explicacao
				}
			}

			atosTextos = append(atosTextos, atoTexto{
				positivo:  tipo.Positivo,
				negativo:  tipo.Negativo,
				motivo:    motivoStr,
				realizado: isRealizado,
			})
		}
	}

	var realizados, naoRealizados []atoTexto
	for _, at := range atosTextos {
		if at.realizado {
			realizados = append(realizados, at)
		} else {
			naoRealizados = append(naoRealizados, at)
		}
	}

	var atoStr string
	switch {
	case len(naoRealizados) == 0:
		var names []string
		for _, r := range realizados {
			names = append(names, r.positivo)
		}
		atoStr = joinWithE(names)
	case len(realizados) == 0:
		var names []string
		for _, nr := range naoRealizados {
			names = append(names, nr.negativo)
		}
		atoStr = joinWithE(names)
	default:
		var posNames []string
		for _, r := range realizados {
			posNames = append(posNames, r.positivo)
		}
		parts := []string{joinWithE(posNames)}
		for _, nr := range naoRealizados {
			cert := "Certifico mais, que não pude realizar a " + nr.negativo
			if nr.motivo != "" {
				cert += " pois " + nr.motivo
			}
			parts = append(parts, cert)
		}
		atoStr = strings.Join(parts, ". ")
	}

	drStr := strings.TrimSpace(m.Diligencias)
	if drStr != "" {
		drStr = " - " + drStr
	}

	obsStr := strings.TrimSpace(m.Obs)
	if obsStr != "" {
		obsStr = " - " + obsStr
	}

	contatoStr := strings.Join(contatosStr, " , ")
	if strings.TrimSpace(contatoStr) != "" {
		contatoStr = " - " + contatoStr
	}

	replaces := map[string]string{
		"{{MANDADO}}":     extrairNumeroMandado(m.Mandado),
		"{{CARGA}}":       dataCargaCurta,
		"{{DATA}}":        joinWithE(datasExtensas),
		"{{HORARIO}}":     joinWithE(horarios),
		"{{DATAHORARIO}}": joinWithE(dataHorarios),
		"{{ATO}}":         atoStr,
		"{{CIDADE}}":      m.Cidade,
		"{{ENDERECO}}":    m.Endereco,
		"{{ENDEREÇO}}":    m.Endereco,
		"{{NOME}}":        nomeText,
		"{{POSICAO}}":     positionText,
		"{{CPF}}":         documentoFormatado(m.TipoDocumento, m.Documento),
		"{{CONTATO}}":     contatoStr,
		"{{OBS}}":         drStr + obsStr,
	}

	if m.IsPJ {
		replaces["{{NOME}}"] = fmt.Sprintf("%s, na pessoa de seu representante legal %s", m.Nome, m.RepresentanteNome)
		replaces["{{CPF}}"] = fmt.Sprintf("%s CPF %s", m.Documento, m.RepresentanteDoc)
	}

	return replaces
}

func formatarData(dataStr string) string {
	partes := strings.Split(dataStr, "/")
	if len(partes) == 3 {
		if mesExtenso, ok := entities.Meses[partes[1]]; ok {
			return fmt.Sprintf("%s %s %s", partes[0], mesExtenso, partes[2])
		}
	}
	return dataStr
}

func extrairNumeroMandado(mandado string) string {
	idx := strings.Index(mandado, "/")
	if idx == -1 {
		return mandado
	}
	sufixo := mandado[idx+1:]
	partes := strings.SplitN(sufixo, "-", 2)
	numero := strings.TrimLeft(partes[0], "0")
	if numero == "" {
		numero = "0"
	}
	if len(partes) == 2 {
		return numero + "-" + partes[1]
	}
	return numero
}

func joinWithE(arr []string) string {
	switch len(arr) {
	case 0:
		return ""
	case 1:
		return arr[0]
	default:
		return strings.Join(arr[:len(arr)-1], ", ") + " e " + arr[len(arr)-1]
	}
}

var palavrasMinusculas = map[string]bool{
	"n.": true, "n°": true, "nº": true,
	"bairro": true,
	"de": true, "da": true, "do": true, "dos": true, "das": true,
	"e": true,
}

func documentoFormatado(tipo, numero string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, numero)

	var formatted string
	switch {
	case len(digits) == 11: // CPF: 123.456.789-01
		formatted = digits[:3] + "." + digits[3:6] + "." + digits[6:9] + "-" + digits[9:]
	case len(digits) == 14: // CNPJ: 12.345.678/0001-90
		formatted = digits[:2] + "." + digits[2:5] + "." + digits[5:8] + "/" + digits[8:12] + "-" + digits[12:]
	case len(digits) >= 7 && len(digits) <= 9: // RG
		switch len(digits) {
		case 9: // 12.345.678-9 (com dígito verificador)
			formatted = digits[:2] + "." + digits[2:5] + "." + digits[5:8] + "-" + digits[8:]
		case 8: // 12.345.678 (sem dígito verificador)
			formatted = digits[:2] + "." + digits[2:5] + "." + digits[5:8]
		case 7: // 1.234.456 (RG antigo)
			formatted = digits[:1] + "." + digits[1:4] + "." + digits[4:7]
		}
	default:
		formatted = numero
	}

	if tipo == "" {
		return formatted
	}
	return tipo + " " + formatted
}

func normalizarMaiusculas(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || trimmed != strings.ToUpper(trimmed) {
		return s
	}

	words := strings.Fields(strings.ToLower(trimmed))
	for i, w := range words {
		key := strings.TrimSuffix(w, ",")
		if !palavrasMinusculas[key] {
			r := []rune(w)
			r[0] = unicode.ToUpper(r[0])
			words[i] = string(r)
		}
	}
	return strings.Join(words, " ")
}
