package domain

import (
	"context"
)

//go:generate mockgen -destination=mocks/repo_mock.gen.go -package=mocks . MerchRepository
type MerchRepository interface {
	GetMerchPrice(ctx context.Context, item string) (int, error)
	GetUserBalance(ctx context.Context, username string) (int, error)
	UpdateUserBalance(ctx context.Context, username string, newBalance int) error
	SavePurchase(ctx context.Context, username, item string, price int) error
	UserExists(ctx context.Context, username string) (bool, error)
	TransferCoins(ctx context.Context, fromUser, toUser string, amount int) error
	GetUserInventory(ctx context.Context, username string) ([]string, error)
	GetUserTransactionHistory(ctx context.Context, username string) (CoinHistory, error)
	GetUserPurchases(ctx context.Context, username string) ([]InventoryItem, error)
}

//go:generate mockgen -destination=mocks/merch_service_mock.gen.go -package=mocks . MerchService
type MerchService interface {
	BuyMerch(ctx context.Context, username, item string) error
	SendCoin(ctx context.Context, fromUser, toUser string, amount int) error
	GetUserInfo(ctx context.Context, username string) (UserInfo, error)
}
