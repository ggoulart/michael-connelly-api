package characters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	receivedCharacter := Character{Name: "Harry Bosch"}
	tests := []struct {
		name    string
		setup   func(s *StorageCharacterMock)
		want    Character
		wantErr error
	}{
		{
			name: "failed to save character",
			setup: func(s *StorageCharacterMock) {
				s.On("Save", ctx, receivedCharacter).Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved character",
			setup: func(s *StorageCharacterMock) {
				savedCharacter := Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				s.On("Save", ctx, receivedCharacter).Return(savedCharacter, nil)
			},
			want: Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			tt.setup(storageCharacter)

			s := NewService(storageCharacter)

			got, err := s.Create(ctx, receivedCharacter)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_GetById(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(s *StorageCharacterMock)
		want    Character
		wantErr error
	}{
		{
			name: "failed to get character",
			setup: func(s *StorageCharacterMock) {
				s.On("GetById", ctx, "a-random-character-id").Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved character",
			setup: func(s *StorageCharacterMock) {
				returnedCharacter := Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				s.On("GetById", ctx, "a-random-character-id").Return(returnedCharacter, nil)
			},
			want: Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			tt.setup(storageCharacter)

			s := NewService(storageCharacter)

			got, err := s.GetById(ctx, "a-random-character-id")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_GetByName(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(s *StorageCharacterMock)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to get character",
			setup: func(m *StorageCharacterMock) {
				m.On("GetByName", ctx, "Harry Bosch").Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully get character",
			setup: func(m *StorageCharacterMock) {
				character := Character{ID: "random-id", Name: "Harry Bosch"}
				m.On("GetByName", ctx, "Harry Bosch").Return(character, nil)
			},
			want: Character{ID: "random-id", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			tt.setup(storageCharacter)

			s := NewService(storageCharacter)
			got, err := s.GetByName(ctx, "Harry Bosch")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

type StorageCharacterMock struct {
	mock.Mock
}

func (s *StorageCharacterMock) GetByName(ctx context.Context, characterName string) (Character, error) {
	args := s.Called(ctx, characterName)
	return args.Get(0).(Character), args.Error(1)
}

func (s *StorageCharacterMock) Save(ctx context.Context, character Character) (Character, error) {
	args := s.Called(ctx, character)
	return args.Get(0).(Character), args.Error(1)
}

func (s *StorageCharacterMock) GetById(ctx context.Context, characterID string) (Character, error) {
	args := s.Called(ctx, characterID)
	return args.Get(0).(Character), args.Error(1)
}
