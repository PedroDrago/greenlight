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
	if movie.Validate(v); !v.Valid() {
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

	if movie.Validate(v); !v.Valid() {
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

type params struct {
	title    string
	genres   []string
	page     int32
	pageSize int32
	sort     string
}

func (app *application) listMoviesHandler(writer http.ResponseWriter, req *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()
	qs := req.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Sort = app.readString(qs, "sort", "id")
	if !v.Valid() {
		app.failedValidationResponse(writer, req, v.Errors)
		return
	}
	input.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if input.Validate(v); !v.Valid() {
		app.failedValidationResponse(writer, req, v.Errors)
		return
	}
	movies, metadata, err := app.models.Movies.List(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
		return
	}
	err = app.writeJSON(writer, http.StatusOK, envelope{"metadata": metadata, "movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}
