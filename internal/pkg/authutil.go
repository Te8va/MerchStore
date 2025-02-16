package pkg

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Te8va/MerchStore/pkg/jwt"
)

func ExtractUsernameFromRequest(r *http.Request, jwtKey string) (string, error) {
	tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := jwt.ParseJWT(tokenStr, []byte(jwtKey))
	if err != nil {
		return "", err
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("unauthorized user")
	}

	return username, nil
}
