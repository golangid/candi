package base

import (
	"crypto/rsa"

	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/publisher"
)

// Dependency base
type Dependency interface {
	GetMiddleware() middleware.Middleware
	GetPublisher() publisher.Publisher
}

// Option func type
type Option func(*deps)
type deps struct {
	mw         middleware.Middleware
	pub        publisher.Publisher
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// SetMiddleware option func
func SetMiddleware(mw middleware.Middleware) Option {
	return func(d *deps) {
		d.mw = mw
	}
}

// SetPublisher option func
func SetPublisher(pub publisher.Publisher) Option {
	return func(d *deps) {
		d.pub = pub
	}
}

// InitDependency constructor
func InitDependency(opts ...Option) Dependency {
	opt := new(deps)

	for _, o := range opts {
		o(opt)
	}

	return opt
}

func (d *deps) GetMiddleware() middleware.Middleware {
	return d.mw
}

func (d *deps) GetPublisher() publisher.Publisher {
	return d.pub
}
