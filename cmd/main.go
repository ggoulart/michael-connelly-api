package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/ggoulart/michael-connelly-api/internal/characters"
	"github.com/ggoulart/michael-connelly-api/internal/health"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func main() {
	region := ""
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

	adapter := chiadapter.New(r)
	lambda.Start(adapter.ProxyWithContext)
}

func router(booksController *books.Controller, charactersController *characters.Controller, healthController *health.Controller) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/health", func(r chi.Router) {
		r.Get("/", healthController.Health)
	})

	r.Route("/books", func(r chi.Router) {
		r.Post("/", booksController.Create)
		r.Get("/{bookID}", booksController.GetById)
	})

	r.Route("/characters", func(r chi.Router) {
		r.Post("/", charactersController.Create)
		r.Get("/{characterID}", charactersController.GetById)
	})

	return r
}
