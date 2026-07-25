package controllers

import "error-logging/pkg/context"

type AuthController interface {
	Me(c *context.ApiContext)
	Logout(c *context.ApiContext)
	GoogleLogin(c *context.ApiContext)
	GoogleCallback(c *context.ApiContext)
}
