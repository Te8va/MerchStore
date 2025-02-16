package checker

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJSONContentTypeCorrect(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		expectedResult bool
	}{
		{
			name:           "Valid JSON content type",
			contentType:    "application/json",
			expectedResult: true,
		},
		{
			name:           "Invalid content type",
			contentType:    "text/plain",
			expectedResult: false,
		},
		{
			name:           "No Content-Type header",
			contentType:    "",
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &http.Request{
				Header: make(http.Header),
			}
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			result := IsJSONContentTypeCorrect(req)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
