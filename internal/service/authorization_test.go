package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/Te8va/MerchStore/internal/domain"
	"github.com/Te8va/MerchStore/internal/domain/mocks"
	appErrors "github.com/Te8va/MerchStore/internal/errors"
)

func TestRegisterOrAuthenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthorizationRepository(ctrl)
	testService := NewAuthorization(mockRepo, "test_secret")

	testPassword := "password123"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	validUser := &domain.User{
		Username: "testuser",
		Password: string(passwordHash),
		Token:    "valid_token",
	}

	testCases := []struct {
		name        string
		username    string
		password    string
		mockRepo    func()
		expectedErr error
		expectedTok string
	}{
		{
			name:     "success existing user",
			username: "testuser",
			password: testPassword,
			mockRepo: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(validUser, nil).Times(1)
			},
			expectedErr: nil,
			expectedTok: validUser.Token,
		},
		{
			name:     "wrong password",
			username: "testuser",
			password: "wrongpassword",
			mockRepo: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(validUser, nil).Times(1)
			},
			expectedErr: appErrors.ErrWrongPassword,
			expectedTok: "",
		},
		{
			name:     "user not found, new registration",
			username: "newuser",
			password: testPassword,
			mockRepo: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "newuser").Return(nil, appErrors.ErrUserNotFound).Times(1)
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectedErr: nil,
			expectedTok: "",
		},
		{
			name:     "database error",
			username: "testuser",
			password: testPassword,
			mockRepo: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(nil, errors.New("db error")).Times(1)
			},
			expectedErr: errors.New("service.RegisterOrAuthenticate: db error"),
			expectedTok: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockRepo()

			token, err := testService.RegisterOrAuthenticate(context.Background(), testCase.username, testCase.password)

			if testCase.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
			}
		})
	}
}

func TestRegisterNewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthorizationRepository(ctrl)
	testService := NewAuthorization(mockRepo, "test_secret")

	testCases := []struct {
		name        string
		username    string
		password    string
		mockRepo    func()
		expectedErr error
	}{
		{
			name:     "success registration",
			username: "newuser",
			password: "newpassword123",
			mockRepo: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "newuser").Return(nil, appErrors.ErrUserNotFound).Times(1)
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:     "database error",
			username: "newuser",
			password: "password123",
			mockRepo: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "newuser").Return(nil, errors.New("db error")).Times(1)
			},
			expectedErr: errors.New("service.RegisterOrAuthenticate: db error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockRepo()

			_, err := testService.RegisterOrAuthenticate(context.Background(), testCase.username, testCase.password)

			if testCase.expectedErr != nil {
				require.Error(t, err)
				if errors.Is(err, appErrors.ErrAlreadyRegistered) {
					require.True(t, errors.Is(err, appErrors.ErrAlreadyRegistered))
				} else {
					require.Equal(t, testCase.expectedErr.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
