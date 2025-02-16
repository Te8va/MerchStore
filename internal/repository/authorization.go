package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Te8va/MerchStore/internal/domain"
	appErrors "github.com/Te8va/MerchStore/internal/errors"
	"github.com/Te8va/MerchStore/pkg/logger"
)

type AuthorizationService struct {
	pool *pgxpool.Pool
}

func NewAuthorizationService(pool *pgxpool.Pool) *AuthorizationService {
	return &AuthorizationService{pool: pool}
}

func (r *AuthorizationService) CreateUser(ctx context.Context, user domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository.CreateUser: failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			logger.Logger().Errorln("CreateUser: failed to rollback transaction:", err)
		}
	}()

	_, err = tx.Exec(ctx, "INSERT INTO users(username, password, token) VALUES($1, $2, $3)", user.Username, user.Password, user.Token)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return appErrors.ErrAlreadyRegistered
		}
		return fmt.Errorf("repository.CreateUser: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository.CreateUser: failed to commit transaction: %w", err)
	}

	return nil
}

func (r *AuthorizationService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User

	err := r.pool.QueryRow(ctx, "SELECT username, password, token FROM users WHERE username = $1", username).
		Scan(&user.Username, &user.Password, &user.Token)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.GetUserByUsername: %w", err)
	}

	return &user, nil
}
