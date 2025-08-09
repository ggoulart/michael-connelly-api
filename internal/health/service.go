package health

import "context"

type DynamoClient interface {
	Ping(ctx context.Context) error
}
type Service struct {
	dynamo DynamoClient
}

func NewService(dynamo DynamoClient) *Service {
	return &Service{dynamo: dynamo}
}

func (s *Service) Health(ctx context.Context) map[string]bool {
	servicesHealth := map[string]bool{"api": true}

	if err := s.dynamo.Ping(ctx); err != nil {
		servicesHealth["db"] = false
	} else {
		servicesHealth["db"] = true
	}

	return servicesHealth
}
