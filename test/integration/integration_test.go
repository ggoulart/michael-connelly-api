package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			body:       `{"title": "The Black Echo", "year": 1992, "blurb": "For LAPD homicide cop Harry Bosch — hero, maverick, nighthawk — the body in the drainpipe at Mulholland dam is more than another anonymous statistic.  This one is personal. The dead man, Billy Meadows, was a fellow Vietnam “tunnel rat” who fought side by side with him in a nightmare underground war that brought them to the depths of hell.  Now, Bosch is about to relive the horrors of Nam.  From a dangerous maze of blind alleys to a daring criminal heist beneath the city to the tortuous link that must be uncovered, his survival instincts will once again be tested to their limit. Joining with an enigmatic female FBI agent, pitted against enemies within his own department, Bosch must make the agonizing choice between justice and vengeance, as he tracks down a killer whose true face will shock him. The Black Echo won the Edgar Award for Best First Mystery Novel awarded by the Mystery Writers of America."}`,
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusCreated, w.Code)

				id, ok := resp["id"].(string)
				require.True(t, ok, "expected id field to be a string")
				require.NotEmpty(t, id, "expected id to be non-empty")
				assert.Equal(t, "The Black Echo", resp["title"])
				assert.Equal(t, float64(1992), resp["year"])
				assert.Contains(t, resp["blurb"], "For LAPD homicide cop Harry Bosch")

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := router.NewRouter()

			req := httptest.NewRequest(tt.httpMethod, tt.targetURL, strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer meu_token_secreto")
			w := httptest.NewRecorder()

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

	assert.NoError(t, err)

	endpoint, err := container.Endpoint(context.Background(), "http")
	require.NoError(t, err)

	viper.Set("aws.dynamodb.endpoint", endpoint)

	return container
}
