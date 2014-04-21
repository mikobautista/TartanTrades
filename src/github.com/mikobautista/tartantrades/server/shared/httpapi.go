package shared

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HTTP Server Running")
}

type HandlerType func(http.ResponseWriter, *http.Request)

func NewHttpServer(port int, pathHandlerMap map[string]HandlerType) {
	http.HandleFunc("/", handler) // redirect all urls to the handler function
	for k, v := range pathHandlerMap {
		http.HandleFunc(k, v)
	}
	http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil) // listen for connections at port 9999 on the local machine
}
