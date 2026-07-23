package controllers

import "error-logging/pkg/context"

type ServiceController interface {
	CreateService(c *context.ApiContext)
	ListServices(c *context.ApiContext)
}
