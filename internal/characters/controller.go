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
	var characterDTO CharacterDTO
	if err := ctx.BindJSON(&characterDTO); err != nil {
		ctx.Error(err)
		return
	}

	createdCharacter, err := c.manager.Create(ctx, characterDTO.ToCharacter(), characterDTO.BookTitles)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, NewCharacterDTO(createdCharacter))
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

	ctx.JSON(http.StatusOK, NewCharacterDTO(character))
}

func (c *Controller) getByName(ctx *gin.Context, characterName string) {
	character, err := c.manager.GetByName(ctx, characterName)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, NewCharacterDTO(character))
}

type GetByRequest struct {
	Character string `uri:"character" binding:"required"`
}

type CharacterDTO struct {
	ID         string     `json:"id,omitempty"`
	Name       string     `json:"name" binding:"required"`
	Actors     []ActorDTO `json:"actors,omitempty"`
	BookTitles []string   `json:"bookTitles,omitempty"`
}

type ActorDTO struct {
	Name string `json:"name" binding:"required"`
	IMDB string `json:"imdb" binding:"required"`
}

func NewCharacterDTO(character Character) CharacterDTO {
	var booksTitles []string
	for _, b := range character.Books {
		booksTitles = append(booksTitles, b.Title)
	}

	return CharacterDTO{
		ID:         character.ID,
		Name:       character.Name,
		BookTitles: booksTitles,
	}
}

func (r *CharacterDTO) ToCharacter() Character {
	var actorList []Actor

	for _, actor := range r.Actors {
		actorList = append(actorList, Actor{
			Name: actor.Name,
			IMDB: actor.IMDB,
		})
	}

	character := Character{Name: r.Name, Actors: actorList}

	return character
}
