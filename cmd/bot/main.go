// Package main Запуск приложения
package main

import (
	"PluginMattermost/internal/config"
	"PluginMattermost/internal/handler"
	"PluginMattermost/internal/logger"
	"PluginMattermost/internal/server"
	"PluginMattermost/internal/storage"
	"fmt"
)

func main() {
	cfg := config.MustGetConfig()

	log := logger.SetupLogger(cfg.Env)

	// Инициализируем хранилище (Tarantool)
	store := storage.MustNewTarantoolStorage(cfg.TarantoolHost, cfg.TarantoolPort, log)

	// Создаём PollHandler
	pollHandler := handler.NewPollHandler(store, log)

	// Инициализируем роутер и сервер
	router := server.NewRouter(pollHandler)
	server.MustStartServer(fmt.Sprintf(":%d", cfg.PortHTTP), router)
}
