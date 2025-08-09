package health

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Manager interface {
	Health(ctx context.Context) map[string]bool
}

type Controller struct {
	manager Manager
}

func NewController(manager Manager) *Controller {
	return &Controller{manager: manager}
}

func (c *Controller) Health(ctx *gin.Context) {
	ctx.JSON(200, c.manager.Health(ctx))
}
