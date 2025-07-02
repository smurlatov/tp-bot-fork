# Travelpayouts Go Service

Легковесный Go сервис для создания аффилиатных ссылок через Travelpayouts API.

## Особенности

- 🚀 Легковесный и быстрый Go сервис  
- 🔗 Два эндпоинта: создание ссылок из URL и по имени бренда
- 📊 Полное логирование всех операций
- 🗄️ SQLite база данных для хранения брендов
- 🐳 Docker поддержка для легкого развертывания
- 🚀 Готов для развертывания на Railway

## Архитектура

```
POST /api/getFromLink    - Создает аффилиатную ссылку из прямой ссылки
POST /api/getFromBrand   - Создает аффилиатную ссылку по имени бренда  
GET  /health             - Проверка состояния сервиса
```

## Быстрый старт

### Docker (рекомендуется)

```bash
# Клонируем репозиторий
git clone <your-repo>
cd tp-go-service

# Собираем и запускаем
docker-compose up --build

# Сервис будет доступен на http://localhost:8080
```

### Локальная разработка

```bash
# Устанавливаем зависимости
go mod tidy

# Запускаем сервис
go run .

# Сервис будет доступен на http://localhost:8080
```

## Использование

### Создание ссылки из URL

```bash
curl -X POST http://localhost:8080/api/getFromLink \
  -H "Content-Type: application/json" \
  -d '{
    "link": "https://www.booking.com/hotel/us/plaza.html",
    "token": "your_travelpayouts_token",
    "trs": "your_trs_parameter",
    "marker": "your_marker"
  }'
```

### Создание ссылки по имени бренда

```bash
curl -X POST http://localhost:8080/api/getFromBrand \
  -H "Content-Type: application/json" \
  -d '{
    "brand_name": "booking",
    "token": "your_travelpayouts_token", 
    "trs": "your_trs_parameter",
    "marker": "your_marker"
  }'
```

## Развертывание на Railway

1. Создайте аккаунт на [Railway](https://railway.app)
2. Подключите ваш GitHub репозиторий
3. Railway автоматически обнаружит Dockerfile и соберет образ
4. Сервис будет развернут и доступен по предоставленному URL

### Переменные окружения для Railway

Установите следующие переменные окружения в Railway:

```
PORT=8080
GIN_MODE=release
```

## Поддерживаемые бренды

По умолчанию сервис поддерживает:

- `booking` - Booking.com
- `agoda` - Agoda
- `aviasales` - Aviasales  
- `hotels` - Hotels.com
- `expedia` - Expedia

Список можно расширить, добавив новые записи в базу данных.

## API Документация

Подробная документация API доступна в файле [API.md](./API.md).

## Структура проекта

```
tp-go-service/
├── main.go              # Основной сервер и роутинг
├── travelpayouts.go     # Логика работы с Travelpayouts API
├── go.mod               # Go модуль
├── Dockerfile           # Docker образ
├── docker-compose.yml   # Docker Compose конфигурация
├── .dockerignore        # Исключения для Docker
├── README.md            # Документация
└── API.md               # API документация
```

## Логирование

Сервис использует структурированное JSON логирование через logrus:

- Все HTTP запросы логируются с методом, путем и IP
- Операции с базой данных логируются  
- Запросы к Travelpayouts API логируются
- Ошибки логируются с полным контекстом

## Технологии

- **Go 1.21** - основной язык
- **Gin** - HTTP веб-фреймворк
- **GORM** - ORM для работы с базой данных
- **SQLite** - легковесная база данных
- **Logrus** - структурированное логирование
- **Docker** - контейнеризация

## Безопасность

- Все входящие данные валидируются
- Используется structured logging без вывода секретных данных
- Graceful shutdown для корректного завершения работы
- Health check эндпоинт для мониторинга

## Мониторинг

Сервис предоставляет:

- Health check на `/health`
- Структурированные логи в JSON формате
- HTTP метрики через Gin middleware

## Разработка

### Добавление нового бренда

1. Добавьте запись в базу данных через SQL или код:

```go
brand := Brand{
    BrandName: "newbrand", 
    BrandLink: "https://newbrand.com"
}
db.Create(&brand)
```

### Интеграция с реальным Travelpayouts API

В файле `travelpayouts.go` есть заготовка функции `callTravelPayoutsAPI()`. 
Замените вызов `buildAffiliateLink()` на `callTravelPayoutsAPI()` когда будет 
доступна точная документация по API.

## Лицензия

MIT License

## Поддержка

При возникновении вопросов или проблем создайте issue в репозитории.
