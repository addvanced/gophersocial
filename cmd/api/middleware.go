package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("missing credentials"))
				return
			}

			if username != app.config.auth.basic.username || password != app.config.auth.basic.password {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials: %s / %s", username, password))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
