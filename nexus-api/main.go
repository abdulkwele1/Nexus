package main

import (
	"context"
	"encoding/json"
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
)

type APIConfig struct {
	APIPort string
}

type APIService struct {
	config         APIConfig
	server         http.Server
	databaseClient *database.PostgresClient
	*logging.ServiceLogger
}

func main() {
	// setup logger
	logLevel := os.Getenv("LOG_LEVEL")
	serviceLogger, err := logging.New(os.Getenv("LOG_LEVEL"))

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
		Logger:                &serviceLogger,
	}

	serviceLogger.Debug().Msgf("loaded databaseClient config %+v", databaseConfig)

	// create database client
	databaseClient, err := database.NewPostgresClient(databaseConfig)

	if err != nil {
		panic(fmt.Errorf("error %s creating database client with %+v", err, databaseConfig))
	}

	// run migrations based on configuration
	if databaseConfig.RunDatabaseMigrations {
		go func() {
			for {
				ranMigrations, err := database.Migrate(serviceCtx, databaseClient.DB, *migrations.Migrations, &serviceLogger)

				if err != nil {
					serviceLogger.Error().Msgf("error %s running migrations %+v, will retry in 3 seconds", err, migrations.Migrations)

					time.Sleep(3 * time.Second)

					continue
				}

				serviceLogger.Info().Msgf("successfully ran migrations %+v", ranMigrations)

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
	_ = APIService{
		config:         apiConfig,
		ServiceLogger:  &serviceLogger,
		databaseClient: &databaseClient,
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
	router.HandleFunc("/hello", service.CorsMiddleware(HelloServer))

	// attach router to default http server mux
	http.Handle("/", router)

	// run api service listening on the configured port
	http.ListenAndServe(fmt.Sprintf(":%s", apiConfig.APIPort), nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	serviceLogger.Debug().Msgf("api called with %s \n", name)
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
		serviceLogger.Debug().Msgf("error %s parsing %+v", err, request)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct{}{})
		return
	}

	serviceLogger.Debug().Msgf("login username %s, password %s\n", request.Username, request.Password)

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

	serviceLogger.Debug().Msgf("password hash for user %s in our system is %s", request.Username, passwordHashForUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
