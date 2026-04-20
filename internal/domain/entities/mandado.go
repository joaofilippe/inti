package entities

type Mandado struct {
	Mandado           string
	Lote              string
	Situacao          string
	DataCarga         string
	Atos              []Ato
	Diligencias       string
	IsPJ              bool
	Contatos          []Contato
	Cidade            string
	Endereco          string
	Nome              string
	Sexo              string
	Posicao           string
	Documento         string
	TipoDocumento     string
	RepresentanteNome string
	RepresentanteDoc  string
	Obs               string
}
