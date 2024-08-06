package main

import (
	"context"
	"fmt"
	"github.com/justinas/nosurf"
	"net/http"
)

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' fonts.bunny.net; font-src fonts.bunny.net;")
		writer.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-Frame-Options", "deny")
		writer.Header().Set("X-XSS-Protection", "0")

		writer.Header().Set("Server", "Go")

		next.ServeHTTP(writer, request)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var (
			ip     = request.RemoteAddr
			proto  = request.Proto
			method = request.Method
			uri    = request.URL.RequestURI()
		)

		app.logger.Info("received request:", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(writer, request)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				writer.Header().Set("Connection", "close")
				app.serverError(writer, request, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(writer, request)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !app.isAuthenticated(request) {
			http.Redirect(writer, request, "/user/login", http.StatusSeeOther)
			return
		}

		writer.Header().Set("Cache-Control", "no-store")

		next.ServeHTTP(writer, request)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		id := app.sessionManager.GetInt(request.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(writer, request)
			return
		}

		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(writer, request, err)
			return
		}

		if exists {
			ctx := context.WithValue(request.Context(), isAuthenticatedContextKey, true)
			request = request.WithContext(ctx)
		}

		next.ServeHTTP(writer, request)
	})
}
