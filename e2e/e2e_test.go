package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/Te8va/MerchStore/internal/config"
	"github.com/Te8va/MerchStore/internal/domain"
	"github.com/Te8va/MerchStore/internal/handler"
	"github.com/Te8va/MerchStore/internal/middleware"
	"github.com/Te8va/MerchStore/internal/repository"
	"github.com/Te8va/MerchStore/internal/service"
	"github.com/Te8va/MerchStore/pkg/jwt"
)

var (
	httpServer *http.Server
	merchRepo  *repository.MerchService
	authRepo   *repository.AuthorizationService
	once       = &sync.Once{}
	pool       *pgxpool.Pool
)

func NewServer() *http.Server {
	once.Do(func() {
		cfg := config.Config{}
		if err := env.Parse(&cfg); err != nil {
			log.Fatalf("Failed to parse env: %v", err)
		}

		cfg.PostgresConn = "postgres://merch:merch@localhost:5433/merch?sslmode=disable"

		m, err := migrate.New("file://../migrations", cfg.PostgresConn)
		if err != nil {
			log.Fatalf("Error initializing migrations: %v", err)
		}

		err = repository.ApplyMigrations(m)
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error applying migrations: %v", err)
		}

		log.Println("Migrations applied successfully")

		pool, err = repository.GetPgxPool(cfg.PostgresConn)
		if err != nil {
			log.Fatalf("Error creating PostgreSQL pool: %v", err)
		}

		log.Println("Postgres connection pool created")

		merchRepo = repository.NewMerchService(pool)
		merchService := service.NewMerch(merchRepo)
		merchHandler := handler.NewMerchHandler(merchService, cfg.JWTKey)

		authRepo = repository.NewAuthorizationService(pool)
		authService := service.NewAuthorization(authRepo, cfg.JWTKey)
		authHandler := handler.NewAuthorizationHandler(authService)

		mux := http.NewServeMux()

		mux.Handle("GET /api/info", middleware.Log(http.HandlerFunc(merchHandler.GetUserInfoHandler)))
		mux.Handle("POST /api/sendCoin", middleware.Log(http.HandlerFunc(merchHandler.SendCoinHandler)))
		mux.Handle("GET /api/buy/{item}", middleware.Log(http.HandlerFunc(merchHandler.BuyMerchHandler)))
		mux.Handle("POST /api/auth", middleware.Log(http.HandlerFunc(authHandler.AuthHandler)))

		server := &http.Server{
			Addr:     fmt.Sprintf("%s:%d", cfg.ServiceHost, cfg.ServicePort),
			ErrorLog: log.New(os.Stdout, "", 0),
			Handler:  mux,
		}

		go func() {
			log.Printf("Server started, listening on port %d", cfg.ServicePort)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe failed: %v", err)
			}
		}()

		httpServer = server
	})

	return httpServer
}

