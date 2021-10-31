package service

import (
	"context"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"

	"github.com/jarm-proxy/internal/api"
	"github.com/jarm-proxy/internal/config"
)

type service struct {
	logger           *logan.Entry
	cfg 			 config.Config
}

func NewService(cfg config.Config) *service {
	return &service{
		logger:    cfg.Log(),
		cfg:	   cfg,
	}
}

func (s *service) Run(ctx context.Context) error{
	r := api.Router(s.cfg)

	err := http.Serve(s.cfg.Listener(), r)
	if err != nil {
		return errors.Wrap(err, "server stopped with error")
	}

	return nil
}

