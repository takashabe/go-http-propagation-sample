package main

import (
	"context"
	"net/http"
)

const PreviewHeader = "X-PREVIEW"

type addedPreviewKey struct{}

func PreviewMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get(PreviewHeader)
			if h == "" {
				next.ServeHTTP(w, r)
				return
			}
			r = r.WithContext(context.WithValue(
				r.Context(),
				addedPreviewKey{},
				h,
			))
			next.ServeHTTP(w, r)
		})
}

type PreviewTransport struct {
	Base http.RoundTripper
}

func (t *PreviewTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func (t *PreviewTransport) RoundTrip(req *http.Request) (
	*http.Response, error) {

	tr := t.base()
	r := req.Clone(req.Context())
	v, ok := r.Context().Value(addedPreviewKey{}).(string)
	if ok {
		r.Header.Add(PreviewHeader, v)
	}
	return tr.RoundTrip(r)
}

func (t *PreviewTransport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.base().(canceler); ok {
		cr.CancelRequest(req)
	}
}

func NewClient(ctx context.Context) *http.Client {
	cl := http.DefaultClient
	tr := &PreviewTransport{
		Base: cl.Transport,
	}
	cl.Transport = tr
	return cl
}
