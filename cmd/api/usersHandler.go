package main

import (
	"errors"
	"net/http"

	"github.com/PedroDrago/greenlight/internal/data"
	"github.com/PedroDrago/greenlight/internal/validator"
)

func (app *application) createUserHandler(writer http.ResponseWriter, req *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(writer, req, &input)
	if err != nil {
		app.badRequestResponse(writer, req, err)
		return
	}
	usr := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	err = usr.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
		return
	}
	v := validator.New()

	if usr.Validate(v); !v.Valid() {
		app.failedValidationResponse(writer, req, v.Errors)
		return
	}
	err = app.models.Users.Insert(usr)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(writer, req, v.Errors)
		default:
			app.serverErrorResponse(writer, req, err)
		}
		return
	}
	err = app.mailer.Send(usr.Email, "user_welcome.tmpl.html", usr)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
		return
	}
	err = app.writeJSON(writer, http.StatusCreated, envelope{"user": usr}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}
