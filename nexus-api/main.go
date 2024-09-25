package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"nexus-api/clients/database"
	"nexus-api/clients/database/schemas/postgres/migrations"
	"nexus-api/logging"
	"nexus-api/service"

	"os"

	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var (
	serviceCtx    = context.Background()
	serviceLogger *logging.ServiceLogger
	apiService    *APIService
)

type APIConfig struct {
	APIPort string
}

type APIService struct {
	Config         APIConfig
	DatabaseClient *database.PostgresClient
	*logging.ServiceLogger
}

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
	// setup logger
	logLevel := os.Getenv("LOG_LEVEL")
	logger, err := logging.New(os.Getenv("LOG_LEVEL"))
	serviceLogger = &logger

	if err != nil {
		panic(fmt.Errorf("error %s creating serviceLogger with level %s", err, logLevel))
	}

	// parse database configuration from the environment
	databaseConfig := database.PostgresDatabaseConfig{
		DatabaseName:          os.Getenv("DATABASE_NAME"),
		DatabaseEndpointURL:   os.Getenv("DATABASE_ENDPOINT_URL"),
		DatabaseUsername:      os.Getenv("DATABASE_USERNAME"),
		DatabasePassword:      os.Getenv("DATABASE_PASSWORD"),
		SSLEnabled:            os.Getenv("DATABASE_SSL_ENABLED") == "true",
		QueryLoggingEnabled:   os.Getenv("DATABASE_QUERY_LOGGING_ENABLED") == "true",
		RunDatabaseMigrations: os.Getenv("RUN_DATABASE_MIGRATIONS") == "true",
		Logger:                serviceLogger,
	}
	serviceLogger.Debug().Msgf("loaded databaseClient confiEg %+v", databaseConfig)

	// create database client
	databaseClient, err := database.NewPostgresClient(databaseConfig)

	if err != nil {
		panic(fmt.Errorf("error %s creating database client with %+v", err, databaseConfig))
	}

	// run migrations based on configuration
	if databaseConfig.RunDatabaseMigrations {
		go func() {
			for {
				ranMigrations, err := database.Migrate(serviceCtx, databaseClient.DB, *migrations.Migrations, serviceLogger)

				if err != nil {
					serviceLogger.Error().Msgf("error %s running migrations %+v, will retry in 3 seconds", err, migrations.Migrations)

					time.Sleep(3 * time.Second)

					continue
				}

				serviceLogger.Info().Msgf("successfully ran migrations %+v", ranMigrations)

				// Example of how to update password for user
				// Step 1: check if user exists
				// check if username doesn't exist in our system
				userNameToUpdatePasswordFor := "levi"
				newPassword := "newPassword"
				loginAuthentication, err := database.GetLoginAuthenticationByUserName(serviceCtx, apiService.DatabaseClient.DB, userNameToUpdatePasswordFor)
				if err != nil {
					// handle case where user doesn't exist
					if errors.Is(err, database.ErrorNoLoginAuthenticationForUsername) {
						apiService.Debug().Msgf("%s for %s", err, userNameToUpdatePasswordFor)

						return
					}
					// handle case where issue talkng to database
					apiService.Error().Msgf("error %s looking up loginAuthentication for %s", err, userNameToUpdatePasswordFor)

					return
				}
				// Step 2: compute hash for new password
				newPasswordHash, err := HashPassword(newPassword)

				if err != nil {
					// handle error
					apiService.Error().Msgf("error %s hashing new password %s", err, newPassword)
				}
				// Step 3: update password hash for user in database
				loginAuthentication.PasswordHash = newPasswordHash
				err = loginAuthentication.Update(serviceCtx, apiService.DatabaseClient.DB)
				if err != nil {
					// handle error
					apiService.Error().Msgf("error %s updating loginAuthentication for %+v", err, loginAuthentication)
				}
				return
			}
		}()
	}

	// parse api config from the environment
	apiConfig := APIConfig{
		APIPort: os.Getenv("API_PORT"),
	}

	serviceLogger.Debug().Msgf("loaded api config %+v", apiConfig)

	// create api service
	apiService = &APIService{
		Config:         apiConfig,
		ServiceLogger:  serviceLogger,
		DatabaseClient: &databaseClient,
	}

	// #TODO make into a unit test
	//generate a hash for password123
	hash, err := HashPassword("password123")
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}
	serviceLogger.Debug().Msgf("Hash for password123: %s", hash)
	serviceLogger.Debug().Msg("api server starting")

	// setup api request router
	router := mux.NewRouter()

	// setup handler functions to run whenever a specific api endpoint is called
	router.HandleFunc("/login", service.CorsMiddleware(LoginHandler))
	router.HandleFunc("/hello", service.CorsMiddleware(AuthMiddleware(HelloServer)))         // Protect the hello route
	router.HandleFunc("/settings", service.CorsMiddleware(AuthMiddleware(SettingsHandler)))  // Protect the settings route
	router.HandleFunc("/home", service.CorsMiddleware(AuthMiddleware(HomeHandler)))          // Protect the home route
	router.HandleFunc("/solar", service.CorsMiddleware(AuthMiddleware(SolarHandler)))        //protects solar route
	router.HandleFunc("/loations", service.CorsMiddleware(AuthMiddleware(LocationsHandler))) //p protects location route

	http.Handle("/", router)

	// run api service listening on the configured port
	http.ListenAndServe(fmt.Sprintf(":%s", apiConfig.APIPort), nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	serviceLogger.Debug().Msgf("api called with %s \n", name)
	fmt.Fprintf(w, "Hello, %s!", name)
}

func LocationsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Settings page - only accessible with a valid cookie! User: %s", username)
}

func SolarHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Settings page - only accessible with a valid cookie! User: %s", username)
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
		serviceLogger.Debug().Msgf("error %s parsing %+v", err, request)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request"})
		return
	}

	serviceLogger.Debug().Msgf("login username %s, password %s\n", request.Username, request.Password)

	// check if username doesn't exist in our system
	loginAuthentication, err := database.GetLoginAuthenticationByUserName(serviceCtx, apiService.DatabaseClient.DB, request.Username)
	if err != nil {
		if errors.Is(err, database.ErrorNoLoginAuthenticationForUsername) {
			apiService.Debug().Msgf("%s for %s", err, request.Username)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct{}{})
			return
		}

		apiService.Error().Msgf("error %s looking up loginAuthentication for %s", err, request.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct{}{})
		return
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return
	}

	match := CheckPasswordHash(request.Password, loginAuthentication.PasswordHash)

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

	serviceLogger.Debug().Msgf("password hash for user %s in our system is %s", request.Username, loginAuthentication.PasswordHash)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
