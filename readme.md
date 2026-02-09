## DelayedNotifier

Простой сервис отложенных уведомлений на Go:

- **API + UI**: HTTP‑сервер, который отдаёт HTML/JS‑интерфейс и REST‑API.
- **PostgreSQL**: хранение сообщений и их статусов (источник правды).
- **RabbitMQ**: очередь задач на отправку.
- **Worker**: фоновой обработчик, который ждёт до `scheduled_at` и помечает уведомления как отправленные.
- **Redis**: кэш статусов для быстрых ответов по `GET /api/notifications/{id}/status`.

---

## Запуск

Требования:

- Docker + Docker Compose

Запуск всего стека:

```bash
docker-compose up --build
```

После старта:

- UI и API доступны по адресу `http://localhost:8080`.
- RabbitMQ UI: `http://localhost:15672` (`guest` / `guest`).
- PostgreSQL: `localhost:5432` (postgres/postgres).
- Redis: `localhost:6379`.

Остановка:

```bash
docker-compose down
```

---

## Архитектура

- `cmd/` — основной бинарь API/HTTP‑сервера.
- `worker/cmd/` — отдельный бинарь воркера (читает очередь и обновляет статусы).
- `internal/domain` — доменные сущности (`Message` и статусы).
- `internal/port` — интерфейсы (Repository, MessageQueue, StatusCache, Usecases).
- `internal/adapter/repository/postgres` — работа с PostgreSQL.
- `internal/adapter/rabbitmq` — продьюсер в RabbitMQ.
- `worker/internal/rabbitmq` — consumer из очереди.
- `internal/adapter/cache/redis` — кэш статусов на Redis.
- `internal/usecases` — бизнес‑логика.
- `internal/input/http` — HTTP‑слой (handlers + встроенный UI).

Поток данных:

1. Клиент отправляет запрос через UI или напрямую в API (`POST /api/notifications`).
2. Usecase:
   - генерирует `id`;
   - сохраняет сообщение в БД со статусом `Scheduled`;
   - отправляет полное сообщение в RabbitMQ.
3. Воркер читает сообщение из очереди, ждёт до `scheduled_at`, обновляет статус в БД на `Sent` и кладёт статус в Redis.
4. Запрос статуса (`GET /api/notifications/{id}/status`) сначала идёт в Redis, при промахе — в БД, затем кэширует результат.

---

## API

Базовый URL: `http://localhost:8080`

### Создать уведомление

- **POST** `/api/notifications`
- **Body (JSON)**:

```json
{
  "text": "Напомнить про созвон",
  "scheduled_at": "2026-02-10T11:00:00+03:00",
  "user_id": 1,
  "telegram_chat_id": 123456789
}
```

- **Ответ 201**:

```json
{ "id": "uuid" }
```

### Список уведомлений

- **GET** `/api/notifications`
- **Ответ 200** — массив объектов `Message`:

```json
[
  {
    "id": "uuid",
    "text": "Напомнить про созвон",
    "status": "Scheduled",
    "scheduled_at": "2026-02-10T11:00:00+03:00",
    "user_id": 1,
    "telegram_chat_id": 123456789
  }
]
```

### Статус уведомления

- **GET** `/api/notifications/{id}/status`
- **Ответ 200**:

```json
{ "status": "Scheduled" }
```

### Удаление уведомления

- **DELETE** `/api/notifications/{id}`
- **Ответ 204** — без тела.

---

## Тесты

Юнит‑тесты покрывают основную бизнес‑логику и HTTP‑слой:

- `internal/usecases/message_test.go` — поведение `MessageUsecases`
  (валидация `userId`, установка `id` и `status`, отправка в очередь,
  использование и наполнение кэша статусов).
- `internal/input/http/handler_test.go` — обработчики HTTP:
  создание, список, получение статуса и удаление уведомления.
- `internal/adapter/cache/redis/redis_test.go` — базовая проверка обработки
  отсутствующих ключей (поведение при `redis.Nil`).

Запуск тестов:

```bash
go test ./...
```

