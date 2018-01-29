package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Router is the router for all of the handlers
var Router *mux.Router

// Request is the basic type of all requests
type Request struct {
	Action string
	Params interface{}
}

func init() {
	Router = mux.NewRouter()
}

// Add adds the handler to the global handler list
func Add(resource string, handler func(http.ResponseWriter, *http.Request)) {
	Router.HandleFunc("/"+resource, handler)
}
