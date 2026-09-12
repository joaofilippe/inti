package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode"

	"google.golang.org/genai"

	"github.com/joaofilippe/inti/internal/api/dto"
	"github.com/joaofilippe/inti/internal/application/repository"
	"github.com/joaofilippe/inti/internal/infra/cache"
)

// ExtractService encapsula a lógica de extração de mandados via IA.
type ExtractService struct {
	apiKey string
	cache  *cache.Cache
	repo   *repository.MandadoRepository
}

func NewExtractService(apiKey string, c *cache.Cache, repo *repository.MandadoRepository) *ExtractService {
	return &ExtractService{apiKey: apiKey, cache: c, repo: repo}
}

func (s *ExtractService) ExtrairMandado(ctx context.Context, data []byte, lote string) (dto.MandadoExtraido, error) {
	key := hashKey(data)

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var result dto.MandadoExtraido
		if json.Unmarshal([]byte(cached), &result) == nil {
			normalizarExtraido(&result)
			result.Lote = lote
			return result, nil
		}
	} else if !cache.IsNil(err) {
		return dto.MandadoExtraido{}, fmt.Errorf("cache indisponível: %w", err)
	}

	dados, err := extrairDadosMandado(ctx, data, s.apiKey)
	if err != nil {
		return dto.MandadoExtraido{}, err
	}

	dados.Lote = lote

	if raw, err := json.Marshal(dados); err == nil {
		_ = s.cache.Set(ctx, key, string(raw))
	}

	if err := s.repo.SalvarExtraido(ctx, &dados); err != nil {
		log.Printf("erro ao salvar mandado extraído no banco: %v", err)
	}

	return dados, nil
}

func (s *ExtractService) ExtrairLote(ctx context.Context, data []byte, lote string) ([]dto.MandadoExtraido, error) {
	key := hashKey(data)

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var results []dto.MandadoExtraido
		if json.Unmarshal([]byte(cached), &results) == nil {
			normalizarLoteExtraido(results)
			for i := range results {
				results[i].Lote = lote
			}
			return results, nil
		}
	} else if !cache.IsNil(err) {
		return nil, fmt.Errorf("cache indisponível: %w", err)
	}

	dados, err := extrairDadosLote(ctx, data, s.apiKey)
	if err != nil {
		return nil, err
	}

	for i := range dados {
		dados[i].Lote = lote
	}

	if raw, err := json.Marshal(dados); err == nil {
		_ = s.cache.Set(ctx, key, string(raw))
	}

	if err := s.repo.SalvarLoteExtraido(ctx, dados); err != nil {
		log.Printf("erro ao salvar lote extraído no banco: %v", err)
	}

	return dados, nil
}

// --- helpers internos ---

var preposicoes = map[string]bool{
	"de": true, "da": true, "do": true,
	"das": true, "dos": true, "e": true, "a": true, "o": true,
}

func toTitleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		if i == 0 || !preposicoes[w] {
			runes := []rune(w)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

const geminiModel = "gemini-2.5-flash"

const promptSingle = `Esta é a folha de rosto de um mandado judicial brasileiro. Extraia os dados e retorne SOMENTE JSON válido, sem texto adicional, sem blocos de código markdown.

No canto superior da folha (margem ou cabeçalho superior), há anotações manuais organizadas em sequência da esquerda para a direita:
1. Data da Carga: data em que o mandado foi recebido/carregado (no formato DD/MM).
2. Sequência de códigos SEPARADOS POR HÍFEN: [Tipo do Ato] - [Encontrado] - [Resultado]
   MUITO IMPORTANTE: Os números anotados após a data da carga são separados por hífens (" - ") e representam TRÊS CAMPOS DISTINTOS:
   - Primeiro campo (antes do primeiro hífen): Código do Tipo do Ato (ex: "1", "2"). Quando houver MAIS DE UM tipo de ato, eles virão agrupados com a letra "e" (por exemplo: "2 e 4", "1 e 2"). NUNCA inclua os números após o primeiro hífen aqui!
   - Segundo campo (entre os hífens): Encontrado ("2" para Sim/Encontrado, "3" para Não/Não Encontrado).
   - Terceiro campo (após o segundo hífen): Resultado ("4" para Positivo, "5" para Parcial, "6" para Negativo).

   EXEMPLOS CRÍTICOS:
   - Se estiver anotado "2 - 2 - 4":
     TipoAto: "2" (Intimação) -> NUNCA coloque "2 - 2 - 4" no TipoAto! Apenas o primeiro "2" é o Tipo do Ato.
     Encontrado: "2" (Sim)
     Resultado: "4" (Positivo)
   - Se estiver anotado "2 e 4 - 2 - 4":
     TipoAto: "2 e 4" (Intimação e Penhora)
     Encontrado: "2" (Sim)
     Resultado: "4" (Positivo)
   - Se estiver anotado "1 - 3 - 6":
     TipoAto: "1" (Citação)
     Encontrado: "3" (Não)
     Resultado: "6" (Negativo)

3. Data e Hora do Cumprimento: data (formato DD/MM) e horário (formato HH:MM) em que o ato foi cumprido, localizada ao final da linha.

No lado esquerdo inferior da folha (margem ou rodapé inferior esquerdo), há anotações manuais de diligências:
- Diligências: data e hora da(s) diligência(s) realizada(s), no formato DD/MM HH:mm em texto (ex: "15/05 14:30"). Caso haja mais de uma diligência anotada, separe por vírgula em texto (ex: "15/05 14:30, 16/05 09:15"); se não houver anotação, deixe vazio.

Campos a extrair:
- Mandado: número do mandado judicial impresso (ex: "205.2026/000511-3")
- NumeroProcesso: número do processo judicial de origem (ex: "1234567-89.2024.8.26.0100"); deixar vazio se não encontrado
- Nome: nome completo do destinatário
- Documento: CPF, CNPJ ou RG (somente dígitos, sem pontuação, priorizar CPF caso haja mais de um)
- Sexo: "M" para masculino, "F" para feminino (inferir pelo nome se não explícito)
- Posicao: papel processual (ex: "Requerido", "Réu", "Executado")
- Endereco: endereço completo (rua, número, bairro)
- Cidade: nome da cidade
- DataCarga: data da carga localizada no canto superior à esquerda (no formato DD/MM); deixar vazio se não encontrado
- TipoAto: apenas o(s) número(s) do código do ato localizado ANTES do primeiro hífen (ex: "2", ou se múltiplos atos "2 e 4"); NUNCA coloque toda a sequência "2 - 2 - 4" aqui; extrair apenas o(s) número(s) do ato; deixar vazio se não encontrado
- Encontrado: "2" para sim, "3" para não, que é o código numérico localizado ENTRE os hífens da sequência; deixar vazio se não encontrado
- Resultado: "4" para Positivo, "5" para Parcial, "6" para Negativo, que é o código numérico localizado APÓS o segundo hífen da sequência; deixar vazio se não encontrado
- DataCumprimento: data do cumprimento anotada no canto superior ao final (no formato DD/MM); deixar vazio se não encontrado
- HoraCumprimento: horário do cumprimento anotado no canto superior ao final (no formato HH:MM); deixar vazio se não encontrado
- Diligencias: anotação manual de diligência(s) no lado esquerdo inferior da folha, com data e hora no formato DD/MM HH:mm em texto (ex: "15/05 14:30"); deixar vazio se não encontrado
- Whatsapp: número de WhatsApp preenchido manualmente caso haja anotação adicional/papelzinho; deixar vazio se não encontrado
- CPF: CPF preenchido manualmente caso haja anotação adicional/papelzinho; deixar vazio se não encontrado
- Email: e-mail preenchido manualmente caso haja anotação adicional/papelzinho; deixar vazio se não encontrado

Retorne exatamente este JSON:
{"Mandado":"","NumeroProcesso":"","Nome":"","Documento":"","Sexo":"","Posicao":"","Endereco":"","Cidade":"","DataCarga":"","TipoAto":"","Encontrado":"","Resultado":"","DataCumprimento":"","HoraCumprimento":"","Diligencias":"","Whatsapp":"","CPF":"","Email":""}`

const promptLote = `Este PDF contém múltiplas folhas de rosto de mandados judiciais brasileiros. Para CADA página que contiver uma folha de rosto, extraia os dados e retorne SOMENTE um array JSON válido, sem texto adicional, sem blocos de código markdown.

No canto superior de cada folha (margem ou cabeçalho superior), há anotações manuais organizadas em sequência da esquerda para a direita:
1. Data da Carga: data em que o mandado foi recebido/carregado (no formato DD/MM).
2. Sequência de códigos SEPARADOS POR HÍFEN: [Tipo do Ato] - [Encontrado] - [Resultado]
   MUITO IMPORTANTE: Os números anotados após a data da carga são separados por hífens (" - ") e representam TRÊS CAMPOS DISTINTOS:
   - Primeiro campo (antes do primeiro hífen): Código do Tipo do Ato (ex: "1", "2"). Quando houver MAIS DE UM tipo de ato, eles virão agrupados com a letra "e" (por exemplo: "2 e 4", "1 e 2"). NUNCA inclua os números após o primeiro hífen aqui!
   - Segundo campo (entre os hífens): Encontrado ("2" para Sim/Encontrado, "3" para Não/Não Encontrado).
   - Terceiro campo (após o segundo hífen): Resultado ("4" para Positivo, "5" para Parcial, "6" para Negativo).

   EXEMPLOS CRÍTICOS:
   - Se estiver anotado "2 - 2 - 4":
     TipoAto: "2" (Intimação) -> NUNCA coloque "2 - 2 - 4" no TipoAto! Apenas o primeiro "2" é o Tipo do Ato.
     Encontrado: "2" (Sim)
     Resultado: "4" (Positivo)
   - Se estiver anotado "2 e 4 - 2 - 4":
     TipoAto: "2 e 4" (Intimação e Penhora)
     Encontrado: "2" (Sim)
     Resultado: "4" (Positivo)
   - Se estiver anotado "1 - 3 - 6":
     TipoAto: "1" (Citação)
     Encontrado: "3" (Não)
     Resultado: "6" (Negativo)

3. Data e Hora do Cumprimento: data (formato DD/MM) e horário (formato HH:MM) em que o ato foi cumprido, localizada ao final da linha.

No lado esquerdo inferior da folha (margem ou rodapé inferior esquerdo), há anotações manuais de diligências:
- Diligências: anotações manuais contendo data e hora da(s) diligência(s), no formato DD/MM HH:mm em texto (ex: "15/05 14:30"). Caso haja mais de uma diligência anotada, separe por vírgula em texto (ex: "15/05 14:30, 16/05 09:15"); se não houver anotação, deixe vazio.

Campos por mandado:
- Mandado: número do mandado judicial impresso
- NumeroProcesso: número do processo judicial de origem (ex: "1234567-89.2024.8.26.0100"); deixar vazio se não encontrado
- Nome: nome completo do destinatário
- Documento: CPF, CNPJ ou RG (somente dígitos, sem pontuação, priorizar CPF caso haja mais de um)
- Sexo: "M" para masculino, "F" para feminino
- Posicao: papel processual (ex: "Requerido", "Réu", "Executado")
- Endereco: endereço completo, não incluir o CEP do endereço
- Cidade: nome da cidade
- DataCarga: data da carga localizada no canto superior à esquerda (no formato DD/MM); deixar vazio se não encontrado
- TipoAto: apenas o(s) número(s) do código do ato localizado ANTES do primeiro hífen (ex: "2", ou se múltiplos atos "2 e 4"); NUNCA coloque toda a sequência "2 - 2 - 4" aqui; extrair apenas o(s) número(s) do ato; deixar vazio se não encontrado
- Encontrado: "2" para sim, "3" para não, que é o código numérico localizado ENTRE os hífens da sequência; deixar vazio se não encontrado
- Resultado: "4" para Positivo, "5" para Parcial, "6" para Negativo, que é o código numérico localizado APÓS o segundo hífen da sequência; deixar vazio se não encontrado
- DataCumprimento: data do cumprimento anotada no canto superior ao final (no formato DD/MM); deixar vazio se não encontrado
- HoraCumprimento: horário do cumprimento anotado no canto superior ao final (no formato HH:MM); deixar vazio se não encontrado
- Diligencias: anotação manual de diligência(s) no lado esquerdo inferior da folha, com data e hora no formato DD/MM HH:mm em texto (ex: "15/05 14:30"); deixar vazio se não encontrado
- Whatsapp: número de WhatsApp preenchido manualmente caso haja anotação adicional/papelzinho; deixar vazio se não encontrado
- CPF: CPF preenchido manualmente caso haja anotação adicional/papelzinho; deixar vazio se não encontrado
- Email: e-mail preenchido manualmente caso haja anotação adicional/papelzinho; deixar vazio se não encontrado

Retorne exatamente este array JSON:
[{"Mandado":"","NumeroProcesso":"","Nome":"","Documento":"","Sexo":"","Posicao":"","Endereco":"","Cidade":"","DataCarga":"","TipoAto":"","Encontrado":"","Resultado":"","DataCumprimento":"","HoraCumprimento":"","Diligencias":"","Whatsapp":"","CPF":"","Email":""}]`

func detectMime(data []byte) string {
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return "application/pdf"
	}
	return http.DetectContentType(data)
}

