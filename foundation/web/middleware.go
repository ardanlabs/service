package web

// MidFunc is a handler function designed to run code before and/or after
// another Handler. It is designed to remove boilerplate or other concerns not
// direct to any given app Handler.
type MidFunc func(handler HandlerFunc) HandlerFunc

// wrapMiddleware creates a new handler by wrapping middleware around a final
// handler. The middlewares' Handlers will be executed by requests in the order
// they are provided.
func wrapMiddleware(mw []MidFunc, handler HandlerFunc) HandlerFunc {

	// Loop backwards through the middleware invoking each one. Replace the
	// handler with the new wrapped handler. Looping backwards ensures that the
	// first middleware of the slice is the first to be executed by requests.
	for i := len(mw) - 1; i >= 0; i-- {
		mwFunc := mw[i]
		if mwFunc != nil {
			handler = mwFunc(handler)
		}
	}

	return handler
}
