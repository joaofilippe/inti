package dto

type AtoDTO struct {
	CodigoAto           int    `json:"CodigoAto"`
	DataCumprimento     string `json:"DataCumprimento"`
	Horario             string `json:"Horario"`
	Realizado           *bool  `json:"Realizado,omitempty"`
	MotivoNaoRealizacao string `json:"MotivoNaoRealizacao,omitempty"`
}

type MandadoPositivoDTO struct {
	Mandado           string       `json:"Mandado"`
	Lote              string       `json:"Lote"`
	Situacao          string       `json:"Situacao"`
	DataCarga         string       `json:"DataCarga"`
	Atos              []AtoDTO     `json:"Atos"`
	Diligencias       string       `json:"Diligencias"`
	IsPJ              bool         `json:"IsPJ"`
	Contato           []ContatoDTO `json:"Contato"`
	Cidade            string       `json:"Cidade"`
	Endereco          string       `json:"Endereco"`
	RepresentanteNome string       `json:"RepresentanteNome"`
	RepresentanteDoc  string       `json:"RepresentanteDoc"`
	Nome              string       `json:"Nome"`
	Sexo              string       `json:"Sexo"`
	Posicao           string       `json:"Posicao"`
	Documento         string       `json:"Documento"`
	TipoDocumento     string       `json:"TipoDocumento"`
	Obs               string       `json:"Obs"`
}