func chamarGemini(ctx context.Context, apiKey string, data []byte, mimeType, prompt string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao criar cliente Gemini: %w", err)
	}

	contents := []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			{InlineData: &genai.Blob{MIMEType: mimeType, Data: data}},
			{Text: prompt},
		}, genai.RoleUser),
	}

	resp, err := client.Models.GenerateContent(ctx, geminiModel, contents, nil)
	if err != nil {
		return "", fmt.Errorf("erro na chamada à API Gemini: %w", err)
	}

	return resp.Text(), nil
}

func limparJSON(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "```"); idx != -1 {
		if end := strings.Index(s[idx:], "\n"); end != -1 {
			s = s[idx+end+1:]
		}
		if close := strings.LastIndex(s, "```"); close != -1 {
			s = s[:close]
		}
	}
	return strings.TrimSpace(s)
}

func inferirTipoDocumento(doc string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, doc)
	switch len(digits) {
	case 11:
		return "CPF"
	case 14:
		return "CNPJ"
	default:
		if len(digits) >= 7 && len(digits) <= 9 {
			return "RG"
		}
		return ""
	}
}

func extrairCodigosTipoAto(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var codigos []string
	var current strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				codigos = append(codigos, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		codigos = append(codigos, current.String())
	}
	if len(codigos) > 0 {
		return strings.Join(codigos, ", ")
	}
	return s
}

