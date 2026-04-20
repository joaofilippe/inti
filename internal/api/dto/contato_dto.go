package dto

type ContatoDTO struct {
	Tipo            string `json:"Tipo"`
	Valor           string `json:"Valor"`
	IsTerceiro      bool   `json:"IsTerceiro"`
	NomeTerceiro    string `json:"NomeTerceiro"`
	RelacaoTerceiro string `json:"RelacaoTerceiro"`
}
