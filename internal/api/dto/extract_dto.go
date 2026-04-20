package dto

type MandadoExtraido struct {
	Mandado        string `json:"Mandado"`
	Lote           string `json:"Lote"`
	NumeroProcesso string `json:"NumeroProcesso"`
	Nome           string `json:"Nome"`
	Documento      string `json:"Documento"`
	Sexo           string `json:"Sexo"`
	Posicao        string `json:"Posicao"`
	Endereco       string `json:"Endereco"`
	Cidade         string `json:"Cidade"`
	TipoDocumento  string `json:"TipoDocumento"`
	Whatsapp       string `json:"Whatsapp"`
	CPF            string `json:"CPF"`
	Email          string `json:"Email"`
}
