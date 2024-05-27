package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type envelope map[string]any

func (app *application) readJSON(_ http.ResponseWriter, req *http.Request, dst any) error {
	err := json.NewDecoder(req.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
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
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			app.errorLog.Println("foda")
			return err
		}
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
