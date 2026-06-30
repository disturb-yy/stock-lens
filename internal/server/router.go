package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReadinessCheck func() error

type RouteRegistrar func(*gin.Engine)

type RouterOption func(*routerConfig)

type routerConfig struct {
	registrars []RouteRegistrar
}

func WithRoutes(registrar RouteRegistrar) RouterOption {
	return func(cfg *routerConfig) {
		if registrar != nil {
			cfg.registrars = append(cfg.registrars, registrar)
		}
	}
}

func NewRouter(ready ReadinessCheck, options ...RouterOption) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	cfg := routerConfig{}
	for _, option := range options {
		option(&cfg)
	}

	router := gin.New()
	router.Use(RequestIDMiddleware(), AccessLogMiddleware(nil))
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/readyz", func(c *gin.Context) {
		if ready != nil {
			if err := ready(); err != nil {
				c.Status(http.StatusServiceUnavailable)
				return
			}
		}
		c.Status(http.StatusOK)
	})
	for _, registrar := range cfg.registrars {
		registrar(router)
	}
	return router
}
