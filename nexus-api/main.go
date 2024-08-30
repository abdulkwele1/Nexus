package main

import (
	"fmt"
	"net/http"
	"nexus-api/service"
)

func main() {
	// attach router to default http server mux
	http.Handle("/", service.CorsMiddleware(HelloServer))

	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	fmt.Printf("api called with %s \n", name)
	fmt.Fprintf(w, "Hello, %s!", name)
}
