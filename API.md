# Travelpayouts Go Service API

Легковесный Go сервис для создания аффилиатных ссылок через Travelpayouts API.

## Эндпоинты

### 1. POST /api/getFromLink

Создает аффилиатную ссылку из прямой ссылки.

**Запрос:**
```json
{
  "link": "https://www.booking.com/hotel/us/plaza.html",
  "token": "your_travelpayouts_token",
  "trs": "your_trs_parameter", 
  "marker": "your_marker"
}
```

**Ответ (успешный):**
```json
{
  "link": "https://www.booking.com/hotel/us/plaza.html?sub_id=social_tool_main&token=your_token&trs=your_trs&marker=your_marker&utm_source=travelpayouts&utm_medium=affiliate&utm_campaign=social_tool"
}
```

**Ответ (ошибка):**
```json
{
  "error": "Описание ошибки"
}
```

### 2. POST /api/getFromBrand

Создает аффилиатную ссылку для бренда по его названию.

**Запрос:**
```json
{
  "brand_name": "booking",
  "token": "your_travelpayouts_token", 
  "trs": "your_trs_parameter",
  "marker": "your_marker"
}
```

**Ответ (успешный):**
```json
{
  "link": "https://www.booking.com?sub_id=social_tool_main&token=your_token&trs=your_trs&marker=your_marker&utm_source=travelpayouts&utm_medium=affiliate&utm_campaign=social_tool"
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
- **token** - токен Travelpayouts API
- **trs** - параметр trs для отслеживания
- **marker** - маркер партнера
- **sub_id** - автоматически устанавливается в "social_tool_main"

## Коды ошибок

- `400` - Некорректные параметры запроса
- `404` - Бренд не найден
- `500` - Внутренняя ошибка сервера

## Примеры использования

### Curl

```bash
# Создание ссылки из прямой ссылки
curl -X POST http://localhost:8080/api/getFromLink \
  -H "Content-Type: application/json" \
  -d '{
    "link": "https://www.booking.com",
    "token": "your_token",
    "trs": "your_trs", 
    "marker": "your_marker"
  }'

# Создание ссылки по имени бренда
curl -X POST http://localhost:8080/api/getFromBrand \
  -H "Content-Type: application/json" \
  -d '{
    "brand_name": "booking",
    "token": "your_token",
    "trs": "your_trs",
    "marker": "your_marker"
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
    token: 'your_token',
    trs: 'your_trs',
    marker: 'your_marker'
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
    token: 'your_token', 
    trs: 'your_trs',
    marker: 'your_marker'
  })
})
.then(response => response.json())
.then(data => console.log(data));
``` 