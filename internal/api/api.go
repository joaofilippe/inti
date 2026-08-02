package api

import (
	"os"

	echojwt "github.com/labstack/echo-jwt/v4"
	handler "github.com/joaofilippe/inti/internal/api/handlers"
	"github.com/joaofilippe/inti/internal/infra/server"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type API struct {
	server       *server.Server
	authH        handler.AuthHandler
	mandadoH     MandadoHandler
	extractH     ExtractHandler
	tipoAtoH     TipoAtoHandler
	motivoAtoH   MotivoNaoRealizacaoHandler
}

func New(srv *server.Server, authH handler.AuthHandler, mandadoH MandadoHandler, extractH ExtractHandler, tipoAtoH TipoAtoHandler, motivoAtoH MotivoNaoRealizacaoHandler) *API {
	a := &API{
		server:     srv,
		authH:      authH,
		mandadoH:   mandadoH,
		extractH:   extractH,
		tipoAtoH:   tipoAtoH,
		motivoAtoH: motivoAtoH,
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

	// Public routes
	e.POST("/api/auth/login", a.authH.Login)
	e.POST("/api/auth/register", a.authH.Register)

	e.GET("/api/tipos-ato", a.tipoAtoH.ListarTiposAto)
	e.GET("/api/motivos-nao-realizacao", a.motivoAtoH.ListarMotivosNaoRealizacao)

	// Protected routes
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_please_change"
	}
	
	apiGroup := e.Group("/api")
	apiGroup.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(secret),
	}))

	apiGroup.POST("/extract", a.extractH.ExtrairMandado)
	apiGroup.POST("/mandados", a.mandadoH.GerarMandado)
	apiGroup.GET("/mandados/resumo", a.extractH.ListarResumo)

	apiGroup.POST("/batch/extract", a.extractH.ExtrairLote)
	apiGroup.POST("/batch/pre-cadastro/excel", a.extractH.ExtrairDeExcel)
	apiGroup.POST("/batch/mandados", a.mandadoH.GerarLote)
}

func (a *API) Start() error {
	return a.server.Start()
}
