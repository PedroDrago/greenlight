package main

import (
	"errors"
	"fmt"
	"net/http"

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
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(writer, req, v.Errors)
		return
	}
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))
	err = app.writeJSON(writer, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}

func (app *application) showMovieHandler(writer http.ResponseWriter, req *http.Request) {
	id, err := app.getIdParam(req)
	if err != nil {
		app.notFoundResponse(writer, req)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(writer, req)
		default:
			app.serverErrorResponse(writer, req, err)
		}
		return
	}
	err = app.writeJSON(writer, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}

func (app *application) updateMovieHandler(writer http.ResponseWriter, req *http.Request) {
	id, err := app.getIdParam(req)
	if err != nil {
		app.notFoundResponse(writer, req)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(writer, req)
		default:
			app.serverErrorResponse(writer, req, err)
		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	err = app.readJSON(writer, req, &input)
	if err != nil {
		app.badRequestResponse(writer, req, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(writer, req, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(writer, req)
		default:
			app.serverErrorResponse(writer, req, err)
		}
		return
	}

	err = app.writeJSON(writer, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}

func (app *application) deleteMovieHandler(writer http.ResponseWriter, req *http.Request) {
	id, err := app.getIdParam(req)
	if err != nil {
		app.notFoundResponse(writer, req)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(writer, req)
		default:
			app.serverErrorResponse(writer, req, err)
		}
		return
	}
	err = app.writeJSON(writer, http.StatusOK, envelope{"message": "Movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}
