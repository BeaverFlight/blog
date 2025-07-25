package main

import (
	"blog/pkg/auth"
	"blog/pkg/dbwork"
	"blog/pkg/handlers"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func enableCORS(router *mux.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*") // Разрешить все домены
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}

func main() {
	router := mux.NewRouter()

	router.Use(handlers.LoggingMiddleware)
	enableCORS(router)

	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/register", handlers.Register).Methods("POST")
	router.HandleFunc("/article/{id}", handlers.GetArticle).Methods("GET")
	router.HandleFunc("/article", handlers.GetAllArticle).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(auth.AuthMiddleware())

	protected.HandleFunc("/article", handlers.CreateArticle).Methods("POST")
	protected.HandleFunc("/article/{id}", handlers.DeleteArticle).Methods("DELETE")
	protected.HandleFunc("/article", handlers.UpdateArticle).Methods("PUT")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func init() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}

	configDB := dbwork.PostgresDBParams{
		User:     os.Getenv("USER"),
		Password: os.Getenv("PASSWORD"),
		Host:     os.Getenv("HOST"),
		Port:     port,
		SSLMode:  os.Getenv("SSLMODE"),
		DBName:   os.Getenv("DBNAME"),
	}

	err = dbwork.InitializationDB(configDB)
	if err != nil {
		log.Fatal(err)
	}
	hours, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	if err != nil {
		log.Fatal(err)
	}

	auth.InitializationSecret(os.Getenv("JWT_SECRET"), hours)
}
