package characters

import (
	"context"
)

type StorageCharacter interface {
	Save(ctx context.Context, character Character) (Character, error)
	GetById(ctx context.Context, characterID string) (Character, error)
	GetByName(ctx context.Context, characterName string) (Character, error)
}

type Service struct {
	storageCharacter StorageCharacter
}

func NewService(storageCharacter StorageCharacter) *Service {
	return &Service{storageCharacter: storageCharacter}
}

func (s *Service) Create(ctx context.Context, character Character) (Character, error) {
	savedCharacter, err := s.storageCharacter.Save(ctx, character)
	if err != nil {
		return Character{}, err
	}

	return savedCharacter, nil
}

func (s *Service) GetById(ctx context.Context, characterID string) (Character, error) {
	character, err := s.storageCharacter.GetById(ctx, characterID)
	if err != nil {
		return Character{}, err
	}

	return character, nil
}

func (s *Service) GetByName(ctx context.Context, characterName string) (Character, error) {
	character, err := s.storageCharacter.GetByName(ctx, characterName)
	if err != nil {
		return Character{}, err
	}

	return character, nil
}
