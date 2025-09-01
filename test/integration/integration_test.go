package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ggoulart/michael-connelly-api/cmd/router"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration(t *testing.T) {
	container := setupDynamoDB(t)
	defer container.Terminate(context.Background())

	tests := []struct {
		name       string
		httpMethod string
		targetURL  string
		body       string
		assert     func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "create book",
			httpMethod: http.MethodPost,
			targetURL:  "/books",
			body:       `{"title": "The Black Echo", "year": 1992, "blurb": "Dummy blurb"}`,
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				assert.Equal(t, http.StatusCreated, w.Code)

				id, ok := resp["id"].(string)
				assert.True(t, ok, "expected id field to be a string")
				assert.NotEmpty(t, id, "expected id to be non-empty")
				assert.Equal(t, "The Black Echo", resp["title"])
				assert.Equal(t, float64(1992), resp["year"])
				assert.Equal(t, resp["blurb"], "Dummy blurb")

			},
		},
		{
			name:       "get all books",
			httpMethod: http.MethodGet,
			targetURL:  "/books",
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Len(t, resp, 2)

				assert.Equal(t, "249c03ef-a428-47ec-81f2-af81c2c19397", resp[0]["id"])
				assert.Equal(t, "The Black Ice", resp[0]["title"])
				assert.Equal(t, float64(1993), resp[0]["year"])
				assert.Equal(t, resp[0]["blurb"], "Dummy blurb")

				assert.Equal(t, "5b881832-c740-465e-991e-e7393f90604d", resp[1]["id"])
				assert.Equal(t, "The Concrete Blonde", resp[1]["title"])
				assert.Equal(t, float64(1994), resp[1]["year"])
				assert.Equal(t, resp[1]["blurb"], "Dummy blurb 2")
			},
		},
		{
			name:       "get book by id",
			httpMethod: http.MethodGet,
			targetURL:  "/books/249c03ef-a428-47ec-81f2-af81c2c19397",
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				assert.Equal(t, http.StatusOK, w.Code)

				assert.Equal(t, "249c03ef-a428-47ec-81f2-af81c2c19397", resp["id"])
				assert.Equal(t, "The Black Ice", resp["title"])
				assert.Equal(t, float64(1993), resp["year"])
				assert.Equal(t, resp["blurb"], "Dummy blurb")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.NewRouter()

			req := httptest.NewRequest(tt.httpMethod, tt.targetURL, strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer meu_token_secreto")
			w := httptest.NewRecorder()

			clearTable(t, "books")
			seedBooks(t)

			r.ServeHTTP(w, req)

			tt.assert(t, w)
		})
	}
}

func setupDynamoDB(t *testing.T) testcontainers.Container {
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForListeningPort("8000/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	endpoint, err := container.Endpoint(context.Background(), "http")
	require.NoError(t, err)

	viper.Set("aws.dynamodb.endpoint", endpoint)

	return container
}

func seedBooks(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	require.NoError(t, err)

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(viper.GetString("aws.dynamodb.endpoint"))
		o.Credentials = credentials.NewStaticCredentialsProvider("local", "local", "local")
	})

	books := []map[string]interface{}{
		{
			"id":    "249c03ef-a428-47ec-81f2-af81c2c19397",
			"title": "The Black Ice",
			"year":  1993,
			"blurb": "Dummy blurb",
		},
		{
			"id":    "5b881832-c740-465e-991e-e7393f90604d",
			"title": "The Concrete Blonde",
			"year":  1994,
			"blurb": "Dummy blurb 2",
		},
	}

	for _, book := range books {
		av, err := attributevalue.MarshalMap(book)
		require.NoError(t, err)

		_, err = client.PutItem(context.Background(), &dynamodb.PutItemInput{
			TableName: aws.String("books"),
			Item:      av,
		})
		require.NoError(t, err)
	}
}

func clearTable(t *testing.T, table string) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	require.NoError(t, err)

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(viper.GetString("aws.dynamodb.endpoint"))
		o.Credentials = credentials.NewStaticCredentialsProvider("local", "local", "local")
	})

	out, err := client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName:            aws.String(table),
		ProjectionExpression: aws.String("id"),
	})
	require.NoError(t, err)

	for _, item := range out.Items {
		_, err := client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
			TableName: aws.String(table),
			Key: map[string]types.AttributeValue{
				"id": item["id"],
			},
		})
		require.NoError(t, err)
	}
}
