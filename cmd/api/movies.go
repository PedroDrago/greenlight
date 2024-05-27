package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PedroDrago/greenlight/internal/data"
	"github.com/PedroDrago/greenlight/internal/validator"
)

func (app *application) createMovieHandler(writer http.ResponseWriter, req *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	err := app.readJSON(writer, req, &input)
	if err != nil {
		app.badRequestResponse(writer, req, err)
		return
	}
	v := validator.New()
	movie := data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	if data.ValidateMovie(v, &movie); !v.Valid() {
		app.failedValidationResponse(writer, req, v.Errors)
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
