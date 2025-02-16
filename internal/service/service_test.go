package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/MerchStore/internal/domain"
	"github.com/Te8va/MerchStore/internal/domain/mocks"
	appErrors "github.com/Te8va/MerchStore/internal/errors"
)

func TestSendCoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMerchRepository(ctrl)
	merchService := NewMerch(mockRepo)

	testCases := []struct {
		name        string
		fromUser    string
		toUser      string
		amount      int
		mockRepo    func()
		expectedErr error
	}{
		{
			name:     "success",
			fromUser: "user1",
			toUser:   "user2",
			amount:   100,
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user2").Return(true, nil).Times(1)
				mockRepo.EXPECT().GetUserBalance(context.Background(), "user1").Return(200, nil).Times(1)
				mockRepo.EXPECT().TransferCoins(context.Background(), "user1", "user2", 100).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:     "user not found",
			fromUser: "user1",
			toUser:   "nonexistentUser",
			amount:   100,
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "nonexistentUser").Return(false, nil).Times(1)
			},
			expectedErr: appErrors.ErrUserNotFound,
		},
		{
			name:     "insufficient balance",
			fromUser: "user1",
			toUser:   "user2",
			amount:   1000,
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user2").Return(true, nil).Times(1)
				mockRepo.EXPECT().GetUserBalance(context.Background(), "user1").Return(100, nil).Times(1)
			},
			expectedErr: appErrors.ErrInsufficientBalance,
		},
		{
			name:     "db error",
			fromUser: "user1",
			toUser:   "user2",
			amount:   100,
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user2").Return(false, errors.New("db error")).Times(1)
			},
			expectedErr: errors.New("service.SendCoin: db error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockRepo()

			err := merchService.SendCoin(context.Background(), testCase.fromUser, testCase.toUser, testCase.amount)

			if testCase.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestGetUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMerchRepository(ctrl)
	merchService := NewMerch(mockRepo)

	testCases := []struct {
		name         string
		username     string
		mockRepo     func()
		expectedErr  error
		expectedInfo domain.UserInfo
	}{
		{
			name:     "success",
			username: "user1",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user1").Return(true, nil).Times(1)
				mockRepo.EXPECT().GetUserBalance(context.Background(), "user1").Return(500, nil).Times(1)
				mockRepo.EXPECT().GetUserPurchases(context.Background(), "user1").Return([]domain.InventoryItem{
					{Name: "item1", Quantity: 2},
					{Name: "item2", Quantity: 1},
				}, nil).Times(1)
				mockRepo.EXPECT().GetUserTransactionHistory(context.Background(), "user1").Return(domain.CoinHistory{
					Received: []domain.ReceivedTransaction{{FromUser: "user2", Amount: 100}},
					Sent:     []domain.SentTransaction{{ToUser: "user3", Amount: 50}},
				}, nil).Times(1)
			},
			expectedErr: nil,
			expectedInfo: domain.UserInfo{
				Coins: 500,
				Inventory: []domain.InventoryItem{
					{Name: "item1", Quantity: 2},
					{Name: "item2", Quantity: 1},
				},
				CoinHistory: domain.CoinHistory{
					Received: []domain.ReceivedTransaction{{FromUser: "user2", Amount: 100}},
					Sent:     []domain.SentTransaction{{ToUser: "user3", Amount: 50}},
				},
			},
		},
		{
			name:     "user not found",
			username: "nonexistentUser",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "nonexistentUser").Return(false, nil).Times(1)
			},
			expectedErr: appErrors.ErrUserNotFound,
		},
		{
			name:     "db error",
			username: "user1",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user1").Return(false, errors.New("db error")).Times(1)
			},
			expectedErr: errors.New("service.GetUserInfo: db error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockRepo()

			info, err := merchService.GetUserInfo(context.Background(), testCase.username)

			if testCase.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedInfo, info)
			}
		})
	}
}

func TestBuyMerch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMerchRepository(ctrl)
	merchService := NewMerch(mockRepo)

	testCases := []struct {
		name        string
		username    string
		item        string
		mockRepo    func()
		expectedErr error
	}{
		{
			name:     "success",
			username: "user1",
			item:     "merch1",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user1").Return(true, nil).Times(1)
				mockRepo.EXPECT().GetMerchPrice(context.Background(), "merch1").Return(100, nil).Times(1)
				mockRepo.EXPECT().GetUserBalance(context.Background(), "user1").Return(200, nil).Times(1)
				mockRepo.EXPECT().UpdateUserBalance(context.Background(), "user1", 100).Times(1)
				mockRepo.EXPECT().SavePurchase(context.Background(), "user1", "merch1", 100).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:     "user not found",
			username: "nonexistentUser",
			item:     "merch1",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "nonexistentUser").Return(false, nil).Times(1)
			},
			expectedErr: appErrors.ErrUserNotFound,
		},
		{
			name:     "insufficient balance",
			username: "user1",
			item:     "merch1",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user1").Return(true, nil).Times(1)
				mockRepo.EXPECT().GetMerchPrice(context.Background(), "merch1").Return(200, nil).Times(1)
				mockRepo.EXPECT().GetUserBalance(context.Background(), "user1").Return(100, nil).Times(1)
			},
			expectedErr: appErrors.ErrInsufficientBalance,
		},
		{
			name:     "item not found",
			username: "user1",
			item:     "nonexistentItem",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user1").Return(true, nil).Times(1)
				mockRepo.EXPECT().GetMerchPrice(context.Background(), "nonexistentItem").Return(0, errors.New("item not found")).Times(1)
			},
			expectedErr: appErrors.ErrItemNotFound,
		},
		{
			name:     "db error",
			username: "user1",
			item:     "merch1",
			mockRepo: func() {
				mockRepo.EXPECT().UserExists(context.Background(), "user1").Return(true, nil).Times(1)
				mockRepo.EXPECT().GetMerchPrice(context.Background(), "merch1").Return(100, nil).Times(1)
				mockRepo.EXPECT().GetUserBalance(context.Background(), "user1").Return(200, nil).Times(1)
				mockRepo.EXPECT().UpdateUserBalance(context.Background(), "user1", 100).Times(1)
				mockRepo.EXPECT().SavePurchase(context.Background(), "user1", "merch1", 100).Return(errors.New("db error")).Times(1)
			},
			expectedErr: errors.New("service.BuyMerch: db error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockRepo()

			err := merchService.BuyMerch(context.Background(), testCase.username, testCase.item)

			if testCase.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
