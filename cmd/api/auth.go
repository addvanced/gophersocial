package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8"`
}

// registerUserHandler godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserRequest	true	"User data"
//	@Success		201		{object}	User				"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload RegisterUserRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.StructCtx(ctx, payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.logger.Info("Registering user", "username", payload.Username, "email", payload.Email)

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	token := app.generateToken()

	if err := app.store.Users.CreateAndInvite(ctx, user, token, app.config.mail.inviteExpDuration); err != nil {
		app.logger.Error("Error inviting user", "error", err)
		switch err {
		case store.ErrDuplicateEmail:
			app.conflictResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.conflictResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.jsonResponse(w, http.StatusCreated, user)
}

func (app *application) generateToken() string {
	plainToken := uuid.New().String()

	// store
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	return hashToken
}
