package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Te8va/MerchStore/internal/domain"
	"github.com/Te8va/MerchStore/internal/domain/mocks"
	appErrors "github.com/Te8va/MerchStore/internal/errors"
	"github.com/Te8va/MerchStore/internal/handler"
	"github.com/Te8va/MerchStore/internal/service"
	"github.com/Te8va/MerchStore/pkg/jwt"
)

func TestSendCoinHandler_BusinessScenarios(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMerchSrv := mocks.NewMockMerchService(ctrl)
	mockAuthRepo := mocks.NewMockAuthorizationRepository(ctrl)

	jwtKey := "test_jwt_key"

	merchHandler := handler.NewMerchHandler(mockMerchSrv, jwtKey)
	authService := service.NewAuthorization(mockAuthRepo, jwtKey)
	authHandler := handler.NewAuthorizationHandler(authService)

	senderToken, err := jwt.CreateJWT("sender", []byte(jwtKey), time.Now().Add(24*time.Hour))
	assert.NoError(t, err)

	receiverToken, err := jwt.CreateJWT("receiver", []byte(jwtKey), time.Now().Add(24*time.Hour))
	assert.NoError(t, err)

	mockAuthRepo.EXPECT().GetUserByUsername(gomock.Any(), "sender").Return(nil, appErrors.ErrUserNotFound).Times(1)
	mockAuthRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockAuthRepo.EXPECT().GetUserByUsername(gomock.Any(), "sender").Return(&domain.User{
		Username: "sender",
		Password: "$2a$10$examplehash",
		Token:    senderToken,
	}, nil).AnyTimes()

	mockAuthRepo.EXPECT().GetUserByUsername(gomock.Any(), "receiver").Return(nil, appErrors.ErrUserNotFound).Times(1)
	mockAuthRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockAuthRepo.EXPECT().GetUserByUsername(gomock.Any(), "receiver").Return(&domain.User{
		Username: "receiver",
		Password: "$2a$10$examplehash",
		Token:    receiverToken,
	}, nil).AnyTimes()

	// Регистрация пользователей (отправитель и получатель)
	users := []string{"sender", "receiver"}
	for _, username := range users {
		rr := httptest.NewRecorder()
		reqBody := map[string]interface{}{
			"username": username,
			"password": "password",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		authHandler.AuthHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	mockAuthRepo.EXPECT().GetUserByUsername(gomock.Any(), "unknown").Return(nil, appErrors.ErrUserNotFound).AnyTimes()

	// Успешная отправка монет
	mockMerchSrv.EXPECT().SendCoin(gomock.Any(), "sender", "receiver", 10).Return(nil)
	rr := httptest.NewRecorder()
	reqBody := map[string]interface{}{
		"toUser": "receiver",
		"amount": 10,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/send-coin", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+senderToken)
	req.Header.Set("Content-Type", "application/json")
	merchHandler.SendCoinHandler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Ошибка: отправка монет без регистрации получателя
	mockMerchSrv.EXPECT().SendCoin(gomock.Any(), "sender", "unknown", 10).Return(appErrors.ErrUserNotFound)
	rr = httptest.NewRecorder()
	reqBody = map[string]interface{}{
		"toUser": "unknown",
		"amount": 10,
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest(http.MethodPost, "/api/send-coin", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+senderToken)
	req.Header.Set("Content-Type", "application/json")
	merchHandler.SendCoinHandler(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Ошибка: отправка монет самому себе
	rr = httptest.NewRecorder()
	reqBody = map[string]interface{}{
		"toUser": "sender",
		"amount": 10,
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest(http.MethodPost, "/api/send-coin", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+senderToken)
	req.Header.Set("Content-Type", "application/json")
	merchHandler.SendCoinHandler(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Ошибка: недостаточно средств
	mockMerchSrv.EXPECT().SendCoin(gomock.Any(), "sender", "receiver", 9999).Return(appErrors.ErrInsufficientBalance)
	rr = httptest.NewRecorder()
	reqBody = map[string]interface{}{
		"toUser": "receiver",
		"amount": 9999,
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest(http.MethodPost, "/api/send-coin", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+senderToken)
	req.Header.Set("Content-Type", "application/json")
	merchHandler.SendCoinHandler(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
