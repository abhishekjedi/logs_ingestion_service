package controllers

import "error-logging/pkg/context"

type ProjectController interface {
	CreateProject(c *context.ApiContext)
	GetProject(c *context.ApiContext)
	ListProjects(c *context.ApiContext)
}
