package ctx

import (
	"context"
	"gitlab.com/distributed_lab/logan/v3"
	"net/http"

	"github.com/jarm-proxy/internal/config"
)

type key int

const (
	keyLog key = iota
	keyCfg
)

func SetLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, keyLog, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(keyLog).(*logan.Entry)
}

func SetCfg(config config.Config) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, keyCfg, config)
	}
}

func Cfg(r *http.Request) config.Config {
	return r.Context().Value(keyCfg).(config.Config)
}