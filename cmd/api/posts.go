package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required,min=3,max=200"`
	Content string   `json:"content" validate:"required,min=3,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := Validate.StructCtx(ctx, payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		// TODO: Change after auth
		UserID: 1,
	}

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	postIdParam := strings.TrimSpace(chi.URLParam(r, "postId"))
	if postIdParam == "" {
		app.badRequestResponse(w, r, errors.New("missing postId URL parameter"))
		return
	}

	postId, err := strconv.ParseInt(postIdParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("invalid postId URL parameter: %s", postIdParam))
		return
	}

	ctx := r.Context()
	post, err := app.store.Posts.GetByID(ctx, postId)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, fmt.Errorf("post with ID '%d' was not found", postId))
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}
