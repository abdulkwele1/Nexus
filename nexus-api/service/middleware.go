package service

import (
	"context"
	"encoding/json"
	"net/http"
	"nexus-api/api"
	"nexus-api/clients/database"
	"time"
)

const (
	UsernameContextKey = "username"
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
