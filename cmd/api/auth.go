package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/addvanced/gophersocial/internal/mailer"
	"github.com/addvanced/gophersocial/internal/store"
	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email,max=320"`
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
//	@Router			/auth/user [post]
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

	user := &store.User{
		Username: payload.Username,
		Email:    strings.ToLower(strings.TrimSpace(payload.Email)),
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// plain token to be used for email...
	plainToken, hashedToken := app.generateToken()
	if err := app.store.Users.CreateAndInvite(ctx, user, hashedToken, app.config.mail.inviteExpDuration); err != nil {
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

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken),
	}

	err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, (app.config.env != "production"))
	if err != nil {
		app.logger.Errorw("could not send welcome email", "error", err, "user", user)

		// rollback user creation if email fails
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("could not rollback user creation", "error", err, "user", user)
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) generateToken() (plainToken string, hashToken string) {
	// Plain Token to send to user
	plainToken = uuid.New().String()

	// Hashed Token to store in DB
	hash := sha256.Sum256([]byte(plainToken))
	hashToken = hex.EncodeToString(hash[:])
	return
}
