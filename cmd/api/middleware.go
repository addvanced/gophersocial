package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/net/context"
)

func (app *application) AuthTokenMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedErrorResponse(w, r, errors.New("authorization header is missing"))
				return
			}

			bearer, token, found := strings.Cut(strings.TrimSpace(authHeader), " ")
			if !found || strings.ToLower(bearer) != "bearer" {
				app.unauthorizedErrorResponse(w, r, errors.New("authorization header is invalid"))
				return
			}

			jwtToken, err := app.authenticator.ValidateToken(token)
			if err != nil {
				app.unauthorizedErrorResponse(w, r, err)
				return
			} else if !jwtToken.Valid {
				app.unauthorizedErrorResponse(w, r, errors.New("token is invalid"))
				return
			}

			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				app.unauthorizedErrorResponse(w, r, errors.New("token does not contain valid claims"))
				return
			}

			userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
			if err != nil {
				app.unauthorizedErrorResponse(w, r, errors.New("token does not contain valid user ID"))
				return
			}

			ctx := r.Context()
			user, err := app.store.Users.GetByID(ctx, userID)
			if err != nil {
				app.unauthorizedErrorResponse(w, r, err)
				return
			}

			userCtx := context.WithValue(ctx, userCtxKey, &user)
			next.ServeHTTP(w, r.WithContext(userCtx))
		})
	}
}

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
