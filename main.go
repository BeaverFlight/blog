package main

import (
	"blog/pkg/auth"
	"blog/pkg/handlers"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func enableCORS(router *mux.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
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
	configFile, err := os.ReadFile("JWTSecret.json")
	var config struct {
		JWTSecret string `json:"jwt_secret"`
	}

	json.Unmarshal(configFile, &config)
	router := mux.NewRouter()
	enableCORS(router)

	err = handlers.StartDB()
	if err != nil {
		log.Println(err)
		return
	}

	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/register", handlers.Register).Methods("POST")
	router.HandleFunc("/article/{id}", handlers.GetArticle).Methods("GET")
	router.HandleFunc("/article", handlers.GetAllArticle).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(auth.AuthMiddleware(config.JWTSecret))

	protected.HandleFunc("/article", handlers.CreateArticle).Methods("POST")
	protected.HandleFunc("/article/{id}", handlers.DeleteArticle).Methods("DELETE")
	protected.HandleFunc("/article", handlers.UpdateArticle).Methods("PUT")
	log.Fatal(http.ListenAndServe(":8080", router))
}
