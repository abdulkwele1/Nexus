package service

import (
	"context"
	"encoding/json"
	"net/http"
	"nexus-api/api"
	"nexus-api/clients/database"
	"time"
)

type contextKey string

const (
	UsernameContextKey contextKey = "user"
)

func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization,  Access-Control-Allow-Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Middleware to check for valid session cookie
func AuthMiddleware(next http.HandlerFunc, apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie == nil {
			apiService.Trace().Msgf("no cookie found for request %+v", r)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if the cookie value matches any user's cookie
		loginCookie, err := database.GetLoginCookie(r.Context(), apiService.DatabaseClient.DB, cookie.Value)

		if err != nil {
			apiService.Trace().Msgf("no matching cookie %s for request %+v", cookie.Value, r)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// handle case if cookie is expired
		if loginCookie.Expiration.Before(time.Now()) {
			apiService.Trace().Msgf("expired cookie %s for request %+v", cookie.Value, r)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Cookie expired"})
			return
		}

		// Attach username to request context for later use
		ctx := context.WithValue(r.Context(), UsernameContextKey, loginCookie.UserName)
		r = r.WithContext(ctx)
		// call next handler
		next(w, r)
	}
}

// AdminMiddleware checks if the authenticated user has admin privileges
func AdminMiddleware(next http.HandlerFunc, apiService *APIService) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {

		// Get username from context (set by AuthMiddleware)
		username, ok := r.Context().Value(UsernameContextKey).(string)
		if !ok {
			apiService.Error().Msgf("Username not found in context for admin request")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Get user info to check role
		user, err := database.GetLoginAuthenticationByUserName(r.Context(), apiService.DatabaseClient.DB, username)
		if err != nil {
			apiService.Error().Msgf("Error getting user info for admin check: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if user has admin privileges
		if user.Role != "admin" && user.Role != "root_admin" {
			apiService.Error().Msgf("ACCESS DENIED: user %s with role '%s' attempted to access admin endpoint", user.UserName, user.Role)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Insufficient privileges"})
			return
		}

		// User is authenticated and has admin privileges, call next handler
		next(w, r)
	}, apiService)
}
