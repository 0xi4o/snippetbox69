package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"net/http"
	"runtime/debug"
	"time"
)

func (app *application) clientError(writer http.ResponseWriter, status int) {
	http.Error(writer, http.StatusText(status), status)
}

func (app *application) render(writer http.ResponseWriter, request *http.Request, status int, page string, data templateData) {
	ts, ok := app.templates[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(writer, request, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(writer, request, err)
	}

	writer.WriteHeader(status)

	buf.WriteTo(writer)
}

func (app *application) serverError(writer http.ResponseWriter, request *http.Request, err error) {
	var (
		method = request.Method
		uri    = request.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)

	http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) newTemplateData(request *http.Request) templateData {
	return templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(request.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(request),
		CSRFToken:       nosurf.Token(request),
	}
}

func (app *application) decodePostForm(request *http.Request, dst any) error {
	err := request.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, request.PostForm)

	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *application) isAuthenticated(request *http.Request) bool {
	isAuthenticated, ok := request.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
