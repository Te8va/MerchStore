package domain

import "context"

//go:generate mockgen -destination=mocks/authorization_repo_mock.gen.go -package=mocks . AuthorizationRepository
type AuthorizationRepository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

//go:generate mockgen -destination=mocks/authorization_service_mock.gen.go -package=mocks . AuthorizationService
type AuthorizationService interface {
	RegisterOrAuthenticate(ctx context.Context, username, password string) (string, error)
}
