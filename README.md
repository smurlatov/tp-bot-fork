# Travelpayouts Go Service

Легковесный Go сервис для создания аффилиатных ссылок через **реальный** Travelpayouts API.

## Особенности

- 🚀 Легковесный и быстрый Go сервис  
- 🔗 Два эндпоинта: создание ссылок из URL и по имени бренда
- ✅ **Реальная интеграция с Travelpayouts API** (`POST https://api.travelpayouts.com/links/v1/create`)
- 📊 Полное логирование всех операций
- 🗄️ SQLite база данных для хранения брендов
- 🐳 Docker поддержка для легкого развертывания
- 🚀 Готов для развертывания на Railway

## Архитектура

```
POST /api/getFromLink    - Создает аффилиатную ссылку из прямой ссылки через Travelpayouts API
POST /api/getFromBrand   - Создает аффилиатную ссылку по имени бренда через Travelpayouts API  
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

# Сервис запустится на http://localhost:8080
```

## Использование API

### 1. Создание ссылки из URL

```bash
curl -X POST http://localhost:8080/api/getFromLink \
  -H "Content-Type: application/json" \
  -d '{
    "link": "https://www.booking.com/hotel/us/plaza.html",
    "token": "YOUR_TRAVELPAYOUTS_TOKEN",
    "trs": "197987",
    "marker": "339296"
  }'
```

**Успешный ответ:**
```json
{
  "link": "https://yesim.tp.st/kn3kv29H?erid=2VtzqwiKLkx"
}
```

### 2. Создание ссылки по бренду

```bash
curl -X POST http://localhost:8080/api/getFromBrand \
  -H "Content-Type: application/json" \
  -d '{
    "brand_name": "booking",
    "token": "YOUR_TRAVELPAYOUTS_TOKEN",
    "trs": "197987",
    "marker": "339296"
  }'
```

## Управление брендами

Добавить новый бренд можно напрямую в SQLite базу данных:

```sql
INSERT INTO brands (brand_name, brand_link) 
VALUES ('booking', 'https://www.booking.com/');
```

Или программно через приложение (функционал можно расширить).

## Переменные окружения

- `PORT` - Порт для запуска сервера (по умолчанию: 8080)
- `DATABASE_PATH` - Путь к SQLite базе данных
- `GIN_MODE` - Режим Gin фреймворка (release/debug)

## Развертывание на Railway

1. Коммитим код в Git репозиторий
2. Подключаем репозиторий к Railway
3. Railway автоматически определит Dockerfile и развернет сервис
4. Получим публичный URL для использования API

## Структура проекта

```
tp-go-service/
├── main.go              # Основная логика сервера и маршруты
├── travelpayouts.go     # Интеграция с Travelpayouts API
├── go.mod              # Go модули и зависимости
├── Dockerfile          # Docker конфигурация
├── docker-compose.yml  # Docker Compose для локальной разработки
├── API.md              # Документация API
└── README.md           # Этот файл
```

## Логирование

Сервис использует структурированное логирование с `logrus`. Все запросы к API, ошибки и успешные операции логируются с соответствующими уровнями важности.

## Обработка ошибок

Сервис правильно обрабатывает ошибки от Travelpayouts API и возвращает понятные сообщения об ошибках пользователю.

## Технологии

- **Go 1.21** - Основной язык
- **Gin** - HTTP веб-фреймворк
- **GORM** - ORM для работы с базой данных
- **SQLite** - Легковесная база данных
- **Logrus** - Структурированное логирование
- **Docker** - Контейнеризация

## Лицензия

MIT License

## Поддержка

При возникновении вопросов или проблем создайте issue в репозитории.
