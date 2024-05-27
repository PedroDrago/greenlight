package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(_ *http.Request, err error) {
	app.errorLog.Println(err)
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

func (app *application) notFoundResponse(writer http.ResponseWriter, req *http.Request) {
	message := "Resource could not be found"
	app.errorResponse(writer, req, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(writer http.ResponseWriter, req *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", req.Method)
	app.errorResponse(writer, req, http.StatusMethodNotAllowed, message)
}
