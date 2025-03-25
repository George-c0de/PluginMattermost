package server

import (
	"PluginMattermost/internal/middleware"
	"log"
	"net/http"

	"PluginMattermost/internal/handler"
	"github.com/gorilla/mux"
)

// NewRouter настраивает роутер и регистрирует эндпоинты.
func NewRouter(pollHandler *handler.PollHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/create", pollHandler.CreatePoll).Methods("POST")
	router.HandleFunc("/closed", pollHandler.EndPoll).Methods("POST")
	router.HandleFunc("/results", pollHandler.Results).Methods("GET")
	router.HandleFunc("/health", healthHandler).Methods("GET")

	router.Handle("/vote", middleware.UserIDMiddleware(http.HandlerFunc(pollHandler.Vote))).Methods("POST")
	router.Handle("/vote", middleware.UserIDMiddleware(http.HandlerFunc(pollHandler.DeleteVote))).Methods("DELETE")

	return router
}

// healthHandler обеспечивает базовую проверку "живости" сервиса.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("health check write error: %v", err)
	}
}

// StartServer запускает HTTP сервер на указанном адресе.
func StartServer(addr string, router *mux.Router) error {
	log.Printf("Server is running on %s", addr)
	return http.ListenAndServe(addr, router)
}
