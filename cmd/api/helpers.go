package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PedroDrago/greenlight/internal/validator"
)

type envelope map[string]any

func (app *application) readJSON(writer http.ResponseWriter, req *http.Request, dst any) error {
	maxBytes := 1_048_576
	req.Body = http.MaxBytesReader(writer, req.Body, int64(maxBytes))
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contain badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type for field %d", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unkwnown key %s", fieldName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) writeJSON(writer http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t") // NOTE: marshalIndent is much slower, but improoves readability in the terminal. I guess this is not worth using in a real application, but as a study project it makes easier to test.
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		writer.Header()[key] = value
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	writer.Write(js)
	return nil
}

func (app *application) getIdParam(req *http.Request) (int64, error) {
	idStr := req.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return int64(id), nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(s)
	// TODO: negative numbers validation
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return n
}

func (app *application) background(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}
