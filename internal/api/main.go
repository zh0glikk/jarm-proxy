package api

import (
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"

	"github.com/jarm-proxy/internal/api/ctx"
	"github.com/jarm-proxy/internal/api/handlers"
	"github.com/jarm-proxy/internal/config"
)

func Router(cfg config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(cfg.Log()),
		ape.LoganMiddleware(cfg.Log()),
		ape.CtxMiddleware(
			ctx.SetLog(cfg.Log()),
			ctx.SetCfg(cfg),
		),
	)

	//mb init db state via terraform or smth else
	//check if it exists in db
	//take fingerprint
	//compare with one from db
	//return 200 or 403

	r.Get("/", handlers.ProxyHandler)

	return r
}
