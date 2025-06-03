package characters

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Manager interface {
	Create(ctx context.Context, character Character, bookTitles []string) (Character, error)
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
		ctx.Error(err)
		return
	}

	createdCharacter, err := c.manager.Create(ctx, characterRequest.ToCharacter(), characterRequest.BookTitles)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, NewCharacterResponse(createdCharacter))
}

func (c *Controller) GetBy(ctx *gin.Context) {
	var getByRequest GetByRequest
	if err := ctx.BindUri(&getByRequest); err != nil {
		ctx.Error(err)
		return
	}

	characterID, err := uuid.Parse(getByRequest.Character)
	if err != nil {
		c.getByName(ctx, getByRequest.Character)
		return
	}

	c.getById(ctx, characterID.String())
	return
}

func (c *Controller) getById(ctx *gin.Context, characterID string) {
	character, err := c.manager.GetById(ctx, characterID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, NewCharacterResponse(character))
}

func (c *Controller) getByName(ctx *gin.Context, characterName string) {
	character, err := c.manager.GetByName(ctx, characterName)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, NewCharacterResponse(character))
}

type CharacterRequest struct {
	Name       string   `json:"name" binding:"required"`
	BookTitles []string `json:"bookTitles"`
}

func (r *CharacterRequest) ToCharacter() Character {
	return Character{
		Name: r.Name,
	}
}

type GetByRequest struct {
	Character string `uri:"character" binding:"required"`
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
