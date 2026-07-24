package controllers

import "error-logging/pkg/context"

type IssueController interface {
	ListIssues(c *context.ApiContext)
	GetIssue(c *context.ApiContext)
	GetTimeseries(c *context.ApiContext)
	GetEvents(c *context.ApiContext)
}
