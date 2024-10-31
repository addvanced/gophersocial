package store

import (
	"net/http"
	"strconv"
	"strings"
)

type Pageable struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=20"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"oneof=asc desc ASC DESC"`
}

func (p Pageable) Parse(r *http.Request) Pageable {
	q := r.URL.Query()

	if lq := strings.TrimSpace(q.Get("limit")); lq != "" {
		if limit, err := strconv.Atoi(lq); err == nil {
			p.Limit = limit
		}
	}

	if oq := strings.TrimSpace(q.Get("offset")); oq != "" {
		if offset, err := strconv.Atoi(oq); err == nil {
			p.Offset = offset
		}
	}

	if sort := strings.TrimSpace(strings.ToUpper(q.Get("sort"))); sort != "" {
		p.Sort = sort
	}

	return p
}
