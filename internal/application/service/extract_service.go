package service

import (
	"bytes"
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
	"github.com/joaofilippe/inti/internal/document"
	"github.com/joaofilippe/inti/internal/infra/cache"
	"github.com/joaofilippe/inti/internal/infra/database"
)

// ExtractService encapsula a lógica de extração de mandados via IA.
type ExtractService struct {
	apiKey   string
	cache    *cache.Cache
	repo     *repository.MandadoRepository
	loteRepo *database.LoteRepository
}

func NewExtractService(apiKey string, c *cache.Cache, repo *repository.MandadoRepository, loteRepo *database.LoteRepository) *ExtractService {
	return &ExtractService{apiKey: apiKey, cache: c, repo: repo, loteRepo: loteRepo}
}

func (s *ExtractService) ExtrairMandado(ctx context.Context, data []byte, lote string, adminUserID string) (dto.MandadoExtraido, error) {
	_, err := s.loteRepo.FindOrCreateLote(ctx, lote, adminUserID)
	if err != nil {
		return dto.MandadoExtraido{}, fmt.Errorf("erro ao registrar lote: %w", err)
	}
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

func (s *ExtractService) ExtrairLote(ctx context.Context, data []byte, lote string, adminUserID string) ([]dto.MandadoExtraido, error) {
	_, err := s.loteRepo.FindOrCreateLote(ctx, lote, adminUserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar lote: %w", err)
	}
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

func (s *ExtractService) ListarResumo(ctx context.Context, lote string, adminUserID string) ([]dto.MandadoResumoDTO, error) {
	return s.repo.ListarResumo(ctx, lote, adminUserID)
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

Campos:
- Mandado: número do mandado (ex: "205.2026/000511-3")
- NumeroProcesso: número do processo judicial de origem (ex: "1234567-89.2024.8.26.0100"); deixar vazio se não encontrado
- Nome: nome completo do destinatário
- Documento: CPF, CNPJ ou RG (somente dígitos, sem pontuação, priorizar CPF caso haja mais de um)
- Sexo: "M" para masculino, "F" para feminino (inferir pelo nome se não explícito)
- Posicao: papel processual (ex: "Requerido", "Réu", "Executado")
- Endereco: endereço completo (rua, número, bairro)
- Cidade: nome da cidade (sem a sigla do estado)
- Whatsapp: número de WhatsApp preenchido manualmente no papelzinho colado no documento; deixar vazio se não encontrado
- CPF: CPF preenchido manualmente no papelzinho colado no documento (extrair exatamente como escrito, com ou sem pontuação); deixar vazio se não encontrado
- Email: e-mail preenchido manualmente no papelzinho colado no documento; deixar vazio se não encontrado

Retorne exatamente este JSON:
{"Mandado":"","NumeroProcesso":"","Nome":"","Documento":"","Sexo":"","Posicao":"","Endereco":"","Cidade":"","Whatsapp":"","CPF":"","Email":""}`

const promptLote = `Este PDF contém múltiplas folhas de rosto de mandados judiciais brasileiros. Para CADA página que contiver uma folha de rosto, extraia os dados e retorne SOMENTE um array JSON válido, sem texto adicional, sem blocos de código markdown.

Campos por mandado:
- Mandado: número do mandado
- NumeroProcesso: número do processo judicial de origem (ex: "1234567-89.2024.8.26.0100"); deixar vazio se não encontrado
- Nome: nome completo do destinatário
- Documento: CPF, CNPJ ou RG (somente dígitos, sem pontuação, priorizar CPF caso haja mais de um)
- Sexo: "M" para masculino, "F" para feminino
- Posicao: papel processual (ex: "Requerido", "Réu", "Executado")
- Endereco: endereço completo, não incluir o CEP do endereço
- Cidade: nome da cidade (sem a sigla do estado)
- Whatsapp: número de WhatsApp preenchido manualmente no papelzinho colado no documento; deixar vazio se não encontrado
- CPF: CPF preenchido manualmente no papelzinho colado no documento (extrair exatamente como escrito, com ou sem pontuação); deixar vazio se não encontrado
- Email: e-mail preenchido manualmente no papelzinho colado no documento; deixar vazio se não encontrado

Retorne exatamente este array JSON:
[{"Mandado":"","NumeroProcesso":"","Nome":"","Documento":"","Sexo":"","Posicao":"","Endereco":"","Cidade":"","Whatsapp":"","CPF":"","Email":""}]`

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

func normalizarExtraido(m *dto.MandadoExtraido) {
	m.Nome = toTitleCase(m.Nome)
	m.Mandado = extrairNumeroMandado(m.Mandado)

	if idx := strings.LastIndex(m.Cidade, "-"); idx != -1 {
		statePart := strings.TrimSpace(m.Cidade[idx+1:])
		if len(statePart) == 2 {
			m.Cidade = strings.TrimSpace(m.Cidade[:idx])
		}
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

// ExtrairDeExcel processa uma planilha Excel e salva os dados no repositório.
func (s *ExtractService) ExtrairDeExcel(ctx context.Context, file []byte, filename string, adminUserID string) ([]dto.MandadoExtraido, error) {
	_, err := s.loteRepo.FindOrCreateLote(ctx, filename, adminUserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar lote: %w", err)
	}
	reader := bytes.NewReader(file)
	mandados, err := document.ParseExcel(reader, filename)
	if err != nil {
		return nil, fmt.Errorf("falha ao interpretar excel: %w", err)
	}

	if len(mandados) > 0 {
		err = s.repo.SalvarLoteExtraido(ctx, mandados)
		if err != nil {
			return nil, fmt.Errorf("falha ao salvar mandados: %w", err)
		}
	}

	return mandados, nil
}

func hashKey(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("extract:%x", h)
}
