package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/addvanced/gophersocial/internal/store"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidUserIDURLParam = errors.New("invalid user ID URL parameter")
	ErrUserAlreadyFollowed   = errors.New("user already followed")
	ErrUserAlreadyUnfollowed = errors.New("user already unfollowed")
	ErrFollowSameUser        = errors.New("cannot follow/unfollow yourself")
)

const userCtxKey ctxKey = "user"

// getUserHandler godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := app.GetInt64URLParam(ctx, "id")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByID(ctx, userID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, ErrUserNotFound)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

// followUserHandler godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userId}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := app.getAuthedUser(ctx)
	if authUser == nil {
		app.internalServerError(w, r, ErrUnauthorized)
		return
	}

	userID, err := app.GetInt64URLParam(ctx, "userId")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	} else if authUser.ID == userID {
		app.badRequestResponse(w, r, ErrFollowSameUser)
		return
	}

	if err := app.store.Follow.Follow(ctx, authUser.ID, userID); err != nil {
		switch err {
		case store.ErrAlreadyExists:
			app.badRequestResponse(w, r, ErrUserAlreadyFollowed)
		case store.ErrConflict:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// unfollowUserHandler gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userId}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := app.getAuthedUser(ctx)
	if authUser == nil {
		app.internalServerError(w, r, ErrUnauthorized)
		return
	}

	userID, err := app.GetInt64URLParam(ctx, "userId")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	} else if authUser.ID == userID {
		app.badRequestResponse(w, r, ErrFollowSameUser)
		return
	}

	if err := app.store.Follow.Unfollow(ctx, authUser.ID, userID); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, ErrUserAlreadyUnfollowed)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// activateUserHandler godoc
//
//	@Summary		Activates a new user profile
//	@Description	Activates a new user profile by the invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		202		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token, err := app.GetStringURLParam(ctx, "token")
	if err != nil {
		app.badRequestResponse(w, r, errors.New("missing activation token"))
		return
	}

	if err := app.store.Users.Activate(ctx, token); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, errors.New("activation token not found"))
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusAccepted, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getAuthedUser(ctx context.Context) *store.User {
	user, _ := ctx.Value(userCtxKey).(*store.User)
	return user
}
