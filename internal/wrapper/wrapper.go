package wrapper

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

type HandleFunc func(http.ResponseWriter, *http.Request)

type RouterWrapper struct {
	Router *mux.Router

	SubPath string
	Methods []string
	Handler HandleFunc

	Children []*RouterWrapper
	Parent   *RouterWrapper
}

func New(path string, methods []string, handler HandleFunc) *RouterWrapper {
	return &RouterWrapper{
		SubPath: path,
		Methods: methods,
		Handler: handler,
	}
}

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

func (w *RouterWrapper) FullPath() string {
	if w.Parent == nil {
		return w.SubPath
	} else {
		re := regexp.MustCompile(`/{2,}`)
		return re.ReplaceAllString(w.Parent.FullPath()+w.SubPath, "/")
	}
}
