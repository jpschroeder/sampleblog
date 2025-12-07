package util

import (
	"net/http"
	"strconv"
)

type IDHandler func(http.ResponseWriter, *http.Request, int64)

// Create a wrapper that parses the ID and passes it to the handler
func WithID(h IDHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")

		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		h(w, r, int64(id))
	}
}
