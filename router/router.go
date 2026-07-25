package router

import (
	"error-logging/dto"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Routes []dto.Route
	CORS   gin.HandlerFunc
}

func (r *Router) Setup() *gin.Engine {
	engine := gin.Default()
	if r.CORS != nil {
		engine.Use(r.CORS)
	}

	api := engine.Group("/api")
	for _, route := range r.Routes {
		route.Register(api)
	}

	return engine
}
