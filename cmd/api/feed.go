package main

import (
	"net/http"

	"github.com/addvanced/gophersocial/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(100), pageable, filter)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
