# syntax=docker/dockerfile:1

########################
# 1) Build — stage
########################
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник
# Опция CGO_ENABLED=0 делает бинарник статически слинкованным — удобно для минимального образа
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot_app .

########################
# 2) Run — минимальный образ
########################
FROM alpine:latest

WORKDIR /app

# Если бот нуждается в CA-сертификатах (например, для TLS/HTTPS запросов),
# лучше их добавить (на всякий). Если не нужны — можно пропустить.
RUN apk add --no-cache ca-certificates

# Копируем скомпилированный бинарник
COPY --from=builder /bot_app ./bot_app

# Опционально: копируем папку с данными/файлами, если бот что-то сохраняет на диск
# COPY data/ ./data/

# Устанавливаем точку входа
ENTRYPOINT ["./bot_app"]
