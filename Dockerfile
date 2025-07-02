# Используем минимальный образ Go для сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем git для загрузки зависимостей
RUN apk add --no-cache git

# Создаем рабочую директорию
WORKDIR /app

# Копируем go mod и sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Используем минимальный образ для runtime
FROM alpine:latest

# Устанавливаем сертификаты для HTTPS запросов и sqlite
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Копируем собранное приложение
COPY --from=builder /app/main .

# Создаем директорию для базы данных
RUN mkdir -p /root/data

# Открываем порт
EXPOSE 8080

# Устанавливаем переменную окружения для базы данных
ENV DATABASE_PATH=/root/data/brands.db

# Запускаем приложение
CMD ["./main"] 