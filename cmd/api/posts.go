package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/addvanced/gophersocial/internal/store"
	"github.com/go-chi/chi/v5"
)

const postCtxKey ctxKey = "post"

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required,min=3,max=200"`
	Content string   `json:"content" validate:"required,min=3,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostRequest struct {
	Title   *string `json:"title" validate:"omitempty,min=3,max=200"`
	Content *string `json:"content" validate:"omitempty,min=3,max=1000"`
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

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostFromCtx(r)

	var payload UpdatePostRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := Validate.StructCtx(ctx, payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if err := app.store.Posts.Update(ctx, post); err != nil {
		switch err {
		case store.ErrDirtyRecord:
			app.conflictResponse(w, r, fmt.Errorf("post with ID '%d' has been modified by another user", post.ID))
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := app.getPostFromCtx(r)

	if err := app.store.Posts.Delete(r.Context(), post.ID); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, fmt.Errorf("post with ID '%d' does not exist", post.ID))
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) addPostToCtxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postId, err := strconv.ParseInt(strings.TrimSpace(chi.URLParam(r, "postId")), 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetByID(ctx, postId)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFoundResponse(w, r, fmt.Errorf("post with ID '%d' was not found", postId))
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, postCtxKey, post)))
	})
}

func (app *application) getPostFromCtx(r *http.Request) *store.Post {
	if post, ok := r.Context().Value(postCtxKey).(*store.Post); ok {
		return post
	}
	return nil
}
