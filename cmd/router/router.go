package router

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/ggoulart/michael-connelly-api/internal/characters"
	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
	"github.com/ggoulart/michael-connelly-api/internal/health"
	"github.com/ggoulart/michael-connelly-api/internal/middleware"
	"github.com/ggoulart/michael-connelly-api/internal/series"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type Dependencies struct {
	BooksController      *books.Controller
	CharactersController *characters.Controller
	HealthController     *health.Controller
	SeriesController     *series.Controller
}

func NewRouter() *gin.Engine {
	d := dependencies()
	r := gin.Default()

	r.Use(middleware.Error())

	r.GET("/health", d.HealthController.Health)

	book := r.Group("/books")
	book.POST("", d.BooksController.Create)
	book.GET("/:bookID", d.BooksController.GetById)

	character := r.Group("/characters")
	character.POST("", d.CharactersController.Create)
	character.GET("/:character", d.CharactersController.GetBy)

	series := r.Group("/series")
	series.POST("", d.SeriesController.Create)

	return r
}

func dependencies() Dependencies {
	region := "us-east-1"
	booksTable := "books"
	characterTable := "characters"
	seriesTable := "series"

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	awsDynamoDBClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(viper.GetString("aws.dynamodb.endpoint"))
		o.Credentials = credentials.NewStaticCredentialsProvider("local", "local", "local")
	})
	_, err = awsDynamoDBClient.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatalf("failed to ping DynamoDB: %v", err)
	}

	uuidGenerator := uuid.New

	dynamodbClient := dynamo.NewClient(awsDynamoDBClient, uuidGenerator)
	err = dynamodbClient.CreateTables(ctx)
	if err != nil {
		log.Fatalf("failed create : %v", err)
	}

	healthService := health.NewService(dynamodbClient)
	healthController := health.NewController(healthService)

	booksRepository := books.NewRepository(dynamodbClient, booksTable)
	booksService := books.NewService(booksRepository)
	booksController := books.NewController(booksService)

	charactersRepository := characters.NewRepository(dynamodbClient, characterTable)
	charactersService := characters.NewService(charactersRepository, booksRepository)
	charactersController := characters.NewController(charactersService)

	seriesRepository := series.NewRepository(dynamodbClient, seriesTable)
	seriesService := series.NewService(seriesRepository, booksRepository)
	seriesController := series.NewController(seriesService)

	return Dependencies{
		BooksController:      booksController,
		CharactersController: charactersController,
		HealthController:     healthController,
		SeriesController:     seriesController,
	}
}
