package entities

type TipoContato string

const (
	ContatoTelefone TipoContato = "TELEFONE"
	ContatoWhatsapp TipoContato = "WHATSAPP"
	ContatoEmail    TipoContato = "EMAIL"
)

type Contato struct {
	Tipo            TipoContato
	Valor           string
	IsTerceiro      bool
	NomeTerceiro    string
	RelacaoTerceiro string
}
