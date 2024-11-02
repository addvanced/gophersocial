package main

import (
	"net/http"

	"github.com/addvanced/gophersocial/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//	@Success		200		{object}	[]PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := app.getAuthedUser(ctx)
	if user == nil {
		app.internalServerError(w, r, ErrUnauthorized)
		return
	}

	pageable := store.Pageable{
		Limit:  10,
		Offset: 0,
		Sort:   "DESC",
	}.Parse(r)

	if err := Validate.StructCtx(ctx, pageable); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	filter, err := new(store.FeedFilter).Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	feed, err := app.store.Posts.GetUserFeed(ctx, user.ID, &pageable, filter)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
