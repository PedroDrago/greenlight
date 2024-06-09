package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(writer http.ResponseWriter, req *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}
	err := app.writeJSON(writer, http.StatusOK, env, nil)
	if err != nil {
		app.logger.Error(err, nil)
		app.serverErrorResponse(writer, req, err)
	}
}
