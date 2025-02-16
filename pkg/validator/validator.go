package validator

import (
	"encoding/json"
	"net/http"

	appErrors "github.com/Te8va/MerchStore/internal/errors"
	"github.com/Te8va/MerchStore/pkg/checker"
)

func ValidateJSONRequest(r *http.Request, v interface{}) error {
	if !checker.IsJSONContentTypeCorrect(r) {
		return appErrors.ErrWrongMIME
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	if err := d.Decode(&v); err != nil {
		return appErrors.ErrWrongJSON
	}

	if d.More() {
		return appErrors.ErrWrongJSON
	}

	return nil
}
