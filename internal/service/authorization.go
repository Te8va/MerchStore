package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Te8va/MerchStore/internal/domain"
	appErrors "github.com/Te8va/MerchStore/internal/errors"
	"github.com/Te8va/MerchStore/pkg/jwt"
)

type Authorization struct {
	repo   domain.AuthorizationRepository
	JWTKey string
}

func NewAuthorization(repo domain.AuthorizationRepository, jwtKey string) *Authorization {
	return &Authorization{repo: repo, JWTKey: jwtKey}
}

func (s *Authorization) RegisterOrAuthenticate(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if err == appErrors.ErrUserNotFound {
			return s.registerNewUser(ctx, username, password)
		}
		return "", fmt.Errorf("service.RegisterOrAuthenticate: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", appErrors.ErrWrongPassword
	}

	return user.Token, nil
}

func (s *Authorization) registerNewUser(ctx context.Context, username, password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("service.registerNewUser: %w", err)
	}

	tokenStr, err := jwt.CreateJWT(username, []byte(s.JWTKey), time.Now().Add(24*time.Hour))
	if err != nil {
		return "", fmt.Errorf("service.registerNewUser: %w", err)
	}

	user := domain.User{
		Username: username,
		Password: string(passwordHash),
		Token:    tokenStr,
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, appErrors.ErrAlreadyRegistered) {
			return "", appErrors.ErrAlreadyRegistered
		}
		return "", fmt.Errorf("service.registerNewUser: %w", err)
	}

	return tokenStr, nil
}
