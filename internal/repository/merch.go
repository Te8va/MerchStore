package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Te8va/MerchStore/internal/domain"
	"github.com/Te8va/MerchStore/pkg/logger"
)

type MerchService struct {
	pool *pgxpool.Pool
}

func NewMerchService(pool *pgxpool.Pool) *MerchService {
	return &MerchService{pool: pool}
}

func (r *MerchService) GetUserBalance(ctx context.Context, username string) (int, error) {
	var balance int
	err := r.pool.QueryRow(ctx, "SELECT balance FROM users WHERE username = $1", username).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("repository.GetUserBalance: %w", err)
	}
	return balance, nil
}

func (r *MerchService) TransferCoins(ctx context.Context, fromUser, toUser string, amount int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository.TransferCoins: could not begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			logger.Logger().Errorln("TransferCoins: failed to rollback transaction:", err)
		}
	}()

	_, err = tx.Exec(ctx, "UPDATE users SET balance = balance - $1 WHERE username=$2", amount, fromUser)
	if err != nil {
		return fmt.Errorf("repository.TransferCoins: could not update sender balance: %w", err)
	}

	_, err = tx.Exec(ctx, "UPDATE users SET balance = balance + $1 WHERE username=$2", amount, toUser)
	if err != nil {
		return fmt.Errorf("repository.TransferCoins: could not update receiver balance: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (from_user, to_user, amount) 
		VALUES ($1, $2, $3)`, fromUser, toUser, amount)
	if err != nil {
		return fmt.Errorf("repository.TransferCoins: could not insert transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Logger().Errorln("TransferCoins: failed to commit transaction:", err)
		return fmt.Errorf("repository.TransferCoins: could not commit transaction: %w", err)
	}

	return nil
}

func (r *MerchService) UpdateUserBalance(ctx context.Context, username string, newBalance int) error {
	_, err := r.pool.Exec(ctx, "UPDATE users SET balance = $1 WHERE username = $2", newBalance, username)
	if err != nil {
		return fmt.Errorf("repository.UpdateUserBalance: %w", err)
	}
	return nil
}

func (r *MerchService) GetMerchPrice(ctx context.Context, item string) (int, error) {
	var price int
	err := r.pool.QueryRow(ctx, "SELECT price FROM merch WHERE item_name = $1", item).Scan(&price)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("repository.GetMerchPrice: item not found")
		}
		return 0, fmt.Errorf("repository.GetMerchPrice: %w", err)
	}
	return price, nil
}

func (r *MerchService) SavePurchase(ctx context.Context, username, item string, price int) error {
	var userId int
	err := r.pool.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userId)
	if err != nil {
		return fmt.Errorf("repository.SavePurchase: could not find user ID: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO inventory (user_id, item_name, quantity) 
		VALUES ($1, $2, 1) 
		ON CONFLICT (user_id, item_name) 
		DO UPDATE SET quantity = inventory.quantity + 1`, userId, item)
	if err != nil {
		return fmt.Errorf("repository.SavePurchase: could not update inventory: %w", err)
	}

	return nil
}

func (r *MerchService) UserExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("repository.UserExists: %w", err)
	}
	return exists, nil
}

func (r *MerchService) GetUserInventory(ctx context.Context, username string) ([]string, error) {
	var inventory []string

	var userId int
	err := r.pool.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userId)
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserInventory: could not find user ID: %w", err)
	}

	rows, err := r.pool.Query(ctx, "SELECT item_name FROM inventory WHERE user_id = $1", userId)
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserInventory: could not retrieve inventory: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var itemName string
		if err := rows.Scan(&itemName); err != nil {
			return nil, fmt.Errorf("repository.GetUserInventory: could not scan item name: %w", err)
		}
		inventory = append(inventory, itemName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.GetUserInventory: error reading rows: %w", err)
	}

	return inventory, nil
}

func (r *MerchService) GetUserTransactionHistory(ctx context.Context, username string) (domain.CoinHistory, error) {
	var history domain.CoinHistory

	rows, err := r.pool.Query(ctx, `
		SELECT from_user, amount FROM transactions WHERE to_user = $1`, username)
	if err != nil {
		return history, fmt.Errorf("repository.GetUserTransactionHistory: could not find user ID: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var received domain.ReceivedTransaction
		err := rows.Scan(&received.FromUser, &received.Amount)
		if err != nil {
			return history, fmt.Errorf("repository.GetUserTransactionHistory: could not scan received transaction: %w", err)
		}
		history.Received = append(history.Received, received)
	}
	if err = rows.Err(); err != nil {
		return history, fmt.Errorf("repository.GetUserTransactionHistory: error reading received rows: %w", err)
	}

	rows, err = r.pool.Query(ctx, `
		SELECT to_user, amount
		FROM transactions
		WHERE from_user = $1
	`, username)
	if err != nil {
		return history, fmt.Errorf("repository.GetUserTransactionHistory: could not get sent transactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sent domain.SentTransaction
		err := rows.Scan(&sent.ToUser, &sent.Amount)
		if err != nil {
			return history, fmt.Errorf("repository.GetUserTransactionHistory: could not scan sent transaction: %w", err)
		}
		history.Sent = append(history.Sent, sent)
	}
	if err = rows.Err(); err != nil {
		return history, fmt.Errorf("repository.GetUserTransactionHistory: error reading sent rows: %w", err)
	}

	return history, nil
}

func (r *MerchService) GetUserPurchases(ctx context.Context, username string) ([]domain.InventoryItem, error) {
	var inventory []domain.InventoryItem

	rows, err := r.pool.Query(ctx, `
		SELECT item_name, quantity 
		FROM inventory 
		INNER JOIN users ON inventory.user_id = users.id
		WHERE users.username = $1
	`, username)
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserPurchases: could not retrieve purchases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.InventoryItem
		if err := rows.Scan(&item.Name, &item.Quantity); err != nil {
			return nil, fmt.Errorf("repository.GetUserPurchases: could not scan purchase: %w", err)
		}
		inventory = append(inventory, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.GetUserPurchases: error reading rows: %w", err)
	}

	return inventory, nil
}
