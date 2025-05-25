package health

import (
	"github.com/gin-gonic/gin"
)

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Health(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"api": true})
}
