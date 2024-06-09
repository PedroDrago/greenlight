package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				writer.Header().Set("Connection", "close")
				app.serverErrorResponse(writer, req, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(writer, req)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	limiter := *rate.NewLimiter(2, 4)
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if !limiter.Allow() {
			app.rateLimitExceededResponse(writer, req)
			return
		}
		next.ServeHTTP(writer, req)
	})
}
