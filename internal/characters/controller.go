package characters

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Manager interface {
	Create(ctx context.Context, character Character) (Character, error)
	GetById(ctx context.Context, characterID string) (Character, error)
}

type Controller struct {
	manager Manager
}

func NewController(manager Manager) *Controller {
	return &Controller{manager: manager}
}

func (c *Controller) Create(ctx *gin.Context) {
	var character Character
	if err := ctx.BindJSON(&character); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	createdCharacter, err := c.manager.Create(ctx, character)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdCharacter)
}

func (c *Controller) GetById(ctx *gin.Context) {
	characterID := ctx.Param("characterID")
	if characterID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "characterID is required"})
		return
	}

	character, err := c.manager.GetById(ctx, characterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, character)
}
