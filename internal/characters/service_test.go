package characters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	receivedCharacter := Character{Name: "Harry Bosch"}
	tests := []struct {
		name    string
		setup   func(s *StorageMock)
		want    Character
		wantErr error
	}{
		{
			name: "failed to save character",
			setup: func(s *StorageMock) {
				s.On("Save", mock.AnythingOfType("backgroundCtx"), receivedCharacter).Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved character",
			setup: func(s *StorageMock) {
				savedCharacter := Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				s.On("Save", mock.AnythingOfType("backgroundCtx"), receivedCharacter).Return(savedCharacter, nil)
			},
			want: Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(StorageMock)
			tt.setup(storage)

			s := NewService(storage)

			got, err := s.Create(context.Background(), receivedCharacter)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_GetById(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(s *StorageMock)
		want    Character
		wantErr error
	}{
		{
			name: "failed to get character",
			setup: func(s *StorageMock) {
				s.On("GetById", mock.AnythingOfType("backgroundCtx"), "a-random-character-id").Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved character",
			setup: func(s *StorageMock) {
				returnedCharacter := Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				s.On("GetById", mock.AnythingOfType("backgroundCtx"), "a-random-character-id").Return(returnedCharacter, nil)
			},
			want: Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(StorageMock)
			tt.setup(storage)

			s := NewService(storage)

			got, err := s.GetById(context.Background(), "a-random-character-id")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

type StorageMock struct {
	mock.Mock
}

func (s *StorageMock) Save(ctx context.Context, character Character) (Character, error) {
	args := s.Called(ctx, character)
	return args.Get(0).(Character), args.Error(1)
}

func (s *StorageMock) GetById(ctx context.Context, characterID string) (Character, error) {
	args := s.Called(ctx, characterID)
	return args.Get(0).(Character), args.Error(1)
}
