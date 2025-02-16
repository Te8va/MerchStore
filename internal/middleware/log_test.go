package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testLogRouter(t *testing.T) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/logs", Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("/logs-with-write", Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("abc"))
		require.NoError(t, err)
	})))

	return mux
}

func request(t *testing.T, ts *httptest.Server, method, body, endpoint, authorization string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+endpoint, bytes.NewBuffer([]byte(body)))
	require.NoError(t, err)

	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	return resp
}

func TestLog(t *testing.T) {
	ts := httptest.NewServer(testLogRouter(t))
	defer ts.Close()

	var testTable = []struct {
		endpoint      string
		method        string
		content       string
		code          int
		body          string
		authorization string
	}{
		{
			"/logs",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
			"",
		},
		{
			"/logs-with-write",
			http.MethodGet,
			"",
			http.StatusOK,
			"abc",
			"",
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.method, testCase.body, testCase.endpoint, testCase.authorization)
		require.Equal(t, testCase.code, resp.StatusCode)
		resp.Body.Close()
	}
}
