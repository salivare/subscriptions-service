package router

import (
	"context"
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

type route struct {
	pattern string
	handler http.Handler
	parts   []string
}

type methodRoutes map[string][]route

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
	routes      methodRoutes
}

func New() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		routes: make(methodRoutes),
	}
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

func (r *Router) Handle(pattern string, h http.Handler) {
	r.mux.Handle(pattern, h)
}

func (r *Router) HandleFunc(pattern string, h http.HandlerFunc) {
	r.mux.HandleFunc(pattern, h)
}

func (r *Router) add(method, pattern string, h http.Handler) {
	if _, ok := r.routes[method]; !ok {
		r.routes[method] = []route{}
	}

	parts := strings.Split(pattern, "/")

	r.routes[method] = append(
		r.routes[method], route{
			pattern: pattern,
			handler: h,
			parts:   parts,
		},
	)
}

func (r *Router) GET(pattern string, h http.HandlerFunc) {
	r.add(http.MethodGet, pattern, h)
}

func (r *Router) POST(pattern string, h http.HandlerFunc) {
	r.add(http.MethodPost, pattern, h)
}

func (r *Router) PATCH(pattern string, h http.HandlerFunc) {
	r.add(http.MethodPatch, pattern, h)
}

func (r *Router) DELETE(pattern string, h http.HandlerFunc) {
	r.add(http.MethodDelete, pattern, h)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	methodRoutes := r.routes[req.Method]

	for _, rt := range methodRoutes {
		ok, params := matchRoute(rt.pattern, req.URL.Path)
		if ok {
			req = withPathParams(req, params)
			r.applyMiddleware(rt.handler).ServeHTTP(w, req)
			return
		}
	}

	r.applyMiddleware(r.mux).ServeHTTP(w, req)
}

func (r *Router) applyMiddleware(h http.Handler) http.Handler {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

func (r *Router) Mux() *http.ServeMux {
	return r.mux
}

func withPathParams(r *http.Request, params map[string]string) *http.Request {
	ctx := r.Context()
	for k, v := range params {
		ctx = context.WithValue(ctx, contextKey(k), v)
	}
	return r.WithContext(ctx)
}

type contextKey string

func PathValue(r *http.Request, key string) string {
	if v, ok := r.Context().Value(contextKey(key)).(string); ok {
		return v
	}
	return ""
}

func QueryValue(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func matchRoute(pattern string, path string) (bool, map[string]string) {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false, nil
	}

	params := map[string]string{}
	for i := range patternParts {
		p := patternParts[i]
		v := pathParts[i]

		if len(p) > 2 && p[0] == '{' && p[len(p)-1] == '}' {
			key := p[1 : len(p)-1]
			params[key] = v
			continue
		}

		if p != v {
			return false, nil
		}
	}

	return true, params
}
