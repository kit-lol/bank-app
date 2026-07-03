# Используем официальный образ Go для сборки
FROM golang:1.25.0-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main .

# Финальный образ (меньше размером)
FROM alpine:latest
WORKDIR /root/

# Копируем собранный бинарник из builder
COPY --from=builder /app/main .

# Копируем шаблоны, статику, конфиги и миграции
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/config/config.yaml ./internal/config/config.yaml
COPY --from=builder /app/migrations ./migrations

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./main"]