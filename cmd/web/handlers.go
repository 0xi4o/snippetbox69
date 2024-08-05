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

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
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

func (app *application) userSignup(writer http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)
	data.Form = userSignupForm{}

	app.render(writer, request, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(writer http.ResponseWriter, request *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(request, &form)
	if err != nil {
		app.clientError(writer, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")

	form.CheckField(validator.MinChars(form.Password, 8), "password", "Password should be at least 8 characters long")

	if !form.Valid() {
		app.logger.Error("Form values invalid")
		data := app.newTemplateData(request)
		data.Form = form
		app.render(writer, request, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(request)
			data.Form = form
			app.render(writer, request, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(writer, request, err)
		}

		return
	}

	app.sessionManager.Put(request.Context(), "flash", "Sign up successful. Please log in.")

	http.Redirect(writer, request, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(writer http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)
	data.Form = userLoginForm{}

	app.render(writer, request, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(writer http.ResponseWriter, request *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(request, &form)
	if err != nil {
		app.clientError(writer, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		app.logger.Error("Form values invalid")
		data := app.newTemplateData(request)
		data.Form = form
		app.render(writer, request, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(request)
			data.Form = form
			app.render(writer, request, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(writer, request, err)
			return
		}

		return
	}

	err = app.sessionManager.RenewToken(request.Context())
	if err != nil {
		app.serverError(writer, request, err)
		return
	}

	app.sessionManager.Put(request.Context(), "authenticatedUserID", id)

	http.Redirect(writer, request, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(writer http.ResponseWriter, request *http.Request) {
	err := app.sessionManager.RenewToken(request.Context())
	if err != nil {
		app.serverError(writer, request, err)
		return
	}

	app.sessionManager.Remove(request.Context(), "authenticatedUserID")

	app.sessionManager.Put(request.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(writer, request, "/user/login", http.StatusSeeOther)
}
