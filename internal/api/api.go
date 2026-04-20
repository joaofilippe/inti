package api

import (
	"github.com/joaofilippe/inti/internal/infra/server"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type API struct {
	server    *server.Server
	mandadoH  MandadoHandler
	extractH  ExtractHandler
	tipoAtoH  TipoAtoHandler
}

func New(srv *server.Server, mandadoH MandadoHandler, extractH ExtractHandler, tipoAtoH TipoAtoHandler) *API {
	a := &API{
		server:   srv,
		mandadoH: mandadoH,
		extractH: extractH,
		tipoAtoH: tipoAtoH,
	}
	a.registerRoutes()
	return a
}

func (a *API) registerRoutes() {
	e := a.server.Echo

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Static("/", "public")
	e.GET("/", func(c echo.Context) error {
		return c.File("public/form.html")
	})

	e.GET("/api/tipos-ato", a.tipoAtoH.ListarTiposAto)

	e.POST("/api/extract", a.extractH.ExtrairMandado)
	e.POST("/api/mandados", a.mandadoH.GerarMandado)

	e.POST("/api/batch/extract", a.extractH.ExtrairLote)
	e.POST("/api/batch/mandados", a.mandadoH.GerarLote)
}

func (a *API) Start() error {
	return a.server.Start()
}
