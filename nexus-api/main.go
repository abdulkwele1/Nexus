package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nexus-api/service"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Print("api server starting")

	router := mux.NewRouter()

	// attach router to default http server mux
	router.HandleFunc("/login", service.CorsMiddleware(LoginHandler))
	router.HandleFunc("/hello", service.CorsMiddleware(HelloServer))

	// attach router to default http server mux
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	fmt.Printf("api called with %s \n", name)
	fmt.Fprintf(w, "Hello, %s!", name)
}

type LoginRequest struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginResponse struct {
	RedirectURL string `json:"redirect_url"`
	Password    string `json:"password"`
	Username    string `json:"username"`
}

var LoginInfo = map[string]string{
	"username1": "passwordHash",
	"username2": "passwordHash",
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		fmt.Printf("error %s parsing %+v", err, request)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct{}{})
		return
	}

	fmt.Printf("login username %s, password %s", request.Username, request.Password)

	response := LoginResponse{
		RedirectURL: "/",
		Password:    request.Password,
		Username:    request.Username,
	}

	passwordHashForUser, exists := LoginInfo[request.Username]

	// username doesn't exist in our system
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		// return access denied
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(struct{}{})
		return
	}

	fmt.Printf("password hash for user %s in our system is %s", request.Username, passwordHashForUser)

	// hash the password provided by the user in the request
	// compare the hash of the password we generated with passwordHashForUser
	// if they are a match, return below
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	// otherwise if not return  error
	// w.Header().Set("Content-Type", "application/json")
	// 	// return access denied
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(struct{}{})
	// 	return
}
