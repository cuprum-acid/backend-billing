# Billing API Documentation

Документация API для биллинг-системы.

## 📖 OpenAPI Спецификация

Полная спецификация API доступна в файле [`openapi.yaml`](./openapi.yaml).

## 🚀 Быстрый старт

### Просмотр документации

Используйте Swagger UI или другие инструменты для просмотра OpenAPI спецификации:

1. **Swagger UI (онлайн)**:
   - Откройте https://editor.swagger.io
   - Загрузите файл `docs/openapi.yaml`

2. **Swagger UI (локально)**:
   ```bash
   # Установите swagger-ui
   docker run -d -p 8081:8080 -e SWAGGER_JSON=/openapi.yaml \
     -v $(pwd)/docs:/openapi.yaml swaggerapi/swagger-ui
   ```
   - Откройте http://localhost:8081

3. **VS Code Extension**:
   - Установите расширение "Swagger Viewer"
   - Откройте `docs/openapi.yaml`

## 📡 Endpoints

### Health Checks

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/health` | Базовая проверка здоровья |
| GET | `/ready` | Проверка готовности (с проверкой БД) |

**Пример ответа:**
```json
{
  "status": "ok",
  "timestamp": "2026-05-25T10:00:00+03:00"
}
```

### Billing Plans

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/plans` | Список всех планов |
| POST | `/plans` | Создать новый план |

**Пример создания плана:**
```bash
curl -X POST http://localhost:8080/plans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "pro",
    "price": "19.99",
    "currency": "USD",
    "billingPeriod": "monthly"
  }'
```

**Пример ответа:**
```json
{
  "id": 1,
  "name": "pro",
  "price": "19.99",
  "currency": "USD",
  "billingPeriod": "monthly",
  "createdAt": "2026-05-25T10:00:00+03:00",
  "updatedAt": "2026-05-25T10:00:00+03:00"
}
```

### Subscriptions

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/subscriptions` | Список всех подписок |
| GET | `/subscriptions/{id}` | Получить подписку по ID |
| POST | `/subscriptions` | Создать подписку |
| POST | `/subscriptions/{id}/cancel` | Отменить подписку |

**Пример создания подписки:**
```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "planRef": "pro"
  }'
```

**Пример ответа:**
```json
{
  "id": 1,
  "userId": "user-123",
  "planRef": "pro",
  "state": "Active",
  "lastPayment": "2026-05-25T10:00:00+03:00",
  "nextBilling": "2026-06-25T10:00:00+03:00",
  "createdAt": "2026-05-25T10:00:00+03:00",
  "updatedAt": "2026-05-25T10:00:00+03:00"
}
```

### Observability

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/metrics` | Prometheus метрики |

**Пример метрик:**
```
# HELP billing_subscriptions_created_total The total number of created subscriptions
# TYPE billing_subscriptions_created_total counter
billing_subscriptions_created_total 5

# HELP billing_subscriptions_canceled_total The total number of canceled subscriptions
# TYPE billing_subscriptions_canceled_total counter
billing_subscriptions_canceled_total 2

# HELP http_rate_limit_requests_total Total number of requests tracked by rate limiter
# TYPE http_rate_limit_requests_total counter
http_rate_limit_requests_total{endpoint="/plans"} 100
```

## 🔐 Аутентификация

API поддерживает два метода аутентификации:

1. **API Key** (заголовок `X-API-Key`)
2. **Bearer Token** (JWT в заголовке `Authorization`)

> ⚠️ **Примечание**: В текущей версии аутентификация не реализована. Эти методы зарезервированы для будущего расширения.

## ❌ Коды ошибок

| Код | Описание |
|-----|----------|
| 400 | Bad Request — некорректный запрос |
| 404 | Not Found — ресурс не найден |
| 409 | Conflict — конфликт (например, дубликат плана) |
| 429 | Too Many Requests — превышен лимит запросов |
| 500 | Internal Server Error — внутренняя ошибка |
| 503 | Service Unavailable — сервис недоступен |

**Пример ответа с ошибкой:**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "validation failed",
  "fields": {
    "name": "This field is required",
    "price": "Invalid price format"
  }
}
```

## 📊 Модели данных

### BillingPlan

| Поле | Тип | Описание |
|------|-----|----------|
| id | integer | Уникальный идентификатор |
| name | string | Уникальное имя плана (3-50 символов) |
| price | string | Цена (строка для точности) |
| currency | string | Валюта (USD, EUR, RUB, GBP, KZT) |
| billingPeriod | string | Период (monthly, yearly) |
| createdAt | datetime | Дата создания |
| updatedAt | datetime | Дата обновления |

### Subscription

| Поле | Тип | Описание |
|------|-----|----------|
| id | integer | Уникальный идентификатор |
| userId | string | ID пользователя (1-255 символов) |
| planRef | string | Ссылка на план (имя плана) |
| state | string | Статус (Active, Canceled, Expired) |
| lastPayment | datetime | Дата последней оплаты |
| nextBilling | datetime | Дата следующего списания |
| createdAt | datetime | Дата создания |
| updatedAt | datetime | Дата обновления |

## 🛡️ Rate Limiting

API использует rate limiting для защиты от злоупотреблений:

- **Лимит**: 10 запросов в секунду
- **Burst**: до 20 запросов
- **Метод**: Token bucket (per-IP)

При превышении лимита возвращается ответ `429 Too Many Requests`.

## 🧪 Тестирование

### Использование curl

```bash
# Health check
curl http://localhost:8080/health

# List plans
curl http://localhost:8080/plans

# Create plan
curl -X POST http://localhost:8080/plans \
  -H "Content-Type: application/json" \
  -d '{"name":"pro","price":"19.99","currency":"USD","billingPeriod":"monthly"}'

# Create subscription
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"userId":"user-123","planRef":"pro"}'

# Cancel subscription
curl -X POST http://localhost:8080/subscriptions/1/cancel
```

### Использование Postman

1. Импортируйте `docs/openapi.yaml` в Postman
2. Коллекция автоматически создастся со всеми endpoints
3. Настройте environment variables (base URL, auth tokens)

### Использование Swagger UI

1. Откройте Swagger UI (см. выше)
2. Выберите endpoint
3. Нажмите "Try it out"
4. Заполните параметры
5. Нажмите "Execute"

## 📈 Метрики

API экспортирует метрики Prometheus:

- `billing_subscriptions_created_total` — создано подписок
- `billing_subscriptions_canceled_total` — отменено подписок
- `http_rate_limit_requests_total` — запросов по rate limiter
- `http_rate_limit_hits_total` — срабатываний rate limiter

## 📝 Changelog

### v1.0.0 (2026-05-25)

- ✅ Initial release
- ✅ Billing plans CRUD
- ✅ Subscriptions lifecycle
- ✅ Health checks
- ✅ Prometheus metrics
- ✅ Rate limiting
- ✅ OpenAPI 3.0 documentation
