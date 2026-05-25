# backend-billing

Традиционная реализация биллинг-системы на Go с PostgreSQL для бакалаврской диссертации *"Exploring Kubernetes as a Platform for Business Logic Using Custom Resource Definitions and the Operator Pattern"*.

Этот проект представляет собой **альтернативный подход** к реализации биллинга по сравнению с `kube-billing` (Kubernetes Operator). Здесь используется классическая архитектура веб-приложения с REST API и реляционной БД.

---

## 📦 Технологии

| Компонент | Технология |
|-----------|------------|
| **Язык** | Go 1.25 |
| **HTTP-фреймворк** | Gorilla Mux |
| **ORM** | GORM |
| **База данных** | PostgreSQL 15 |
| **Трассировка** | OpenTelemetry → Jaeger |
| **Метрики** | Prometheus |
| **Контейнеризация** | Docker, Docker Compose |

---

## 🚀 Запуск проекта

### Через Docker Compose (рекомендуется)

```bash
cd backend-billing
docker-compose up --build
```

После запуска доступны:

| Сервис | URL | Описание |
|--------|-----|----------|
| **API** | `http://localhost:8080` | REST API биллинга |
| **PostgreSQL** | `localhost:5432` | База данных |
| **Jaeger UI** | `http://localhost:16686` | Трассировка запросов |
| **Prometheus** | `http://localhost:9090` | Метрики приложения |

### Переменные окружения

| Переменная | Значение по умолчанию | Описание |
|------------|----------------------|----------|
| `DATABASE_URL` | `host=localhost user=postgres password=password dbname=billing port=5432 sslmode=disable` | Подключение к БД |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `jaeger:4318` | Endpoint для трассировки |
| `PORT` | `8080` | Порт HTTP-сервера |

---

## 📡 API Endpoints

### Billing Plans

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `GET` | `/plans` | Список всех планов |
| `POST` | `/plans` | Создать новый план |

**Пример создания плана:**
```bash
curl -X POST http://localhost:8080/plans \
  -H "Content-Type: application/json" \
  -d '{"name": "pro", "price": "19.99", "currency": "USD", "billingPeriod": "monthly"}'
```

### Subscriptions

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `GET` | `/subscriptions` | Список всех подписок |
| `GET` | `/subscriptions/{id}` | Детали подписки по ID |
| `POST` | `/subscriptions` | Создать новую подписку |
| `POST` | `/subscriptions/{id}/cancel` | Отменить подписку |

**Пример создания подписки:**
```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"userId": "user-123", "planRef": "pro"}'
```

**Пример отмены подписки:**
```bash
curl -X POST http://localhost:8080/subscriptions/1/cancel
```

### Observability

| Endpoint | Описание |
|----------|----------|
| `GET /metrics` | Prometheus-метрики (созданные/отменённые подписки) |

---

## 🏗️ Структура проекта

```
backend-billing/
├── main.go                     # Точка входа: инициализация роутера, middleware, запуск сервера
├── db/
│   └── db.go                   # Подключение к PostgreSQL, авто-миграция схем, tracing
├── handlers/
│   ├── plan_handlers.go        # HTTP-обработчики для BillingPlan (GET/POST)
│   └── subscription_handlers.go # HTTP-обработчики для Subscription (CRUD + метрики)
├── models/
│   └── models.go               # GORM-модели: BillingPlan, Subscription
├── workers/
│   └── workers.go              # Фоновый воркер проверки истечения подписок
├── observability/
│   └── setup.go                # Инициализация логгера (slog) и трассировки (OpenTelemetry)
├── config/
│   └── prometheus.yml          # Конфигурация scrape для Prometheus
├── docker-compose.yml          # Оркестрация: API, PostgreSQL, Jaeger, Prometheus
└── Dockerfile                  # Multi-stage сборка Go-приложения
```

---

## 🗄️ Модели данных

### BillingPlan
```go
type BillingPlan struct {
    ID            uint           `gorm:"primaryKey"`
    Name          string         `gorm:"uniqueIndex"`  // Уникальное имя плана (например, "pro")
    Price         string         // Цена (строка для точности)
    Currency      string         // Валюта (USD, EUR, RUB)
    BillingPeriod string         // Период (monthly, yearly)
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     gorm.DeletedAt // Soft delete
}
```

### Subscription
```go
type Subscription struct {
    ID          uint           `gorm:"primaryKey"`
    UserID      string         // ID пользователя
    PlanRef     string         // Ссылка на BillingPlan.Name
    State       string         // Active, Cancelled, Expired
    LastPayment *time.Time     // Дата последней оплаты
    NextBilling *time.Time     // Дата следующего списания
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt // Soft delete
}
```

---

## 🔄 Бизнес-логика

### Создание подписки
При создании подписки автоматически:
1. Устанавливается `State = "Active"`
2. `LastPayment = time.Now()`
3. `NextBilling = LastPayment + 1 месяц`

### Проверка истечения (Background Worker)
- **Интервал**: каждые 1 минуту (для демонстрации)
- **Логика**: находит все подписки со `State = "Active"` и `NextBilling < now`
- **Действие**: обновляет `State = "Expired"`

### Метрики Prometheus
- `billing_subscriptions_created_total` — счётчик созданных подписок
- `billing_subscriptions_cancelled_total` — счётчик отменённых подписок

---

## 🧪 Разработка

### Локальный запуск без Docker
```bash
# Запустить PostgreSQL (например, через brew или docker)
export DATABASE_URL="host=localhost user=postgres password=password dbname=billing port=5432 sslmode=disable"
go run main.go
```

### Сборка бинарника
```bash
go build -o billing-api .
```

### Тесты
На момент написания тесты не реализованы. Для добавления:
```bash
go test ./...
```

---

## 🔍 Observability

### Логирование
- **Формат**: JSON (structured logging через `slog`)
- **Уровень**: `info`
- **Вывод**: stdout

### Трассировка
- **Инструмент**: OpenTelemetry SDK
- **Экспортер**: OTLP HTTP → Jaeger
- **Интеграция**: 
  - HTTP-роуты (через `otelmux.Middleware`)
  - GORM-запросы (через `tracing.NewPlugin()`)

### Метрики
- **Инструмент**: Prometheus client_golang
- **Сбор**: Prometheus scrape'ит `/metrics` endpoint каждые 5 секунд

---

## 📊 Сравнение с kube-billing

| Аспект | backend-billing | kube-billing |
|--------|-----------------|--------------|
| **Хранение состояния** | PostgreSQL | Kubernetes CRD (etcd) |
| **API** | REST HTTP | Kubernetes API (`kubectl apply`) |
| **Биллинг-цикл** | Фоновый воркер (ticker) | Reconcile + RequeueAfter |
| **Масштабирование** | Ручное / Docker Compose | Kubernetes HPA |
| **Отказоустойчивость** | Зависит от БД | Kubernetes самовосстановление |

---

## ⚠️ Известные ограничения

- **Тесты**: отсутствуют (требуется добавить для production-ready)
- **Валидация**: минимальная (нет проверки уникальности имени плана на уровне приложения)
- **Безопасность**: нет аутентификации/авторизации
- **Конфигурация**: хардкод в `docker-compose.yml`, нет поддержки `.env`

---

## 🎯 Контекст диссертации

Этот проект используется для **сравнительного анализа** двух подходов:
1. **Традиционный** (этот проект): явное управление состоянием через БД, HTTP API, фоновые воркеры
2. **Cloud-Native** (`kube-billing`): декларативное управление через CRD, reconcile-цикл, Kubernetes как платформа

Цель — продемонстрировать преимущества и недостатки каждого подхода для задач биллинга.
