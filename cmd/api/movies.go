package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PedroDrago/greenlight/internal/data"
)

func (app *application) createMovieHandler(writer http.ResponseWriter, req *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}
	err := app.readJSON(writer, req, &input)
	if err != nil {
		app.errorResponse(writer, req, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Fprintf(writer, "%+v\n", input)
}

func (app *application) showMovieHandler(writer http.ResponseWriter, req *http.Request) {
	id, err := app.getIdParam(req)
	if err != nil {
		app.notFoundResponse(writer, req)
		return
	}
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "CasaBlanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	err = app.writeJSON(writer, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}
