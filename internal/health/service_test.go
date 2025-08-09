package health

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Health(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name  string
		setup func(*MockDynamoDBClient)
		want  map[string]bool
	}{
		{
			name: "when failed to ping dynamodb",
			setup: func(m *MockDynamoDBClient) {
				m.On("Ping", ctx).Return(assert.AnError)
			},
			want: map[string]bool{"api": true, "db": false},
		},
		{
			name: "when successful to ping dynamodb",
			setup: func(m *MockDynamoDBClient) {
				m.On("Ping", ctx).Return(nil)
			},
			want: map[string]bool{"api": true, "db": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			s := NewService(mockDynamoDBClient)

			got := s.Health(ctx)

			assert.Equal(t, tt.want, got)
		})
	}
}

type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
