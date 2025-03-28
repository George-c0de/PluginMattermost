FROM golang:1.24 AS builder

WORKDIR /app

# Копируем go.mod и go.sum, чтобы кешировать зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка бинаря
WORKDIR /app/cmd/bot
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot

# Финальный образ
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bot /usr/local/bin/bot

# Запускаем бота
CMD ["bot"]