func TestMerchPurchaseSuccessful(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080"

	// Регистрация и получение токена
	userData := map[string]string{
		"username": "buyer",
		"password": "password123",
	}
	reqBody, err := json.Marshal(userData)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var authResponse map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&authResponse)
	require.NoError(t, err)
	token, ok := authResponse["token"].(string)
	require.True(t, ok, "Expected token to be present in response")

	// Проверка информации о пользователе
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client = http.Client{Timeout: 5 * time.Second}
	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Покупка мерча
	item := "cup"
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", serverAddr, item), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Проверка баланса после покупки
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMerchPurchaseInvalidItemNoMoney(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080"

	// Регистрация и получение токена
	userData := map[string]string{
		"username": "buyer",
		"password": "password123",
	}
	reqBody, err := json.Marshal(userData)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var authResponse map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&authResponse)
	require.NoError(t, err)
	token, ok := authResponse["token"].(string)
	require.True(t, ok, "Expected token to be present in response")

	// Успешная покупка розового худи
	item := "pink-hoody"
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", serverAddr, item), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Проверка баланса после покупки
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Покупка несуществующего предмета
	item = "invaliditem"
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", serverAddr, item), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Успешная покупка розового худи
	item = "pink-hoody"
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", serverAddr, item), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Покупка с недостаточным балансом
	item = "pink-hoody"
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%s", serverAddr, item), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestSendCoinSuccessful(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080"

	// Авторизация отправителя
	userData := map[string]string{
		"username": "sender",
		"password": "pass123",
	}
	reqBody, err := json.Marshal(userData)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var authResponse map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&authResponse)
	require.NoError(t, err)
	token, ok := authResponse["token"].(string)
	require.True(t, ok, "Expected token to be present in response")

	// Авторизация получателя
	userData = map[string]string{
		"username": "receiver",
		"password": "pass12",
	}
	reqBody, err = json.Marshal(userData)
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client = http.Client{Timeout: 5 * time.Second}
	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	// Передача монеток от отправителя к получателю
	transferData := map[string]interface{}{
		"toUser": "receiver",
		"amount": 10,
	}
	reqBody, err = json.Marshal(transferData)
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/sendCoin", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Проверка баланса после перевода
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestSendCoinNotEnogh(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080"

	// Авторизация отправителя
	userData := map[string]string{
		"username": "sender",
		"password": "pass123",
	}
	reqBody, err := json.Marshal(userData)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var authResponse map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&authResponse)
	require.NoError(t, err)
	token, ok := authResponse["token"].(string)
	require.True(t, ok, "Expected token to be present in response")

	// Авторизация получателя
	userData = map[string]string{
		"username": "receiver",
		"password": "pass12",
	}
	reqBody, err = json.Marshal(userData)
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client = http.Client{Timeout: 5 * time.Second}
	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	// Передача слишком много монет от отправителя к получателю
	transferData := map[string]interface{}{
		"toUser": "receiver",
		"amount": 10000,
	}
	reqBody, err = json.Marshal(transferData)
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/sendCoin", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Проверка баланса после перевода
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGetUserInfoForgotAll(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080"

	// Регистрация пользователя и получение токена
	userData := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	reqBody, err := json.Marshal(userData)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", serverAddr), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var authResponse map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&authResponse)
	require.NoError(t, err)
	token, ok := authResponse["token"].(string)
	require.True(t, ok, "Expected token to be present in response")

	// Попытка получения информации без токена
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusUnauthorized, res.StatusCode)

	// Попытка получения информации с неверным токеном
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalidtoken")

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusUnauthorized, res.StatusCode)

	// Успешное получение информации о пользователе
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", serverAddr), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestSendCoinHandler(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080/api/sendCoin"

	initialUsers := []struct {
		username string
		password string
		token    string
	}{
		{"user1", "password123", ""},
		{"user2", "password456", ""},
	}

	jwtKey := "supermegasecret"

	for i := range initialUsers {
		token, err := jwt.CreateJWT(initialUsers[i].username, []byte(jwtKey), time.Now().Add(24*time.Hour))
		require.NoError(t, err)
		initialUsers[i].token = token

		err = authRepo.CreateUser(context.Background(), domain.User{
			Username: initialUsers[i].username,
			Password: initialUsers[i].password,
			Token:    initialUsers[i].token,
		})
		require.NoError(t, err)
	}

	testCases := []struct {
		name           string
		fromUser       string
		toUser         string
		amount         int
		wantStatusCode int
	}{
		{
			name:           "Successful Coin Transfer",
			fromUser:       "user1",
			toUser:         "user2",
			amount:         10,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Insufficient Balance",
			fromUser:       "user1",
			toUser:         "user2",
			amount:         9999,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(map[string]interface{}{
				"toUser": testCase.toUser,
				"amount": testCase.amount,
			})
			require.NoError(t, err)

			var userToken string
			for _, user := range initialUsers {
				if user.username == testCase.fromUser {
					userToken = user.token
					break
				}
			}

			req, err := http.NewRequest(http.MethodPost, serverAddr, bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+userToken)

			client := http.Client{Timeout: 5 * time.Second}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, testCase.wantStatusCode, res.StatusCode)
		})
	}
}

