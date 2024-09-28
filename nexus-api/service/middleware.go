package service

import (
	"context"
	"encoding/json"
	"net/http"
	"nexus-api/api"
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
func AuthMiddleware(next http.HandlerFunc, userCookies api.UserCookies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if the cookie value matches any user's cookie
		for username, userCookie := range userCookies {
			if userCookie == cookie.Value {
				// Attach username to request context for later use
				ctx := context.WithValue(r.Context(), "username", username)
				r = r.WithContext(ctx)
				next(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
	}
}