func separarCodigosAnotados(m *dto.MandadoExtraido) {
	m.TipoAto = strings.TrimSpace(m.TipoAto)
	m.Encontrado = strings.TrimSpace(m.Encontrado)
	m.Resultado = strings.TrimSpace(m.Resultado)

	if strings.Contains(m.TipoAto, "-") {
		parts := strings.Split(m.TipoAto, "-")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) >= 3 {
			last := parts[len(parts)-1]
			penult := parts[len(parts)-2]
			if (penult == "2" || penult == "3") && (last == "4" || last == "5" || last == "6") {
				m.TipoAto = strings.Join(parts[:len(parts)-2], "-")
				m.Encontrado = penult
				m.Resultado = last
			}
		}
	}
}

func normalizarExtraido(m *dto.MandadoExtraido) {
	separarCodigosAnotados(m)

	m.Nome = toTitleCase(m.Nome)
	m.Mandado = extrairNumeroMandado(m.Mandado)
	m.DataCarga = strings.TrimSpace(m.DataCarga)
	m.TipoAto = extrairCodigosTipoAto(m.TipoAto)
	m.DataCumprimento = strings.TrimSpace(m.DataCumprimento)
	m.HoraCumprimento = strings.TrimSpace(m.HoraCumprimento)
	m.Diligencias = strings.TrimSpace(m.Diligencias)

	enc := strings.ToLower(strings.TrimSpace(m.Encontrado))
	if strings.Contains(enc, "2") || strings.Contains(enc, "sim") {
		m.Encontrado = "2"
	} else if strings.Contains(enc, "3") || strings.Contains(enc, "nao") || strings.Contains(enc, "não") {
		m.Encontrado = "3"
	} else {
		m.Encontrado = enc
	}

	res := strings.ToLower(strings.TrimSpace(m.Resultado))
	if strings.Contains(res, "4") || strings.Contains(res, "pos") {
		m.Resultado = "4"
	} else if strings.Contains(res, "5") || strings.Contains(res, "parc") {
		m.Resultado = "5"
	} else if strings.Contains(res, "6") || strings.Contains(res, "neg") {
		m.Resultado = "6"
	} else {
		m.Resultado = res
	}

	if m.CPF != "" {
		digitsCPF := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, m.CPF)
		if len(digitsCPF) > 0 {
			m.Documento = digitsCPF
		}
	}
	m.TipoDocumento = inferirTipoDocumento(m.Documento)
}

func normalizarLoteExtraido(items []dto.MandadoExtraido) {
	for i := range items {
		normalizarExtraido(&items[i])
	}
}

func extrairDadosMandado(ctx context.Context, data []byte, apiKey string) (dto.MandadoExtraido, error) {
	log.Printf("Extraindo mandado via Gemini (%d bytes)", len(data))

	text, err := chamarGemini(ctx, apiKey, data, detectMime(data), promptSingle)
	if err != nil {
		return dto.MandadoExtraido{}, err
	}

	var result dto.MandadoExtraido
	if err := json.Unmarshal([]byte(limparJSON(text)), &result); err != nil {
		return dto.MandadoExtraido{}, fmt.Errorf("erro ao parsear JSON do Gemini: %w\nResposta: %s", err, text)
	}

	normalizarExtraido(&result)
	return result, nil
}

func extrairDadosLote(ctx context.Context, data []byte, apiKey string) ([]dto.MandadoExtraido, error) {
	log.Printf("Extraindo lote via Gemini (%d bytes)", len(data))

	text, err := chamarGemini(ctx, apiKey, data, "application/pdf", promptLote)
	if err != nil {
		return nil, err
	}

	var results []dto.MandadoExtraido
	if err := json.Unmarshal([]byte(limparJSON(text)), &results); err != nil {
		return nil, fmt.Errorf("erro ao parsear JSON do lote: %w\nResposta: %s", err, text)
	}

	normalizarLoteExtraido(results)
	return results, nil
}

func hashKey(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("extract:%x", h)
}
