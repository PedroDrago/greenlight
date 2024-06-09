package main

import (
	"errors"
	"net/http"
	"time"

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
	token, err := app.models.Tokens.New(usr.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": token.PlainText,
			"userID":          usr.ID,
		}
		err = app.mailer.Send(usr.Email, "user_welcome.tmpl.html", data)
		if err != nil {
			app.logger.Error(err, nil)
		}
	})
	err = app.writeJSON(writer, http.StatusAccepted, envelope{"user": usr}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}

func (app *application) activateUserHandler(writer http.ResponseWriter, req *http.Request) {
	var input struct {
		TokenPlainText string `json:"token"`
	}
	err := app.readJSON(writer, req, &input)
	if err != nil {
		app.badRequestResponse(writer, req, err)
		return
	}

	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
		app.failedValidationResponse(writer, req, v.Errors)
		return
	}
	usr, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(writer, req, v.Errors)
		default:
			app.serverErrorResponse(writer, req, err)
		}
		return
	}
	usr.Activated = true
	err = app.models.Users.Update(usr)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(writer, req)
		default:
			app.serverErrorResponse(writer, req, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, usr.ID)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
		return
	}

	err = app.writeJSON(writer, http.StatusOK, envelope{"user": usr}, nil)
	if err != nil {
		app.serverErrorResponse(writer, req, err)
	}
}
