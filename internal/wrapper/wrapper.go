package wrapper

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// RouterWrapper is an interface for wrapper
type RouterWrapper interface {
	Add(child *wrapper) error
	FullPath() string

	Router() *mux.Router
	SetRouter(*mux.Router)

	Children() []RouterWrapper

	Handler() http.HandlerFunc
}

// wrapper wraps router with tree structure
type wrapper struct {
	router *mux.Router

	subPath string
	methods []string
	handler http.HandlerFunc

	children []RouterWrapper
	parent   RouterWrapper
}

// New is a constructor for the wrapper
func New(path string, methods []string, handler http.HandlerFunc) *wrapper {
	return &wrapper{
		subPath: path,
		methods: methods,
		handler: handler,
	}
}

// Router returns its router
func (w *wrapper) Router() *mux.Router {
	return w.router
}

// SetRouter sets its router
func (w *wrapper) SetRouter(r *mux.Router) {
	w.router = r
}

// Children returns its children
func (w *wrapper) Children() []RouterWrapper {
	return w.children
}

func (w *wrapper) Handler() http.HandlerFunc {
	return w.handler
}

// Add adds child as a child (child node of a tree) of w
func (w *wrapper) Add(child *wrapper) error {
	if child == nil {
		return fmt.Errorf("child is nil")
	}

	if child.parent != nil {
		return fmt.Errorf("child already has parent")
	}

	if child.subPath == "" || child.subPath == "/" || child.subPath[0] != '/' {
		return fmt.Errorf("child subpath is not valid")
	}

	child.parent = w
	w.children = append(w.children, child)

	child.router = w.router.PathPrefix(child.subPath).Subrouter()

	if child.handler != nil {
		if len(child.methods) > 0 {
			child.router.Methods(child.methods...).Subrouter().HandleFunc("/", child.handler)
			w.router.Methods(child.methods...).Subrouter().HandleFunc(child.subPath, child.handler)
		} else {
			child.router.HandleFunc("/", child.handler)
			w.router.HandleFunc(child.subPath, child.handler)
		}
	}

	return nil
}

// FullPath builds full path string of the api
func (w *wrapper) FullPath() string {
	if w.parent == nil {
		return w.subPath
	}
	re := regexp.MustCompile(`/{2,}`)
	return re.ReplaceAllString(w.parent.FullPath()+w.subPath, "/")
}
