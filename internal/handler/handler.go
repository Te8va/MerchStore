package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Te8va/MerchStore/internal/domain"
	appErrors "github.com/Te8va/MerchStore/internal/errors"
	"github.com/Te8va/MerchStore/internal/pkg"
	"github.com/Te8va/MerchStore/pkg/logger"
	"github.com/Te8va/MerchStore/pkg/validator"
)

type MerchHandler struct {
	srv    domain.MerchService
	JWTKey string
}

func NewMerchHandler(srv domain.MerchService, jwtKey string) *MerchHandler {
	return &MerchHandler{srv: srv, JWTKey: jwtKey}
}

func (h *MerchHandler) SendCoinHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fromUser, err := pkg.ExtractUsernameFromRequest(r, h.JWTKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}

	if err := validator.ValidateJSONRequest(r, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 || req.ToUser == "" || req.ToUser == fromUser {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = h.srv.SendCoin(r.Context(), fromUser, req.ToUser, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrUserNotFound):
			http.Error(w, "Receiver not found", http.StatusBadRequest)
		case errors.Is(err, appErrors.ErrInsufficientBalance):
			http.Error(w, "Insufficient balance", http.StatusBadRequest)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MerchHandler) GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	username, err := pkg.ExtractUsernameFromRequest(r, h.JWTKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	info, err := h.srv.GetUserInfo(r.Context(), username)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrUserNotFound):
			http.Error(w, "Invalid request", http.StatusBadRequest)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	filteredHistory := pkg.FilterCoinHistory(info.CoinHistory)

	var inventory []domain.InventoryItem
	if info.Inventory != nil {
		inventory = info.Inventory
	} else {
		inventory = []domain.InventoryItem{}
	}

	response := struct {
		Coins       int                    `json:"coins"`
		Inventory   []domain.InventoryItem `json:"inventory"`
		CoinHistory domain.CoinHistory     `json:"coinHistory"`
	}{
		Coins:       info.Coins,
		Inventory:   inventory,
		CoinHistory: filteredHistory,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger().Errorln("GetUserInfoHandler: failed to encode response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *MerchHandler) BuyMerchHandler(w http.ResponseWriter, r *http.Request) {
	username, err := pkg.ExtractUsernameFromRequest(r, h.JWTKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	item := r.URL.Path[len("/api/buy/"):]
	if item == "" {
		http.Error(w, "Invalid item", http.StatusBadRequest)
		return
	}

	err = h.srv.BuyMerch(r.Context(), username, item)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrInsufficientBalance),
			errors.Is(err, appErrors.ErrItemNotFound),
			errors.Is(err, appErrors.ErrUserNotFound):
			http.Error(w, "Bad Request", http.StatusBadRequest)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

type AuthorizationHandler struct {
	srv domain.AuthorizationService
}

func NewAuthorizationHandler(srv domain.AuthorizationService) *AuthorizationHandler {
	return &AuthorizationHandler{srv: srv}
}

func (h *AuthorizationHandler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	var authData domain.AuthorizationData

	if err := validator.ValidateJSONRequest(r, &authData); err != nil {
		WriteHTTPError(w, err, http.StatusBadRequest, "handlers.AuthHandler:")
		return
	}

	if authData.Username == "" || authData.Password == "" {
		WriteHTTPError(w, appErrors.ErrNoLoginOrPassword, http.StatusBadRequest, "handlers.AuthHandler:")
		return
	}

	token, err := h.srv.RegisterOrAuthenticate(r.Context(), authData.Username, authData.Password)
	if err != nil {
		if errors.Is(err, appErrors.ErrWrongPassword) {
			WriteHTTPError(w, appErrors.ErrWrongPassword, http.StatusUnauthorized, "handlers.AuthHandler:")
			return
		}

		WriteHTTPError(w, err, http.StatusInternalServerError, "handlers.AuthHandler:")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(domain.Token{Token: token}); err != nil {
		logger.Logger().Errorln("AuthHandler: failed to encode token:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
