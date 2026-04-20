package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	Echo *echo.Echo
	Addr string
}

func New(addr string) *Server {
	e := echo.New()

	e.Use(requestLogger())
	e.Use(middleware.Recover())

	return &Server{Echo: e, Addr: addr}
}

func (s *Server) Start() error {
	return s.Echo.Start(s.Addr)
}
