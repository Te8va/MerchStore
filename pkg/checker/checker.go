package checker

import "net/http"

func IsJSONContentTypeCorrect(r *http.Request) bool {
	if len(r.Header.Values("Content-Type")) == 0 {
		return false
	}

	for _, contentType := range r.Header.Values("Content-Type") {
		if contentType == "application/json" {
			return true
		}
	}

	return false
}
