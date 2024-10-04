package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreatePostRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostRequest
	if err := readJSON(w, r, &payload); err != nil {
		_ = writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		// TODO: Change after auth
		UserID: 1,
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		_ = writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		_ = writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	postIdParam := strings.TrimSpace(chi.URLParam(r, "postId"))
	if postIdParam == "" {
		_ = writeJSONError(w, http.StatusBadRequest, "missing postId URL parameter")
		return
	}

	postId, err := strconv.ParseInt(postIdParam, 10, 64)
	if err != nil {
		_ = writeJSONError(w, http.StatusBadRequest, "invalid postId URL parameter")
		return
	}

	ctx := r.Context()
	post, err := app.store.Posts.GetByID(ctx, postId)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			_ = writeJSONError(w, http.StatusNotFound, fmt.Sprintf("could not find post with ID %d", postId))
		default:
			log.Printf("could not get post: %+v", err)
			_ = writeJSONError(w, http.StatusInternalServerError, "an internal server error occurred.")
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		_ = writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
