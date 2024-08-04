package main

import (
	"errors"
	"fmt"
	"net/http"
	"snippetbox.i4o.dev/internal/models"
	"snippetbox.i4o.dev/internal/validator"
	"strconv"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *application) snippetCreatePost(writer http.ResponseWriter, request *http.Request) {
	var form snippetCreateForm

	err := app.decodePostForm(request, &form)
	if err != nil {
		app.clientError(writer, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "content", "This field must be equal to 1, 7, or 365")

	if !form.Valid() {
		app.logger.Error("Form values invalid")
		data := app.newTemplateData(request)
		data.Form = form
		app.render(writer, request, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(writer, request, err)
		return
	}

	app.sessionManager.Put(request.Context(), "flash", "Snippet created successfully!")

	http.Redirect(writer, request, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
	//writer.WriteHeader(http.StatusCreated)
	//writer.Write([]byte("Save a new snippet..."))
}

func (app *application) snippetCreate(writer http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)
	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(writer, request, http.StatusOK, "create.tmpl", data)
}

func (app *application) snippetView(writer http.ResponseWriter, request *http.Request) {
	// read id wildcard
	var id, err = strconv.Atoi(request.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(writer, request)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(writer, request)
		} else {
			app.serverError(writer, request, err)
		}

		return
	}

	data := app.newTemplateData(request)
	data.Snippet = snippet

	//execute the template - compile and render
	app.render(writer, request, http.StatusOK, "view.tmpl", data)

	//msg := fmt.Sprintf("Display a specific snippet with ID %d...", id)
	//writer.Write([]byte(msg))
	//fmt.Fprintf(writer, "%+v", snippet)
}

// handler function takes response writer and request as arguments
// same as in expressjs (req, res)
func (app *application) home(writer http.ResponseWriter, request *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(writer, request, err)
		return
	}

	data := app.newTemplateData(request)
	data.Snippets = snippets

	// execute the template - compile and render
	app.render(writer, request, http.StatusOK, "home.tmpl", data)
}