func TestGetUserInfoHandler(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080/api/info"

	initialUsers := []struct {
		username string
		password string
		token    string
	}{
		{"user1", "password123", ""},
		{"user2", "password456", ""},
	}

	jwtKey := "supermegasecret"
	for i := range initialUsers {
		token, err := jwt.CreateJWT(initialUsers[i].username, []byte(jwtKey), time.Now().Add(24*time.Hour))
		require.NoError(t, err)
		initialUsers[i].token = token

		err = authRepo.CreateUser(context.Background(), domain.User{
			Username: initialUsers[i].username,
			Password: initialUsers[i].password,
			Token:    initialUsers[i].token,
		})
		require.NoError(t, err)
	}

	testCases := []struct {
		name           string
		token          string
		wantStatusCode int
	}{
		{
			name:           "Successful User Info",
			token:          initialUsers[0].token,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Unauthorized Access",
			token:          "invalidtoken",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, serverAddr, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+testCase.token)
			req.Header.Set("Content-Type", "application/json")

			client := http.Client{Timeout: 5 * time.Second}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, testCase.wantStatusCode, res.StatusCode)
		})
	}
}

func TestBuyMerchHandler(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080/api/buy/"

	initialUsers := []struct {
		username string
		password string
		token    string
	}{
		{"user1", "password123", ""},
		{"lowbalanceuser", "password123", ""},
	}

	jwtKey := "supermegasecret"
	for i := range initialUsers {
		token, err := jwt.CreateJWT(initialUsers[i].username, []byte(jwtKey), time.Now().Add(24*time.Hour))
		require.NoError(t, err)
		initialUsers[i].token = token

		err = authRepo.CreateUser(context.Background(), domain.User{
			Username: initialUsers[i].username,
			Password: initialUsers[i].password,
			Token:    initialUsers[i].token,
		})
		require.NoError(t, err)

		if initialUsers[i].username == "user1" {
			err = merchRepo.UpdateUserBalance(context.Background(), "user1", 100)
			require.NoError(t, err)
		} else {
			err = merchRepo.UpdateUserBalance(context.Background(), "lowbalanceuser", 1)
			require.NoError(t, err)
		}
	}

	testCases := []struct {
		name           string
		username       string
		item           string
		token          string
		wantStatusCode int
	}{
		{
			name:           "Successful Purchase",
			username:       "user1",
			item:           "cup",
			token:          initialUsers[0].token,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Insufficient Balance",
			username:       "lowbalanceuser",
			item:           "cup",
			token:          initialUsers[1].token,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "Item Not Found",
			username:       "user1",
			item:           "invaliditem",
			token:          initialUsers[0].token,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", serverAddr, testCase.item), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+testCase.token)
			req.Header.Set("Content-Type", "application/json")

			client := http.Client{Timeout: 5 * time.Second}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, testCase.wantStatusCode, res.StatusCode)
		})
	}
}

func TestAuthHandler(t *testing.T) {
	NewServer()
	defer truncateTest(context.Background(), pool)

	serverAddr := "http://localhost:8080/api/auth"

	testCases := []struct {
		name           string
		payload        map[string]interface{}
		wantStatusCode int
		wantToken      bool
	}{
		{
			name: "Successful Registration/Authentication",
			payload: map[string]interface{}{
				"username": "newuser",
				"password": "password123",
			},
			wantStatusCode: http.StatusOK,
			wantToken:      true,
		},
		{
			name: "No Username or Password",
			payload: map[string]interface{}{
				"username": "",
				"password": "",
			},
			wantStatusCode: http.StatusBadRequest,
			wantToken:      false,
		},
		{
			name: "Wrong Password",
			payload: map[string]interface{}{
				"username": "newuser",
				"password": "wrongpassword",
			},
			wantStatusCode: http.StatusUnauthorized,
			wantToken:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reqBody, err := json.Marshal(testCase.payload)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, serverAddr, bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := http.Client{Timeout: 5 * time.Second}
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, testCase.wantStatusCode, res.StatusCode)

			if testCase.wantToken {
				var response map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&response)
				require.NoError(t, err)
				_, ok := response["token"]
				require.True(t, ok, "Expected token to be present in response")
			} else {
				var response map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&response)
				require.NoError(t, err)
				_, ok := response["token"]
				require.False(t, ok, "Expected no token in response")
			}
		})
	}
}

func truncateTest(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE users RESTART IDENTITY CASCADE;
		TRUNCATE TABLE transactions RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		log.Fatalf("Failed to truncate test tables: %v", err)
	}
}
