package validator

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	appErrors "github.com/Te8va/MerchStore/internal/errors"
)

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func TestValidateJSONRequest(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		contentType    string
		expectedError  error
		expectedOutput TestStruct
	}{
		{
			name:          "Incorrect MIME type",
			body:          `{"field1": "value1", "field2": 123}`,
			contentType:   "text/plain",
			expectedError: appErrors.ErrWrongMIME,
		},
		{
			name:          "Invalid JSON format",
			body:          `{"field1": "value1", "field2": 123,}`,
			contentType:   "application/json",
			expectedError: appErrors.ErrWrongJSON,
		},
		{
			name:          "Unknown field in JSON",
			body:          `{"field1": "value1", "field2": 123, "field3": "unexpected"}`,
			contentType:   "application/json",
			expectedError: appErrors.ErrWrongJSON,
		},
		{
			name:          "Valid JSON input",
			body:          `{"field1": "value1", "field2": 123}`,
			contentType:   "application/json",
			expectedError: nil,
			expectedOutput: TestStruct{
				Field1: "value1",
				Field2: 123,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/test", strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", tc.contentType)

			var output TestStruct
			err = ValidateJSONRequest(req, &output)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}
