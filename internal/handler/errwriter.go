package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Te8va/MerchStore/internal/errors"
	"github.com/Te8va/MerchStore/pkg/logger"
)

func WriteHTTPError(w http.ResponseWriter, err error, statusCode int, prefix string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(errors.JSONError{Err: err.Error()}); err != nil {
		logger.Logger().Errorln(prefix, err.Error())
	}
}

func SendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		WriteHTTPError(w, err, http.StatusInternalServerError, "SendJSONResponse:")
	}
}
