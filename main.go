package main

import (
	"blog/pkg/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func enableCORS(router *mux.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
	enableCORS(router)

	err := handlers.StartDB()
	if err != nil {
		log.Println(err)
		return
	}

	router.HandleFunc("/article", handlers.CreateArticle).Methods("POST")
	router.HandleFunc("/article/{id}", handlers.GetArticle).Methods("GET")
	router.HandleFunc("/article/{id}", handlers.DeleteArticle).Methods("DELETE")
	router.HandleFunc("/article", handlers.UpdateArticle).Methods("PUT")
	router.HandleFunc("/register", handlers.Register).Methods("POST")
	router.HandleFunc("/article", handlers.GetAllArticle).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
