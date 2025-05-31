package characters

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Manager interface {
	Create(ctx context.Context, character Character) (Character, error)
	GetById(ctx context.Context, characterID string) (Character, error)
	GetByName(ctx context.Context, characterName string) (Character, error)
}

type Controller struct {
	manager Manager
}

func NewController(manager Manager) *Controller {
	return &Controller{manager: manager}
}

func (c *Controller) Create(ctx *gin.Context) {
	var characterRequest CharacterRequest
	if err := ctx.BindJSON(&characterRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	createdCharacter, err := c.manager.Create(ctx, characterRequest.ToCharacter())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, NewCharacterResponse(createdCharacter))
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

	ctx.JSON(http.StatusOK, NewCharacterResponse(character))
}

func (c *Controller) GetByName(ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query param 'name' is required"})
		return
	}

	character, err := c.manager.GetByName(ctx, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, NewCharacterResponse(character))
}

type CharacterRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *CharacterRequest) ToCharacter() Character {
	return Character{
		Name: r.Name,
	}
}

type CharacterResponse struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	BooksTitles []string `json:"booksTitles,omitempty"`
}

func NewCharacterResponse(character Character) CharacterResponse {
	var booksTitles []string
	for _, b := range character.Books {
		booksTitles = append(booksTitles, b.Title)
	}

	return CharacterResponse{
		ID:          character.ID,
		Name:        character.Name,
		BooksTitles: booksTitles,
	}
}
