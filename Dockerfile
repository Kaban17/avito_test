FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0
WORKDIR /api

COPY . .
RUN go mod download

RUN go build -o api ./cmd/api

FROM alpine:latest
WORKDIR /api

# Установка необходимых пакетов
RUN apk add --no-cache postgresql-client bash

# Копирование бинарного файла и скриптов
COPY --from=builder /api/api /api/api
COPY scripts/init-db-docker.sh /api/scripts/init-db-docker.sh
COPY migrations/ /api/migrations/

# Сделать скрипт исполняемым
RUN chmod +x /api/scripts/init-db-docker.sh

# Команда запуска: сначала инициализация БД, затем запуск приложения
CMD ["/bin/bash", "-c", "/api/scripts/init-db-docker.sh && /api/api"]
