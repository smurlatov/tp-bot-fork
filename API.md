# Travelpayouts Go Service API

Легковесный Go сервис для создания аффилиатных ссылок через **реальный** Travelpayouts API.

## Эндпоинты

### 1. POST /api/getFromLink

Создает аффилиатную ссылку из прямой ссылки через Travelpayouts API.

**Запрос:**
```json
{
  "link": "https://www.booking.com/hotel/us/plaza.html",
  "token": "your_travelpayouts_token",
  "trs": "197987", 
  "marker": "339296"
}
```

**Ответ (успешный):**
```json
{
  "link": "https://c.travelpayouts.com/shortened_affiliate_link"
}
```

**Ответ (ошибка):**
```json
{
  "error": "Описание ошибки от Travelpayouts API"
}
```

### 2. POST /api/getFromBrand

Создает аффилиатную ссылку для бренда по его названию через Travelpayouts API.

**Запрос:**
```json
{
  "brand_name": "booking",
  "token": "your_travelpayouts_token", 
  "trs": "197987",
  "marker": "339296"
}
```

**Ответ (успешный):**
```json
{
  "link": "https://c.travelpayouts.com/shortened_affiliate_link"
}
```

**Ответ (ошибка):**
```json
{
  "error": "Бренд не найден: unknown_brand"
}
```

### 3. GET /health

Проверка состояния сервиса.

**Ответ:**
```json
{
  "status": "ok"
}
```

## Доступные бренды

Сервис поддерживает следующие бренды (можно расширить в базе данных):

- `booking` - https://www.booking.com
- `agoda` - https://www.agoda.com  
- `aviasales` - https://aviasales.com
- `hotels` - https://hotels.com
- `expedia` - https://www.expedia.com

## Параметры

- **link** - исходная ссылка для конвертации
- **brand_name** - название бренда из списка поддерживаемых
- **token** - токен Travelpayouts API (обязательный)
- **trs** - числовой идентификатор TRS (обязательный)
- **marker** - числовой маркер партнера (обязательный)
- **sub_id** - автоматически устанавливается в "social_tool_main"

## Внутренняя логика

Сервис использует **реальный Travelpayouts API**:
- Эндпоинт: `POST https://api.travelpayouts.com/links/v1/create`
- Автоматически добавляет `sub_id=social_tool_main`
- Использует сокращенные ссылки (`shorten: true`)
- Возвращает реальные аффилиатные ссылки от Travelpayouts

## Коды ошибок

- `400` - Некорректные параметры запроса
- `404` - Бренд не найден
- `500` - Внутренняя ошибка сервера или ошибка Travelpayouts API

## Примеры использования

### Curl

```bash
# Создание ссылки из прямой ссылки
curl -X POST http://localhost:8080/api/getFromLink \
  -H "Content-Type: application/json" \
  -d '{
    "link": "https://www.booking.com",
    "token": "your_real_token",
    "trs": "197987", 
    "marker": "339296"
  }'

# Создание ссылки по имени бренда
curl -X POST http://localhost:8080/api/getFromBrand \
  -H "Content-Type: application/json" \
  -d '{
    "brand_name": "booking",
    "token": "your_real_token",
    "trs": "197987",
    "marker": "339296"
  }'
```

### JavaScript

```javascript
// Создание ссылки из прямой ссылки
fetch('http://localhost:8080/api/getFromLink', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    link: 'https://www.booking.com',
    token: 'your_real_token',
    trs: '197987',
    marker: '339296'
  })
})
.then(response => response.json())
.then(data => console.log(data));

// Создание ссылки по имени бренда  
fetch('http://localhost:8080/api/getFromBrand', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    brand_name: 'booking',
    token: 'your_real_token', 
    trs: '197987',
    marker: '339296'
  })
})
.then(response => response.json())
.then(data => console.log(data));
```

## Требования к данным

- **token**: Валидный токен Travelpayouts API
- **trs**: Числовое значение TRS (например, 197987)
- **marker**: Числовое значение маркера (например, 339296)
- **link**: Валидный URL для конвертации
- **brand_name**: Имя бренда из поддерживаемого списка

## Логирование

Сервис логирует:
- Все входящие запросы
- Запросы к Travelpayouts API
- Ответы от Travelpayouts API
- Созданные аффилиатные ссылки
- Ошибки с полным контекстом 