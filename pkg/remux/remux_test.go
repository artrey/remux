package remux_test

import (
	"bytes"
	"github.com/artrey/remux/pkg/remux"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func stubHandler(w http.ResponseWriter, r *http.Request) {
	params, err := remux.PathParams(r.Context())
	if err == nil {
		_, _ = w.Write([]byte(params.Named["id"]))
	}
}

func TestReMux_RegisterPlain(t *testing.T) {
	muxWithHandlers := remux.New()
	err := muxWithHandlers.RegisterPlain(http.MethodGet, "/test", http.HandlerFunc(stubHandler))
	if err != nil {
		t.Errorf("mux problem: %v", err)
		return
	}

	type args struct {
		mux     *remux.ReMux
		method  string
		path    string
		handler http.Handler
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "correct register on empty mux",
			args: args{
				mux:     remux.New(),
				method:  http.MethodGet,
				path:    "/test",
				handler: http.HandlerFunc(stubHandler),
			},
			want: nil,
		},
		{
			name: "correct register on mux with handlers",
			args: args{
				mux:     muxWithHandlers,
				method:  http.MethodGet,
				path:    "/test2",
				handler: http.HandlerFunc(stubHandler),
			},
			want: nil,
		},
		{
			name: "invalid method",
			args: args{
				mux:    remux.New(),
				method: "foo",
			},
			want: remux.ErrInvalidMethod,
		},
		{
			name: "invalid path",
			args: args{
				mux:    remux.New(),
				method: http.MethodGet,
				path:   "test",
			},
			want: remux.ErrInvalidPath,
		},
		{
			name: "invalid handler",
			args: args{
				mux:     remux.New(),
				method:  http.MethodGet,
				path:    "/test",
				handler: nil,
			},
			want: remux.ErrNilHandler,
		},
		{
			name: "ambiguous mapping",
			args: args{
				mux:     muxWithHandlers,
				method:  http.MethodGet,
				path:    "/test",
				handler: http.HandlerFunc(stubHandler),
			},
			want: remux.ErrAmbiguousMapping,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(test.name, func(t *testing.T) {
			got := tt.args.mux.RegisterPlain(tt.args.method, tt.args.path, tt.args.handler)
			if tt.want != got {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReMux_RegisterRegex(t *testing.T) {
	muxWithHandlers := remux.New()

	regexps := make([]*regexp.Regexp, 0)
	for _, rePattern := range []string{
		`^/test/(?P<id>\d+)$`,
		`^/test/(?P<id>\d+)/subtest$`,
		`/test/(?P<id>\d+)$`,
		`^test/(?P<id>\d+)$`,
		`^/test/(?P<id>\d+)`,
		`test/(?P<id>\d+)`,
	} {
		re, err := regexp.Compile(rePattern)
		if err != nil {
			t.Errorf("regexp.Compile problem: %v", err)
			return
		}
		regexps = append(regexps, re)
	}
	err := muxWithHandlers.RegisterRegex(http.MethodGet, regexps[0], http.HandlerFunc(stubHandler))
	if err != nil {
		t.Errorf("mux problem: %v", err)
		return
	}

	type args struct {
		mux     *remux.ReMux
		method  string
		path    *regexp.Regexp
		handler http.Handler
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "correct register on empty mux",
			args: args{
				mux:     remux.New(),
				method:  http.MethodGet,
				path:    regexps[0],
				handler: http.HandlerFunc(stubHandler),
			},
			want: nil,
		},
		{
			name: "correct register on mux with handlers",
			args: args{
				mux:     muxWithHandlers,
				method:  http.MethodGet,
				path:    regexps[1],
				handler: http.HandlerFunc(stubHandler),
			},
			want: nil,
		},
		{
			name: "invalid method",
			args: args{
				mux:    remux.New(),
				method: "foo",
			},
			want: remux.ErrInvalidMethod,
		},
		{
			name: "invalid path ^",
			args: args{
				mux:    remux.New(),
				method: http.MethodGet,
				path:   regexps[2],
			},
			want: remux.ErrInvalidPath,
		},
		{
			name: "invalid path /",
			args: args{
				mux:    remux.New(),
				method: http.MethodGet,
				path:   regexps[3],
			},
			want: remux.ErrInvalidPath,
		},
		{
			name: "invalid path $",
			args: args{
				mux:    remux.New(),
				method: http.MethodGet,
				path:   regexps[4],
			},
			want: remux.ErrInvalidPath,
		},
		{
			name: "invalid path ^/$",
			args: args{
				mux:    remux.New(),
				method: http.MethodGet,
				path:   regexps[5],
			},
			want: remux.ErrInvalidPath,
		},
		{
			name: "invalid handler",
			args: args{
				mux:     remux.New(),
				method:  http.MethodGet,
				path:    regexps[0],
				handler: nil,
			},
			want: remux.ErrNilHandler,
		},
		{
			name: "ambiguous mapping",
			args: args{
				mux:     muxWithHandlers,
				method:  http.MethodGet,
				path:    regexps[0],
				handler: http.HandlerFunc(stubHandler),
			},
			want: remux.ErrAmbiguousMapping,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(test.name, func(t *testing.T) {
			got := tt.args.mux.RegisterRegex(tt.args.method, tt.args.path, tt.args.handler)
			if tt.want != got {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathParams(t *testing.T) {
	mux := remux.New()

	re, err := regexp.Compile(`^/test/(?P<id>\d+)$`)
	if err != nil {
		t.Errorf("regexp.Compile problem: %v", err)
		return
	}
	err = mux.RegisterRegex(http.MethodGet, re, http.HandlerFunc(stubHandler))
	if err != nil {
		t.Errorf("mux problem: %v", err)
		return
	}
	err = mux.RegisterPlain(http.MethodGet, "/test", http.HandlerFunc(stubHandler))
	if err != nil {
		t.Errorf("mux problem: %v", err)
		return
	}

	type args struct {
		mux    *remux.ReMux
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "named agr found",
			args: args{
				mux:    mux,
				method: http.MethodGet,
				path:   "/test/2",
			},
			want: []byte("2"),
		},
		{
			name: "named agr not found",
			args: args{
				mux:    mux,
				method: http.MethodGet,
				path:   "/test",
			},
			want: []byte{},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			response := httptest.NewRecorder()
			tt.args.mux.ServeHTTP(response, request)
			if response.Code != 200 {
				t.Errorf("error response code: %v", response.Code)
			}
			got := response.Body.Bytes()
			if !bytes.Equal(tt.want, got) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestReMux_ServeHTTP(t *testing.T) {
	mux := remux.New()
	err := mux.RegisterPlain(http.MethodGet, "/test", http.HandlerFunc(stubHandler))
	if err != nil {
		t.Errorf("mux problem: %v", err)
		return
	}

	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "try to get unregistered path",
			args: args{
				method: http.MethodGet,
				path:   "/test",
			},
			want: http.StatusOK,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			got := response.Code
			if tt.want != got {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReMux_NotFound(t *testing.T) {
	mux := remux.New()

	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "try to get unregistered path",
			args: args{
				method: http.MethodGet,
				path:   "/test",
			},
			want: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			got := response.Code
			if tt.want != got {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
