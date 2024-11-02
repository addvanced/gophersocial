package store

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const timeFormat = "2006-01-02T15:04:05"

type FeedFilter struct {
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since" validate:"datetime=2006-01-02T15:04:05"`
	Until  string   `json:"until" validate:"datetime=2006-01-02T15:04:05"`
}

func (f FeedFilter) Parse(r *http.Request) (*FeedFilter, error) {
	q := r.URL.Query()

	if tags := strings.TrimSpace(q.Get("tags")); tags != "" {
		f.Tags = strings.Split(tags, ",")
	}

	if search := strings.TrimSpace(q.Get("search")); search != "" {
		f.Search = search
	}

	if since := strings.TrimSpace(q.Get("since")); since != "" {
		if _, err := time.Parse(timeFormat, since); err != nil {
			return &f, fmt.Errorf("since must be in format '%s'", timeFormat)
		}
		f.Since = since
	}

	if until := strings.TrimSpace(q.Get("until")); until != "" {
		if _, err := time.Parse(timeFormat, until); err != nil {
			return &f, fmt.Errorf("until must be in format '%s'", timeFormat)
		}
		f.Until = until
	}

	return &f, nil
}
