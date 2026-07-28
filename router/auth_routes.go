package router

import (
	"error-logging/controllers"
	"error-logging/middleware"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

type AuthRouter struct {
	AuthController controllers.AuthController
	Auth           *middleware.AuthMiddleware
}

func NewAuthRouter(ac controllers.AuthController, auth *middleware.AuthMiddleware) *AuthRouter {
	return &AuthRouter{AuthController: ac, Auth: auth}
}

func (r *AuthRouter) Register(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	auth.POST("/logout", apicontext.Wrap(r.AuthController.Logout))
	auth.GET("/me", r.Auth.RequireUser(), apicontext.Wrap(r.AuthController.Me))
	auth.GET("/google/login", apicontext.Wrap(r.AuthController.GoogleLogin))
	auth.GET("/google/callback", apicontext.Wrap(r.AuthController.GoogleCallback))
}
