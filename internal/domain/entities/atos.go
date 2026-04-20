package entities

type TipoAto struct {
	Codigo    int
	Descricao string
	Positivo  string // vazio se não se aplica
	Negativo  string // vazio se não se aplica
}
