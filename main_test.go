package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPreviewTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "works preview context",
			ctx:  context.WithValue(context.Background(), addedPreviewKey{}, "100"),
			want: "100",
		},
		{
			name: "works non-exist preview",
			ctx:  context.WithValue(context.Background(), "unknown", "100"),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(dummyPreviewTransportHandler))
			defer ts.Close()

			req, err := http.NewRequest("GET", ts.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			req = req.WithContext(tt.ctx)

			cli := NewClient(tt.ctx)
			resp, err := cli.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, string(b)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func dummyPreviewTransportHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.Header.Get(PreviewHeader)))
}

func TestPreviewMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		header  map[string]string
		wantCtx string
	}{
		{
			name: "contain preview header",
			header: map[string]string{
				"origin":      "foo.example.com",
				PreviewHeader: "10",
			},
			wantCtx: "10",
		},
		{
			name: "works non-exist preview header",
			header: map[string]string{
				"origin": "foo.example.com",
			},
			wantCtx: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			for k, v := range tt.header {
				req.Header.Add(k, v)
			}

			handler := PreviewMiddleware(http.HandlerFunc(dummyHandler))
			handler.ServeHTTP(rec, req)
			if diff := cmp.Diff(tt.wantCtx, rec.Body.String()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := r.Context().Value(addedPreviewKey{}).(string)
	if !ok {
		v = ""
	}
	w.Write([]byte(v))
}
