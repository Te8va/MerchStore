package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Te8va/MerchStore/internal/domain"
	appErrors "github.com/Te8va/MerchStore/internal/errors"
)

type Merch struct {
	repo domain.MerchRepository
}

func NewMerch(repo domain.MerchRepository) *Merch {
	return &Merch{repo: repo}
}

func (s *Merch) SendCoin(ctx context.Context, fromUser, toUser string, amount int) error {
	userExists, err := s.repo.UserExists(ctx, toUser)
	if err != nil {
		return fmt.Errorf("service.SendCoin: %w", err)
	}
	if !userExists {
		return appErrors.ErrUserNotFound
	}

	balance, err := s.repo.GetUserBalance(ctx, fromUser)
	if err != nil {
		return fmt.Errorf("service.SendCoin: %w", err)
	}
	if balance < amount {
		return appErrors.ErrInsufficientBalance
	}

	err = s.repo.TransferCoins(ctx, fromUser, toUser, amount)
	if err != nil {
		return fmt.Errorf("could not complete transaction: %w", err)
	}

	return nil
}

func (s *Merch) GetUserInfo(ctx context.Context, username string) (domain.UserInfo, error) {
	var info domain.UserInfo

	userExists, err := s.repo.UserExists(ctx, username)
	if err != nil {
		return info, fmt.Errorf("service.GetUserInfo: %w", err)
	}
	if !userExists {
		return info, appErrors.ErrUserNotFound
	}

	balance, err := s.repo.GetUserBalance(ctx, username)
	if err != nil {
		return info, fmt.Errorf("service.GetUserInfo: %w", err)
	}
	info.Coins = balance

	inventory, err := s.repo.GetUserPurchases(ctx, username)
	if err != nil {
		return info, fmt.Errorf("service.GetUserInfo: %w", err)
	}
	info.Inventory = inventory

	history, err := s.repo.GetUserTransactionHistory(ctx, username)
	if err != nil {
		return info, fmt.Errorf("service.GetUserInfo: %w", err)
	}
	info.CoinHistory = history

	return info, nil
}

func (s *Merch) BuyMerch(ctx context.Context, username, item string) error {
	userExists, err := s.repo.UserExists(ctx, username)
	if err != nil {
		return fmt.Errorf("service.BuyMerch: %w", err)
	}
	if !userExists {
		return appErrors.ErrUserNotFound
	}

	price, err := s.repo.GetMerchPrice(ctx, item)
	if err != nil {
		if strings.Contains(err.Error(), "item not found") {
			return appErrors.ErrItemNotFound
		}
		return fmt.Errorf("service.BuyMerch: %w", err)
	}

	balance, err := s.repo.GetUserBalance(ctx, username)
	if err != nil {
		return fmt.Errorf("service.BuyMerch: %w", err)
	}

	if balance < price {
		return appErrors.ErrInsufficientBalance
	}

	newBalance := balance - price
	if err := s.repo.UpdateUserBalance(ctx, username, newBalance); err != nil {
		return fmt.Errorf("service.BuyMerch: %w", err)
	}

	if err := s.repo.SavePurchase(ctx, username, item, price); err != nil {
		return fmt.Errorf("service.BuyMerch: %w", err)
	}

	return nil
}
