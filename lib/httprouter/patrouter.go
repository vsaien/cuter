package httprouter

import (
	"context"
	"net/http"
	"path"
	"strings"
)

const (
	allowHeader          = "Allow"
	allowMethodSeparator = ", "
	pathVars             = "pathVars"
)

type PatRouter struct {
	trees    map[string]*SearchTree
	notFound http.Handler
}

func NewPatRouter() Router {
	return &PatRouter{
		trees: make(map[string]*SearchTree),
	}
}

func (pr *PatRouter) Handle(method, reqPath string, handler http.Handler) error {
	if !validMethod(method) {
		return ErrInvalidMethod
	}

	if len(reqPath) == 0 || reqPath[0] != '/' {
		return ErrInvalidPath
	}

	cleanPath := path.Clean(reqPath)
	if tree, ok := pr.trees[method]; ok {
		return tree.Add(cleanPath, handler)
	} else {
		tree = NewSearchTree()
		pr.trees[method] = tree
		return tree.Add(cleanPath, handler)
	}
}

func (pr *PatRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := path.Clean(r.URL.Path)
	if tree, ok := pr.trees[r.Method]; ok {
		if result, ok := tree.Search(reqPath); ok {
			if len(result.Params) > 0 {
				r = r.WithContext(context.WithValue(r.Context(), pathVars, result.Params))
			}
			result.Item.(http.Handler).ServeHTTP(w, r)
			return
		}
	}

	if allow, ok := pr.methodNotAllowed(r.Method, reqPath); ok {
		w.Header().Set(allowHeader, allow)
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		pr.handleNotFound(w, r)
	}
}

func (pr *PatRouter) SetNotFoundHandler(handler http.Handler) {
	pr.notFound = handler
}

func (pr *PatRouter) handleNotFound(w http.ResponseWriter, r *http.Request) {
	if pr.notFound != nil {
		pr.notFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (pr *PatRouter) methodNotAllowed(method, path string) (string, bool) {
	var allows []string

	for treeMethod, tree := range pr.trees {
		if treeMethod == method {
			continue
		}

		_, ok := tree.Search(path)
		if ok {
			allows = append(allows, treeMethod)
		}
	}

	if len(allows) > 0 {
		return strings.Join(allows, allowMethodSeparator), true
	} else {
		return "", false
	}
}

func Vars(r *http.Request) map[string]string {
	vars, ok := r.Context().Value(pathVars).(map[string]string)
	if ok {
		return vars
	}

	return nil
}

func validMethod(method string) bool {
	return method == http.MethodDelete || method == http.MethodGet ||
		method == http.MethodHead || method == http.MethodOptions ||
		method == http.MethodPatch || method == http.MethodPost ||
		method == http.MethodPut
}
