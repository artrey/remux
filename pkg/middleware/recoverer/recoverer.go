package recoverer

import (
	"net/http"
)

func Recoverer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(r.(string)))
			}
		}()
		handler.ServeHTTP(w, r)
	})
}
