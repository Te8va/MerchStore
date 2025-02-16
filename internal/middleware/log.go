package middleware

import (
	"net/http"
	"time"

	"github.com/Te8va/MerchStore/pkg/logger"
)

type informativeResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int64
}

func NewInformativeResponseWriter(w http.ResponseWriter) *informativeResponseWriter {
	return &informativeResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (irw *informativeResponseWriter) WriteHeader(statusCode int) {
	irw.statusCode = statusCode
	irw.ResponseWriter.WriteHeader(statusCode)
}

func (irw *informativeResponseWriter) Write(b []byte) (int, error) {
	count, err := irw.ResponseWriter.Write(b)
	irw.contentLength += int64(count)

	return count, err
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Logger().Info("Request HTTP method: ", r.Method, ", request route: ", r.URL.String(), ", length of content in request: ", r.ContentLength)

		irw := NewInformativeResponseWriter(w)

		start := time.Now()
		next.ServeHTTP(irw, r)
		duration := time.Since(start)

		logger.Logger().Info("Response status for ", r.Method, " ", r.URL.String(), ": ", irw.statusCode, ", length of content in response: ", irw.contentLength, ", processing duration: ", duration)
	})
}
