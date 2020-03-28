package services

import (
	"fmt"

	"github.com/agungdwiprasetyo/backend-microservices/internal/factory"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/warung"
)

// InitService from env
func InitService(serviceName constant.Service, params *base.ModuleParam) factory.ServiceFactory {
	switch serviceName {
	case warung.Warung:
		return warung.NewService(params)
	default:
		panic(fmt.Errorf(`service '%s' not registered`, serviceName))
	}
}
