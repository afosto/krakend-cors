package gin

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	krakendcors "github.com/krakendio/krakend-cors/v2"
	"github.com/krakendio/krakend-cors/v2/mux"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/rs/cors"
	wrapper "github.com/rs/cors/wrapper/gin"
)

// New returns a gin.HandlerFunc with the CORS configuration provided in the ExtraConfig
func New(e config.ExtraConfig) gin.HandlerFunc {
	tmp := krakendcors.ConfigGetter(e)
	if tmp == nil {
		return nil
	}
	cfg, ok := tmp.(krakendcors.Config)
	if !ok {
		return nil
	}

	var allowOriginFunc func(origin string) bool

	if len(cfg.AllowOrigins) == 0 {
		allowOriginFunc = func(origin string) bool {
			return true
		}
	}

	return wrapper.New(cors.Options{
		AllowedOrigins:   cfg.AllowOrigins,
		AllowedMethods:   cfg.AllowMethods,
		AllowOriginFunc:  allowOriginFunc,
		AllowedHeaders:   cfg.AllowHeaders,
		ExposedHeaders:   cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           int(cfg.MaxAge.Seconds()),
	})
}

// RunServer defines the interface of a function used by the KrakenD router to start the service
type RunServer func(context.Context, config.ServiceConfig, http.Handler) error

// NewRunServer returns a RunServer wrapping the injected one with a CORS middleware, so it is called before the
// actual router checks the URL, method and other details related to selecting the proper handler for the
// incoming request
func NewRunServer(next RunServer) RunServer {
	return NewRunServerWithLogger(next, nil)
}

// NewRunServerWithLogger returns a RunServer wrapping the injected one with a CORS middleware, so it is called before the
// actual router checks the URL, method and other details related to selecting the proper handler for the
// incoming request
func NewRunServerWithLogger(next RunServer, l logging.Logger) RunServer {
	return func(ctx context.Context, cfg config.ServiceConfig, handler http.Handler) error {

		l, _ := logging.NewLogger("DEBUG", os.Stdout, "")

		corsMw := mux.NewWithLogger(cfg.ExtraConfig, l)
		if corsMw == nil {
			return next(ctx, cfg, handler)
		}
		return next(ctx, cfg, corsMw.Handler(handler))
	}
}
