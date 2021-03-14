package remux

import (
	"errors"
	"net/http"
	"strings"
	"sync"
)

type ReMux struct {
	mu              sync.RWMutex
	plain           map[string]map[string]http.Handler         // method-path-handler
	notFoundHandler http.Handler
}

func New() *ReMux {
	return &ReMux{
		notFoundHandler: http.HandlerFunc(defaultNotFoundHandler),
	}
}

var defaultNotFoundHandler = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
}

var (
	ErrInvalidMethod    = errors.New("invalid http method")
	ErrInvalidPath      = errors.New("invalid path")
	ErrNilHandler       = errors.New("handler is nil")
	ErrAmbiguousMapping = errors.New("ambiguous mapping")
)

func (r *ReMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.mu.RLock()

	var handler http.Handler
	if handlers, exists := r.plain[request.Method]; exists {
		if h, ok := handlers[request.URL.Path]; ok {
			handler = h
		}
	}
	if handler == nil {
		handler = r.notFoundHandler
	}

	r.mu.RUnlock()

	handler.ServeHTTP(writer, request)
}

func (r *ReMux) NotFound(handler http.Handler) error {
	if handler == nil {
		return ErrNilHandler
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.notFoundHandler = handler
	return nil
}

func (r *ReMux) RegisterPlain(method, path string, handler http.Handler) error {
	if !isValidMethod(method) {
		return ErrInvalidMethod
	}
	if !strings.HasPrefix(path, "/") {
		return ErrInvalidPath
	}
	if handler == nil {
		return ErrNilHandler
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plain[method][path]; exists {
		return ErrAmbiguousMapping
	}

	if r.plain == nil {
		r.plain = make(map[string]map[string]http.Handler)
	}
	if r.plain[method] == nil {
		r.plain[method] = make(map[string]http.Handler)
	}

	r.plain[method][path] = handler
	return nil
}

var validMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func isValidMethod(method string) bool {
	for _, m := range validMethods {
		if m == method {
			return true
		}
	}
	return false
}
