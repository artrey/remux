package remux_test

import (
	"github.com/artrey/remux/pkg/remux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func stubHandler(w http.ResponseWriter, r *http.Request) {}

func TestReMux_RegisterPlain(t *testing.T) {
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
