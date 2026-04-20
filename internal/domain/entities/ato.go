package entities

type Ato struct {
	CodigoAto           int    `json:"codigo_ato"`
	DataCumprimento     string `json:"data_cumprimento"`
	Horario             string `json:"horario"`
	Realizado           *bool  `json:"realizado,omitempty"`
	MotivoNaoRealizacao string `json:"motivo_nao_realizacao,omitempty"`
}
