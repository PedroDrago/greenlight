package main

import (
	"fmt"
	"net/http"
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
