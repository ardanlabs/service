// +build go1.7

package httptreemux

import (
	"context"
	"net/http"
)

// ContextGroup is a wrapper around Group, with the purpose of mimicking its API, but with the use of http.HandlerFunc-based handlers.
// Instead of passing a parameter map via the handler (i.e. httptreemux.HandlerFunc), the path parameters are accessed via the request
// object's context.
type ContextGroup struct {
	group *Group
}

// Use appends a middleware handler to the Group middleware stack.
func (cg *ContextGroup) Use(fn MiddlewareFunc) {
	cg.group.Use(fn)
}

// UseHandler is like Use but accepts http.Handler middleware.
func (cg *ContextGroup) UseHandler(middleware func(http.Handler) http.Handler) {
	cg.group.UseHandler(middleware)
}

// UsingContext wraps the receiver to return a new instance of a ContextGroup.
// The returned ContextGroup is a sibling to its wrapped Group, within the parent TreeMux.
// The choice of using a *Group as the receiver, as opposed to a function parameter, allows chaining
// while method calls between a TreeMux, Group, and ContextGroup. For example:
//
//              tree := httptreemux.New()
//              group := tree.NewGroup("/api")
//
//              group.GET("/v1", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
//                  w.Write([]byte(`GET /api/v1`))
//              })
//
//              group.UsingContext().GET("/v2", func(w http.ResponseWriter, r *http.Request) {
//                  w.Write([]byte(`GET /api/v2`))
//              })
//
//              http.ListenAndServe(":8080", tree)
//
func (g *Group) UsingContext() *ContextGroup {
	return &ContextGroup{g}
}

// NewContextGroup adds a child context group to its path.
func (cg *ContextGroup) NewContextGroup(path string) *ContextGroup {
	return &ContextGroup{cg.group.NewGroup(path)}
}

func (cg *ContextGroup) NewGroup(path string) *ContextGroup {
	return cg.NewContextGroup(path)
}

// Handle allows handling HTTP requests via an http.HandlerFunc, as opposed to an httptreemux.HandlerFunc.
// Any parameters from the request URL are stored in a map[string]string in the request's context.
func (cg *ContextGroup) Handle(method, path string, handler http.HandlerFunc) {
	fullPath := cg.group.path + path
	cg.group.Handle(method, path, func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		routeData := &contextData{
			route:  fullPath,
			params: params,
		}
		r = r.WithContext(AddRouteDataToContext(r.Context(), routeData))
		handler(w, r)
	})
}

// Handler allows handling HTTP requests via an http.Handler interface, as opposed to an httptreemux.HandlerFunc.
// Any parameters from the request URL are stored in a map[string]string in the request's context.
func (cg *ContextGroup) Handler(method, path string, handler http.Handler) {
	fullPath := cg.group.path + path
	cg.group.Handle(method, path, func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		routeData := &contextData{
			route:  fullPath,
			params: params,
		}
		r = r.WithContext(AddRouteDataToContext(r.Context(), routeData))
		handler.ServeHTTP(w, r)
	})
}

// GET is convenience method for handling GET requests on a context group.
func (cg *ContextGroup) GET(path string, handler http.HandlerFunc) {
	cg.Handle("GET", path, handler)
}

// POST is convenience method for handling POST requests on a context group.
func (cg *ContextGroup) POST(path string, handler http.HandlerFunc) {
	cg.Handle("POST", path, handler)
}

// PUT is convenience method for handling PUT requests on a context group.
func (cg *ContextGroup) PUT(path string, handler http.HandlerFunc) {
	cg.Handle("PUT", path, handler)
}

// DELETE is convenience method for handling DELETE requests on a context group.
func (cg *ContextGroup) DELETE(path string, handler http.HandlerFunc) {
	cg.Handle("DELETE", path, handler)
}

// PATCH is convenience method for handling PATCH requests on a context group.
func (cg *ContextGroup) PATCH(path string, handler http.HandlerFunc) {
	cg.Handle("PATCH", path, handler)
}

// HEAD is convenience method for handling HEAD requests on a context group.
func (cg *ContextGroup) HEAD(path string, handler http.HandlerFunc) {
	cg.Handle("HEAD", path, handler)
}

// OPTIONS is convenience method for handling OPTIONS requests on a context group.
func (cg *ContextGroup) OPTIONS(path string, handler http.HandlerFunc) {
	cg.Handle("OPTIONS", path, handler)
}

type contextData struct {
	route  string
	params map[string]string
}

func (cd *contextData) Route() string {
	return cd.route
}

func (cd *contextData) Params() map[string]string {
	if cd.params != nil {
		return cd.params
	}
	return map[string]string{}
}

// ContextRouteData is the information associated with the matched path.
// Route() returns the matched route, without expanded wildcards.
// Params() returns a map of the route's wildcards and their matched values.
type ContextRouteData interface {
	Route() string
	Params() map[string]string
}

// ContextParams returns a map of the route's wildcards and their matched values.
func ContextParams(ctx context.Context) map[string]string {
	if cd := ContextData(ctx); cd != nil {
		return cd.Params()
	}
	return map[string]string{}
}

// ContextRoute returns the matched route, without expanded wildcards.
func ContextRoute(ctx context.Context) string {
	if cd := ContextData(ctx); cd != nil {
		return cd.Route()
	}
	return ""
}

// ContextData returns the ContextRouteData associated with the matched path
func ContextData(ctx context.Context) ContextRouteData {
	if p, ok := ctx.Value(contextDataKey).(ContextRouteData); ok {
		return p
	}
	return nil
}

// AddRouteDataToContext can be used for testing handlers, to insert route data into the request's `Context`.
func AddRouteDataToContext(ctx context.Context, data ContextRouteData) context.Context {
	return context.WithValue(ctx, contextDataKey, data)
}

// AddParamsToContext inserts a parameters map into a context using
// the package's internal context key.
func AddParamsToContext(ctx context.Context, params map[string]string) context.Context {
	return AddRouteDataToContext(ctx, &contextData{
		params: params,
	})
}

// AddRouteToContext inserts a route into a context using
// the package's internal context key.
func AddRouteToContext(ctx context.Context, route string) context.Context {
	return AddRouteDataToContext(ctx, &contextData{
		route: route,
	})
}

type contextKey int

// contextDataKey is used to retrieve the path's params map and matched route
// from a request's context.
const contextDataKey contextKey = 0
