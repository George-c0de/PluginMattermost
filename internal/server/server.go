package server

import (
	"PluginMattermost/internal/handler"
	"PluginMattermost/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

// NewRouter настраивает роутер и регистрирует эндпоинты.
func NewRouter(pollHandler *handler.PollHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/create", pollHandler.CreatePoll).Methods("POST")
	router.HandleFunc("/closed", pollHandler.EndPoll).Methods("POST")
	router.HandleFunc("/results", pollHandler.Results).Methods("GET")
	router.HandleFunc("/health", pollHandler.HealthHandler).Methods("GET")

	router.Handle("/vote", middleware.UserIDMiddleware(http.HandlerFunc(pollHandler.Vote))).Methods("POST")
	router.Handle("/vote", middleware.UserIDMiddleware(http.HandlerFunc(pollHandler.DeleteVote))).Methods("DELETE")

	return router
}

// MustStartServer запускает HTTP сервер на указанном адресе.
func MustStartServer(addr string, router *mux.Router) {
	err := http.ListenAndServe(addr, router)
	if err != nil {
		panic(err)
	}
}
