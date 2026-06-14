package entities

type MotivoNaoRealizacao struct {
	ID         int    `json:"id"`
	Codigo     string `json:"codigo"`
	Explicacao string `json:"explicacao"`
}
