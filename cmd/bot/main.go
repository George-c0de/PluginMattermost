package main

import (
	"PluginMattermost/internal/config"
	"PluginMattermost/internal/handler"
	"PluginMattermost/internal/server"
	"PluginMattermost/internal/storage"
	"fmt"
	"log"
)

func main() {
	cfg := config.MustGetConfig()

	// Инициализируем хранилище (Tarantool)
	store, err := storage.NewTarantoolStorage(cfg.TarantoolHost, cfg.TarantoolPort)
	if err != nil {
		log.Fatalf("Failed to connect to Tarantool: %v", err)
	}

	// Создаём PollHandler
	pollHandler := handler.NewPollHandler(store)

	// Инициализируем роутер и сервер
	router := server.NewRouter(pollHandler)
	if err := server.StartServer(fmt.Sprintf(":%d", cfg.PortHttp), router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
