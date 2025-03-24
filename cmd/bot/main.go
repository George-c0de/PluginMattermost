package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"PluginMattermost/internal/handler"
	"PluginMattermost/internal/storage"
)

func main() {
	// Считываем переменные окружения (или используем значения по умолчанию)
	host := os.Getenv("TARANTOOL_HOST")
	if host == "" {
		host = "localhost"
	}
	portStr := os.Getenv("TARANTOOL_PORT")
	if portStr == "" {
		portStr = "3301"
	}
	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid TARANTOOL_PORT: %v", err)
	}

	// Инициализируем хранилище (Tarantool)
	store, err := storage.NewTarantoolStorage(host, portNum)
	if err != nil {
		log.Fatalf("Failed to connect to Tarantool: %v", err)
	}
	log.Println("Connected to Tarantool")

	// Создаём PollHandler
	pollHandler := handler.NewPollHandler(store)

	// Настраиваем роутер
	r := mux.NewRouter()
	r.HandleFunc("/create", pollHandler.CreatePoll).Methods("POST")
	r.HandleFunc("/vote", pollHandler.Vote).Methods("POST")
	r.HandleFunc("/results", pollHandler.Results).Methods("GET")

	// (Опционально) эндпоинт для проверки "живости" сервиса
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	}).Methods("GET")

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
