package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/addvanced/gophersocial/internal/store"
)

const postCtxKey ctxKey = "post"

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required,min=3,max=200"`
	Content string   `json:"content" validate:"required,min=3,max=1000"`
	Tags    []string `json:"tags"`
} // @name CreatePostRequest

type UpdatePostRequest struct {
	Title   *string `json:"title" validate:"omitempty,min=3,max=200"`
	Content *string `json:"content" validate:"omitempty,min=3,max=1000"`
} // @name CreatePostRequest

// createPostHandler godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostRequest	true	"Post request payload"
//	@Success		201		{object}	Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := app.getAuthorizedUser(ctx)
	if authUser == nil {
		app.internalServerError(w, r, ErrUnauthorized)
		return
	}

	var payload CreatePostRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.StructCtx(ctx, payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  authUser.ID,
	}

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

// getPostHandler godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postId	path		int	true	"Post ID"
//	@Success		200		{object}	Post
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	post := app.getPostFromCtx(ctx)
	if post == nil {
		app.internalServerError(w, r, errors.New("could not find post"))
		return
	}

	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

// updatePostHandler godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postId	path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostRequest	true	"Post request payload"
//	@Success		200		{object}	Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	post := app.getPostFromCtx(ctx)
	if post == nil {
		app.internalServerError(w, r, errors.New("could not find post"))
		return
	}

	var payload UpdatePostRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

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

// deletePostHandler godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postId	path		int	true	"Post ID"
//	@Success		204		{object}	string
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	post := app.getPostFromCtx(ctx)
	if post == nil {
		app.internalServerError(w, r, errors.New("could not find post"))
		return
	}

	if err := app.store.Posts.Delete(ctx, post.ID); err != nil {
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
		ctx := r.Context()

		postId, err := app.GetInt64URLParam(ctx, "postId")
		if err != nil {
			app.badRequestResponse(w, r, errors.New("missing post ID"))
			return
		}

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

		postCtx := context.WithValue(ctx, postCtxKey, &post)
		next.ServeHTTP(w, r.WithContext(postCtx))
	})
}

func (app *application) getPostFromCtx(ctx context.Context) *store.Post {
	post, _ := ctx.Value(postCtxKey).(*store.Post)
	return post
}
