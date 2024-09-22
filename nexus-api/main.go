package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"nexus-api/service"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// Middleware to check for valid session cookie
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if the cookie value matches any user's cookie
		for username, userCookie := range UserCookies {
			if userCookie == cookie.Value {
				// Attach username to request context for later use
				ctx := context.WithValue(r.Context(), "username", username)
				r = r.WithContext(ctx)
				next(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
	}
}

func main() {
	hash, err := HashPassword("password123")
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}
	fmt.Println("Hash for password123:", hash)

	fmt.Println("API server starting")

	router := mux.NewRouter()

	router.HandleFunc("/login", service.CorsMiddleware(LoginHandler))
	router.HandleFunc("/hello", service.CorsMiddleware(AuthMiddleware(HelloServer)))        // Protect the hello route
	router.HandleFunc("/settings", service.CorsMiddleware(AuthMiddleware(SettingsHandler))) // Protect the settings route
	router.HandleFunc("/home", service.CorsMiddleware(AuthMiddleware(HomeHandler)))         // Protect the home route

	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	fmt.Printf("API called with %s \n", name)
	fmt.Fprintf(w, "Hello, %s!", name)
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Settings page - only accessible with a valid cookie! User: %s", username)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Home page - only accessible with a valid cookie! User: %s", username)
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
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request"})
		return
	}

	fmt.Printf("login username %s, password %s\n", request.Username, request.Password)

	passwordHashForUser, exists := LoginInfo[request.Username]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return
	}

	match := CheckPasswordHash(request.Password, passwordHashForUser)

	response := LoginResponse{
		RedirectURL: "/",
		Match:       match,
	}

	if !match {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return
	}

	response.Cookie = uuid.NewString()
	UserCookies[request.Username] = response.Cookie

	// Set the cookie with an expiration time
	expiration := time.Now().Add(1 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    response.Cookie,
		Path:     "/",
		Expires:  expiration,
		MaxAge:   3600,  // 1 hour
		HttpOnly: true,  // Optional: helps mitigate XSS
		Secure:   false, // Set to true if serving over HTTPS
	})

	fmt.Printf("password hash for user %s in our system is %s\n", request.Username, passwordHashForUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
