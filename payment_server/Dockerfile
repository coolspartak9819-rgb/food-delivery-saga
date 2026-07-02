# Используем multi-stage build для маленького и безопасного итогового образа
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum из корня проекта
COPY ../go.mod ../go.sum ./

# Копируем папку vendor с зависимостями
COPY ../vendor ./vendor

# Копируем proto файлы
COPY ../proto ./proto

# Копируем код этого конкретного сервиса
COPY . .

# Собираем приложение с флагом -mod=vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/main .

# Финальный этап - создаем минимальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из builder'а
COPY --from=builder /app/main .

# Указываем команду для запуска
CMD ["./main"]
