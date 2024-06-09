package main

import (
	"fmt"
	"net/http"
)

func (app *application) failedValidationResponse(writer http.ResponseWriter, req *http.Request, errors map[string]string) {
	app.errorResponse(writer, req, http.StatusUnprocessableEntity, errors)
}

func (app *application) logError(req *http.Request, err error) {
	app.logger.Error(err, map[string]string{
		"request_method": req.Method,
		"request_url":    req.URL.String(),
	})
}

func (app *application) errorResponse(writer http.ResponseWriter, req *http.Request, status int, message any) {
	env := envelope{"error": message}
	err := app.writeJSON(writer, status, env, nil)
	if err != nil {
		app.logError(req, err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(writer http.ResponseWriter, req *http.Request, err error) {
	message := "Internal Server Error"
	app.logError(req, err)
	app.errorResponse(writer, req, http.StatusInternalServerError, message)
}

func (app *application) badRequestResponse(writer http.ResponseWriter, req *http.Request, err error) {
	app.errorResponse(writer, req, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(writer http.ResponseWriter, req *http.Request) {
	message := "Resource could not be found"
	app.errorResponse(writer, req, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(writer http.ResponseWriter, req *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", req.Method)
	app.errorResponse(writer, req, http.StatusMethodNotAllowed, message)
}

func (app *application) editConflictResponse(writer http.ResponseWriter, req *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(writer, req, http.StatusConflict, message)
}

func (app *application) rateLimitExceededResponse(writer http.ResponseWriter, req *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(writer, req, http.StatusTooManyRequests, message)
}
