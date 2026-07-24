package controllers

import "error-logging/pkg/context"

type AnalyticsController interface {
	GetServiceOverview(c *context.ApiContext)
	GetReleaseHealth(c *context.ApiContext)
}
