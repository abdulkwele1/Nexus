package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nexus-api/service"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type APIConfig struct {
	APIPort string
}

var (
	apiConfig = APIConfig{
		APIPort: os.Getenv("API_PORT"),
	}
)

func main() {
	// #TODO make into a unit test
	//generate a hash for password123
	hash, err := HashPassword("password123")
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}
	fmt.Println("Hash for password123:", hash)

	fmt.Println("api server starting")

	router := mux.NewRouter()

	// setup handler functions to run whenever an api endpoint is called
	router.HandleFunc("/login", service.CorsMiddleware(LoginHandler))
	router.HandleFunc("/hello", service.CorsMiddleware(HelloServer))

	// attach router to default http server mux
	http.Handle("/", router)

	// run api service listening on the configured port
	http.ListenAndServe(fmt.Sprintf(":%s", apiConfig.APIPort), nil)

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
	Match       bool   `json:"match"`
	Cookie      string `json:"cookie"`
}

var LoginInfo = map[string]string{
	"abdul": "$2a$14$KXCe7VMOjZdf/BwSKIFLxu2FRHcr.DAQntjq8OfdqQI69EOQz4gHW",
	"levi":  "$2a$10$HqQx4jxUzfQm1fZYUZRLbOBaMNWHmhSmweH03rl0EykgE4BNfDciO",
}

var UserCookies = map[string]string{}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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

	fmt.Printf("login username %s, password %s\n", request.Username, request.Password)

	// username doesn't exist in our system
	passwordHashForUser, exists := LoginInfo[request.Username]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(struct{}{})
		return
	}

	match := CheckPasswordHash(request.Password, passwordHashForUser)

	response := LoginResponse{
		RedirectURL: "/",
		Match:       match,
	}

	//if passwordHash doesn't match
	if !match {
		w.Header().Set("Content-Type", "application/json")
		// return access denied
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(struct{}{})
		return
	}

	response.Cookie = uuid.NewString()

	UserCookies[request.Username] = response.Cookie

	fmt.Printf("password hash for user %s in our system is %s", request.Username, passwordHashForUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
