package svc

import (
	"github.com/iluyuns/alpha-trade/internal/config"
	"github.com/iluyuns/alpha-trade/internal/middleware"
	"github.com/iluyuns/alpha-trade/internal/pkg/email"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config  config.Config
	Auth    rest.Middleware
	Passkey rest.Middleware
	Email   email.EmailService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		Auth:    middleware.NewAuthMiddleware().Handle,
		Passkey: middleware.NewPasskeyMiddleware().Handle,
		Email:   email.NewAWSSES(&c.AWS),
	}
}
