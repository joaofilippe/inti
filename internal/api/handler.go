package api

import "github.com/labstack/echo/v4"

type MandadoHandler interface {
	GerarMandado(c echo.Context) error
	GerarLote(c echo.Context) error
}

type ExtractHandler interface {
	ExtrairMandado(c echo.Context) error
	ExtrairLote(c echo.Context) error
	ListarResumo(c echo.Context) error
}

type TipoAtoHandler interface {
	ListarTiposAto(c echo.Context) error
}

type MotivoNaoRealizacaoHandler interface {
	ListarMotivosNaoRealizacao(c echo.Context) error
}
