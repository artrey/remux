package remux_test

import (
	"github.com/artrey/remux/pkg/remux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func stubHandler(w http.ResponseWriter, r *http.Request) {}

func TestReMux_RegisterPlain(t *testing.T) {
	ambiguousMappingMux := remux.New()
	err := ambiguousMappingMux.RegisterPlain(http.MethodGet, "/test", http.HandlerFunc(stubHandler))
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
			name: "correct register",
			args: args{
				mux:     remux.New(),
				method:  http.MethodGet,
				path:    "/test",
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
				mux:     ambiguousMappingMux,
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
