# TP Go Service

Go API для работы с TravelPayouts и WeGoTrip с форматированием для ManyChat.

## Быстрый старт

```bash
# Docker
docker-compose up --build

# Локально  
go run main.go
```

Сервис доступен на `http://localhost:8080`

## API

### Health Check
```bash
curl http://localhost:8080/health
```

### Получить экскурсии
```bash
curl -X POST http://localhost:8080/api/getFeed \
  -H "Content-Type: application/json" \
  -d '{"city": "москва", "lang": "ru", "currency": "RUB"}'
```

### Создать аффилиатную ссылку
```bash
curl -X POST http://localhost:8080/api/getFromLink \
  -H "Content-Type: application/json" \
  -d '{
    "link": "https://booking.com/hotel/example",
    "token": "YOUR_TOKEN",
    "trs": "123456",
    "marker": "654321"
  }'
```

## Модули

- `TravelPayouts/` - создание аффилиатных ссылок
- `WeGoTrip/` - получение фида экскурсий 
- `ManyChat/` - форматирование ответов в формате для ManyChat

## Технологии

Go 1.21, Gin, Docker
