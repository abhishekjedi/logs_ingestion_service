package dto

import "github.com/gin-gonic/gin"

type Route interface {
	Register(*gin.RouterGroup)
}
