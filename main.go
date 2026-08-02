// @title           Inti API
// @version         1.0
// @description     API para geração e extração de mandados judiciais.
// @host            localhost:8080
// @BasePath        /

package main

import (
	"context"
	"encoding/json"
	"log"

	_ "github.com/joaofilippe/inti/docs"

	"github.com/joaofilippe/inti/config"
	"github.com/joaofilippe/inti/internal/api"
	handler "github.com/joaofilippe/inti/internal/api/handlers"
	"github.com/joaofilippe/inti/internal/application/repository"
	"github.com/joaofilippe/inti/internal/application/service"
	"github.com/joaofilippe/inti/internal/domain/entities"
	"github.com/joaofilippe/inti/internal/infra/cache"
	"github.com/joaofilippe/inti/internal/infra/database"
	"github.com/joaofilippe/inti/internal/infra/server"
)

func carregarTiposAtoNoCache(ctx context.Context, c *cache.Cache, tiposAto map[int]entities.TipoAto) error {
	type tipoAtoJSON struct {
		Codigo    int    `json:"codigo"`
		Descricao string `json:"descricao"`
		Positivo  string `json:"positivo,omitempty"`
		Negativo  string `json:"negativo,omitempty"`
	}

	lista := make([]tipoAtoJSON, 0, len(tiposAto))
	for _, t := range tiposAto {
		lista = append(lista, tipoAtoJSON{
			Codigo:    t.Codigo,
			Descricao: t.Descricao,
			Positivo:  t.Positivo,
			Negativo:  t.Negativo,
		})
	}

	data, err := json.Marshal(lista)
	if err != nil {
		return err
	}
	return c.SetTiposAto(ctx, string(data))
}

func main() {
	cfg := config.Load()

	redisCache, err := cache.New(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao Redis: %v", err)
	}

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Erro ao aplicar migrations: %v", err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	mandadoRepo := repository.NewMandadoRepository(db)
	tipoAtoRepo := repository.NewTipoAtoRepository(db)
	motivoRepo := repository.NewMotivoNaoRealizacaoRepository(db)
	userRepo := database.NewUserRepository(db)

	tiposAto, err := tipoAtoRepo.CarregarTodos(context.Background())
	if err != nil {
		log.Fatalf("Erro ao carregar tipos de ato: %v", err)
	}

	motivosList, err := motivoRepo.CarregarTodos(context.Background())
	if err != nil {
		log.Fatalf("Erro ao carregar motivos de nao realizacao: %v", err)
	}
	motivosMap := make(map[int]entities.MotivoNaoRealizacao)
	for _, m := range motivosList {
		motivosMap[m.ID] = m
	}

	loteRepo := database.NewLoteRepository(db)

	if err := carregarTiposAtoNoCache(context.Background(), redisCache, tiposAto); err != nil {
		log.Printf("Aviso: não foi possível pré-carregar tipos de ato no Redis: %v", err)
	}

	extractSvc := service.NewExtractService(cfg.GeminiAPIKey, redisCache, mandadoRepo, loteRepo)
	mandadoSvc := service.NewMandadoService(tiposAto, motivosMap)
	authSvc := service.NewAuthService(userRepo)

	mandadoH := handler.NewMandadoHandler(cfg, mandadoSvc)
	extractH := handler.NewExtractHandler(extractSvc)
	tipoAtoH := handler.NewTipoAtoHandler(redisCache, tipoAtoRepo)
	motivoH := handler.NewMotivoNaoRealizacaoHandler(redisCache, motivoRepo)
	authH := handler.NewAuthHandler(authSvc)

	srv := server.New(cfg.ServerAddr)
	a := api.New(srv, authH, mandadoH, extractH, tipoAtoH, motivoH)

	log.Fatal(a.Start())
}
