package main

import (
	"net/http"

	"github.com/amock/rpcapi/cmd/handler"

	_ "github.com/amock/rpcapi/cmd/handler/cluster"
)

func main() {
	http.Handle("/", handler.Router)
	http.ListenAndServe(":8080", nil)
}
