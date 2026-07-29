package controllers

import "error-logging/pkg/context"

type ReplayController interface {
	ListIntegrations(c *context.ApiContext)
	UpsertIntegration(c *context.ApiContext)
	DeleteIntegration(c *context.ApiContext)
	GetSessionContext(c *context.ApiContext)
}
