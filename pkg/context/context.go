package context

import "github.com/gin-gonic/gin"

type ApiContext struct {
	*gin.Context
}

func Wrap(handler func(*ApiContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(&ApiContext{Context: c})
	}
}
