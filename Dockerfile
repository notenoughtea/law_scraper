FROM golang:1.24-alpine AS builder

WORKDIR /build

# Установка зависимостей
RUN apk add --no-cache git ca-certificates

# Копируем go.mod и go.sum
COPY scraper/go.mod scraper/go.sum ./
RUN go mod download

# Копируем весь код
COPY scraper/ ./

# Собираем бинарники
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/cron cmd/cron/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/scraper cmd/scraper/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Установка CA сертификатов и timezone
RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарники из builder
COPY --from=builder /bin/cron ./bin/cron
COPY --from=builder /bin/scraper ./bin/scraper

# Создаем директорию для данных
RUN mkdir -p /app/data/matched

# Устанавливаем временную зону (опционально)
ENV TZ=Europe/Moscow

CMD ["./bin/cron"]

