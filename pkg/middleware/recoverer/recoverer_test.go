package recoverer_test

import (
	"bytes"
	"github.com/artrey/remux/pkg/middleware/recoverer"
	"github.com/artrey/remux/pkg/remux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("panic example")
}

func TestRecoverer(t *testing.T) {
	mux := remux.New()

	err := mux.RegisterPlain(
		http.MethodGet,
		"/panic",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("panic example")
		}),
		recoverer.Recoverer,
	)
	if err != nil {
		t.Errorf("mux problem: %v", err)
		return
	}

	err = mux.RegisterPlain(
		http.MethodGet,
		"/nopanic",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		recoverer.Recoverer,
	)
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
		name     string
		args     args
		wantCode int
		wantBody []byte
	}{
		{
			name: "panic",
			args: args{
				mux:    mux,
				method: http.MethodGet,
				path:   "/panic",
			},
			wantCode: http.StatusInternalServerError,
			wantBody: []byte("panic example"),
		},
		{
			name: "no panic",
			args: args{
				mux:    mux,
				method: http.MethodGet,
				path:   "/nopanic",
			},
			wantCode: http.StatusOK,
			wantBody: []byte{},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			gotCode := response.Code
			if tt.wantCode != gotCode {
				t.Errorf("code error: got %v, want %v", gotCode, tt.wantCode)
			}
			gotBody := response.Body.Bytes()
			if !bytes.Equal(tt.wantBody, gotBody) {
				t.Errorf("body error: got %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}
