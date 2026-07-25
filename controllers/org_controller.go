package controllers

import "error-logging/pkg/context"

type OrgController interface {
	CreateOrg(c *context.ApiContext)
	ListMyOrgs(c *context.ApiContext)
	InviteMember(c *context.ApiContext)
	ListMembers(c *context.ApiContext)
	RemoveMember(c *context.ApiContext)
}
