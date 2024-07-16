package router

import (
	"net/http"
	"strings"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Router struct {
	routes map[string]map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{routes: make(map[string]map[string]HandlerFunc)}
}

func (r *Router) Handle(method, path string, handler HandlerFunc) {
	if _, exists := r.routes[path]; !exists {
		r.routes[path] = make(map[string]HandlerFunc)
	}
	r.routes[path][method] = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var path string
	index := strings.LastIndex(req.URL.Path, "/")
	if index != -1 {
		path = req.URL.Path[:index]
	}

	_, ok := r.routes[path]
	if !ok {
		path = req.URL.Path
	}

	if methodHandlers, exists := r.routes[path]; exists {
		if handler, exists := methodHandlers[req.Method]; exists {
			handler(w, req)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if strings.HasPrefix(req.URL.Path, "/static/") {
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))).ServeHTTP(w, req)
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}
