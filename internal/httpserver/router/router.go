package router

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type methodRoutes map[string]http.Handler
type routeTable map[string]methodRoutes

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
	routes      routeTable
}

func New() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		routes: make(routeTable),
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
		r.routes[method] = make(methodRoutes)
	}
	r.routes[method][pattern] = h
}

func (r *Router) GET(pattern string, h http.HandlerFunc) {
	r.add(http.MethodGet, pattern, h)
}

func (r *Router) POST(pattern string, h http.HandlerFunc) {
	r.add(http.MethodPost, pattern, h)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if methodRoutes, ok := r.routes[req.Method]; ok {
		if h, ok := methodRoutes[req.URL.Path]; ok {
			r.applyMiddleware(h).ServeHTTP(w, req)
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
