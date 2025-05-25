package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/ggoulart/michael-connelly-api/internal/characters"
	"github.com/ggoulart/michael-connelly-api/internal/health"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func main() {
	region := "us-east-1"
	booksTable := "books"
	characterTable := "characters"

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	v := validator.New()

	dynamodbClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		//o.BaseEndpoint = aws.String("https://dynamodb.eu-west-1.amazonaws.com")
		o.BaseEndpoint = aws.String("http://localhost:8000")
	})
	uuidGenerator := uuid.New

	healthController := health.NewController()

	booksRepository := books.NewRepository(dynamodbClient, booksTable, uuidGenerator)
	booksService := books.NewService(booksRepository)
	booksController := books.NewController(booksService, v)

	charactersRepository := characters.NewRepository(dynamodbClient, characterTable, uuidGenerator)
	charactersService := characters.NewService(charactersRepository)
	charactersController := characters.NewController(charactersService, v)

	r := router(booksController, charactersController, healthController)

	adapter := ginadapter.New(r)

	lambda.Start(adapter.ProxyWithContext)
}

func router(booksController *books.Controller, charactersController *characters.Controller, healthController *health.Controller) *gin.Engine {
	r := gin.Default()

	r.GET("/health", healthController.Health)

	book := r.Group("/books")
	book.POST("/", booksController.Create)
	book.GET("/:bookID", booksController.GetById)

	character := r.Group("/characters")
	character.POST("/", charactersController.Create)
	character.GET("/:characterID", charactersController.GetById)

	return r
}
