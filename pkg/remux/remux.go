package remux

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type ReMux struct {
	mu              sync.RWMutex
	plain           map[string]map[string]http.Handler         // method-path-handler
	regex           map[string]map[*regexp.Regexp]http.Handler // method-regexPath-handler
	notFoundHandler http.Handler
}

type Middleware func(handler http.Handler) http.Handler

func New() *ReMux {
	return &ReMux{
		notFoundHandler: http.HandlerFunc(defaultNotFoundHandler),
	}
}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

var paramsContextKey = &contextKey{"remux context key for params"}

var defaultNotFoundHandler = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
}

var (
	ErrInvalidMethod    = errors.New("invalid http method")
	ErrInvalidPath      = errors.New("invalid path")
	ErrNilHandler       = errors.New("handler is nil")
	ErrAmbiguousMapping = errors.New("ambiguous mapping")
	ErrNoParams         = errors.New("no params")
)

type Params struct {
	Named      map[string]string
	Positional []string
}

func (r *ReMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.mu.RLock()

	var handler http.Handler
	if handlers, exists := r.plain[request.Method]; exists {
		if h, ok := handlers[request.URL.Path]; ok {
			handler = h
		}
	}
	if handler == nil {
		if handlers, exists := r.regex[request.Method]; exists {
			for path, h := range handlers {
				if matches := path.FindStringSubmatch(request.URL.Path); matches != nil {
					params := &Params{
						Named:      make(map[string]string),
						Positional: matches[1:],
					}
					for index, name := range path.SubexpNames() {
						if name == "" {
							continue
						}
						params.Named[name] = matches[index]
					}

					ctx := context.WithValue(request.Context(), paramsContextKey, params)
					request = request.WithContext(ctx)

					handler = h
					break
				}
			}
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

func (r *ReMux) RegisterPlain(method, path string, handler http.Handler, middlewares ...Middleware) error {
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

	r.plain[method][path] = wrapHandler(handler, middlewares...)
	return nil
}

func (r *ReMux) RegisterRegex(method string, path *regexp.Regexp, handler http.Handler, middlewares ...Middleware) error {
	if !isValidMethod(method) {
		return ErrInvalidMethod
	}
	if !strings.HasPrefix(path.String(), "^/") {
		return ErrInvalidPath
	}
	if !strings.HasSuffix(path.String(), "$") {
		return ErrInvalidPath
	}
	if handler == nil {
		return ErrNilHandler
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.regex[method][path]; exists {
		return ErrAmbiguousMapping
	}

	if r.regex == nil {
		r.regex = make(map[string]map[*regexp.Regexp]http.Handler)
	}
	if r.regex[method] == nil {
		r.regex[method] = make(map[*regexp.Regexp]http.Handler)
	}

	r.regex[method][path] = wrapHandler(handler, middlewares...)
	return nil
}

func PathParams(ctx context.Context) (*Params, error) {
	params, ok := ctx.Value(paramsContextKey).(*Params)
	if !ok {
		return nil, ErrNoParams
	}
	return params, nil
}

var validMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func wrapHandler(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		handler = m(handler)
	}
	return handler
}

func isValidMethod(method string) bool {
	for _, m := range validMethods {
		if m == method {
			return true
		}
	}
	return false
}
