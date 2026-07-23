package router

import (
	"error-logging/dto"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Routes []dto.Route
}

func (r *Router) Setup() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")

	for _, route := range r.Routes {
		route.Register(api)
	}

	return router
}
