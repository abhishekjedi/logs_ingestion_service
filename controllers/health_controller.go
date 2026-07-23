package controllers

import "error-logging/pkg/context"

type HealthController interface {
	GetAppHealth(c *context.ApiContext)
}
