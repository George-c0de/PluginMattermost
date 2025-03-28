# Система опросов

Данный проект представляет собой веб-сервис для проведения опросов на языке Go с использованием базы данных Tarantool.

# Оглавление

- [Описание проекта](#описание)
- [Структура репозитория](#структура-проекта)
- [Требования](#использование-api)
- [Запуск проекта](#установка)
- [Требования](#требования)

## Описание

Проект позволяет:

- Создавать и управлять опросами.
- Обрабатывать HTTP-запросы с использованием Go.
- Вести логирование работы сервера и ошибок.
- Интегрироваться с базой данных Tarantool для хранения результатов опросов.

## Структура проекта

* main.go — точка входа в приложение.
* config.go — настройки и инициализация конфигурации.
* server.go — реализация HTTP-сервера.
* middleware.go — промежуточное ПО для обработки запросов.
* poll.go — бизнес-логика и модели опросов.
* poll_handler.go — обработка HTTP-запросов, связанных с опросами.
* logger.go — модуль логирования.
* tarantool.go — интеграция с базой данных Tarantool.
* config.lua — дополнительные конфигурационные настройки.

## Использование API

После запуска сервера API доступен по адресу, указанному в конфигурационных файлах (например, http://localhost:8000).

Примеры запросов:

* GET /polls — получение списка опросов.
* POST /polls — создание нового опроса.
* GET /polls/{id} — получение данных конкретного опроса.

Более подробная документация по API может быть добавлена в будущем.

## Установка

1. Для быстрого запуска рекомендуется использовать Docker-compose
   ```bash
   docker-compose up --build -d
   ```
   После чего будет проект будет доступен по адресу из конфигурации или же по умолчанию(localhost:8000) API,
   Tarantool(localhost:3301)

Чтобы задать другие переменные конфига, необходимо прописать ENV параметры при запуске

```bash
# Параметры отображения логов 
ENV=local # по умолчанию
ENV=dev # dev окружение
ENV=prod # prod окружение

# Порт для REST API запросов
PORT_HTTP=8000 # по умолчанию

# Параметры Tarantool, в docker-compose файле переопределяются
TARANTOOL_HOST=tarantool
TARANTOOL_PORT=3301
```

Пример:

   ```bash
   ENV=production docker compose up --build
   ```

### Если поднят локально Tarantool

1. **Клонирование репозитория:**
   ```bash
   git clone git@github.com:George-c0de/PluginMattermost.git
   ```

2. Установка зависимостей:
   Если проект использует Go-модули, выполните:
   ```bash
   go mod download
   ```

3. Сборка проекта:

   ```bash
   go build -o pollapp cmd/bot/main.go
   ```

4. Запуск приложения:
   Запустите собранный бинарный файл:
   ```bash
   ./poolapp
   ```

   Или выполните для быстрого запуска:
   ```bash
   go run cmd/bot/main.go
   ```

## Требования

- [Go](https://golang.org/dl/) (версия 1.16 или выше)
- [Tarantool](https://www.tarantool.io/)