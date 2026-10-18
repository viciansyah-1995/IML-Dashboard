package app

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Page   int
	Size   int
	Offset int
}

func pagination(r *http.Request, total int) Pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 || page > 1000000 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if size < 1 || size > 100 {
		size = 20
	}
	pages := max(1, (total+size-1)/size)
	page = min(page, pages)
	return Pagination{Page: page, Size: size, Offset: (page - 1) * size}
}
