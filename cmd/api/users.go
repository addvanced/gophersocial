package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-chi/chi/v5"
)

const userCtxKey ctxKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

type FollowUserRequest struct {
	UserID int64 `json:"user_id" validate:"required,min=1"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	userToFollow := app.getUserFromCtx(r)

	// Revert back to auth userID from ctx
	var payload FollowUserRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := Validate.StructCtx(ctx, payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Follow.Follow(ctx, userToFollow.ID, payload.UserID); err != nil {
		switch err {
		case store.ErrAlreadyExists:
			app.badRequestResponse(w, r, errors.New("you are already following this user"))
		case store.ErrConflict:
			app.badRequestResponse(w, r, errors.New("you cannot follow yourself"))
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	userToUnfollow := app.getUserFromCtx(r)

	// Revert back to auth userID from ctx
	var payload FollowUserRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := Validate.StructCtx(ctx, payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Follow.Unfollow(ctx, userToUnfollow.ID, payload.UserID); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, errors.New("you are not following this user"))
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) addUserToCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := strconv.ParseInt(strings.TrimSpace(chi.URLParam(r, "userId")), 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetByID(ctx, userId)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFoundResponse(w, r, fmt.Errorf("user with ID '%d' was not found", userId))
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, userCtxKey, user)))
	})
}

func (app *application) getUserFromCtx(r *http.Request) *store.User {
	if user, ok := r.Context().Value(userCtxKey).(*store.User); ok {
		return user
	}
	return nil
}