package wrapper

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// RouterWrapper wraps router with tree structure
type RouterWrapper struct {
	Router *mux.Router

	SubPath string
	Methods []string
	Handler http.HandlerFunc

	Children []*RouterWrapper
	Parent   *RouterWrapper
}

// New is a constructor for the RouterWrapper
func New(path string, methods []string, handler http.HandlerFunc) *RouterWrapper {
	return &RouterWrapper{
		SubPath: path,
		Methods: methods,
		Handler: handler,
	}
}

// Add adds child as a child (child node of a tree) of w
func (w *RouterWrapper) Add(child *RouterWrapper) error {
	if child == nil {
		return fmt.Errorf("child is nil")
	}

	if child.Parent != nil {
		return fmt.Errorf("child already has parent")
	}

	if child.SubPath == "" || child.SubPath == "/" || child.SubPath[0] != '/' {
		return fmt.Errorf("child subpath is not valid")
	}

	child.Parent = w
	w.Children = append(w.Children, child)

	child.Router = w.Router.PathPrefix(child.SubPath).Subrouter()

	if child.Handler != nil {
		if len(child.Methods) > 0 {
			child.Router.Methods(child.Methods...).Subrouter().HandleFunc("/", child.Handler)
			w.Router.Methods(child.Methods...).Subrouter().HandleFunc(child.SubPath, child.Handler)
		} else {
			child.Router.HandleFunc("/", child.Handler)
			w.Router.HandleFunc(child.SubPath, child.Handler)
		}
	}

	return nil
}

// FullPath builds full path string of the api
func (w *RouterWrapper) FullPath() string {
	if w.Parent == nil {
		return w.SubPath
	}
	re := regexp.MustCompile(`/{2,}`)
	return re.ReplaceAllString(w.Parent.FullPath()+w.SubPath, "/")
}
