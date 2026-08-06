# GophProfile

Микросервис для управления аватарками пользователей: загрузка, хранение, обработка и раздача через REST API.

## Стек

- Go, Chi
- PostgreSQL — метаданные
- MinIO / S3 — изображения

## Локальный запуск

Нужны PostgreSQL и MinIO (или `docker compose up postgres minio minio-init`):

```bash
export DATABASE_DSN="postgres://avatar:avatar@localhost:5432/avatar?sslmode=disable"
export S3_ENDPOINT=localhost:9000
export S3_ACCESS_KEY=minioadmin
export S3_SECRET_KEY=minioadmin
export S3_BUCKET=avatars

go run ./cmd/server
```

Сервис: http://localhost:8080  
Health: http://localhost:8080/health

## Docker Compose

Поднимает PostgreSQL, MinIO, RabbitMQ, server и worker:

```bash
docker compose up --build
```

| Сервис | URL |
|--------|-----|
| API | http://localhost:8080 |
| Веб (загрузка) | http://localhost:8080/web/upload |
| MinIO Console | http://localhost:9001 (minioadmin / minioadmin) |
| RabbitMQ Management | http://localhost:15672 (guest / guest) |

Переменные окружения — см. `.env.example`. В compose включён `RABBITMQ_ENABLED=true`: server публикует события, worker создаёт миниатюры и удаляет файлы.

## RabbitMQ и worker

По умолчанию в Docker worker уже запущен. Локально:

```bash
export RABBITMQ_ENABLED=true
go run ./cmd/worker       # отдельный процесс
go run ./cmd/server
```

RabbitMQ Management UI: http://localhost:15672 (guest / guest)

Остановка:

```bash
docker compose down
```

## API

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/avatars` | Загрузка (header `X-User-ID`) |
| GET | `/api/v1/avatars/{id}` | Получение изображения |
| GET | `/api/v1/avatars/{id}/metadata` | Метаданные |
| DELETE | `/api/v1/avatars/{id}` | Удаление (header `X-User-ID`) |
| GET | `/api/v1/users/{user_id}/avatar` | Аватар пользователя или placeholder (200, не 404) |
| GET | `/api/v1/users/{user_id}/avatars` | Список аватарок (пустой массив, если нет) |
| GET | `/health` | Проверка PostgreSQL, S3 и RabbitMQ (если включён) |

**Placeholder** применяется только к `GET /users/{user_id}/avatar`: если у email нет аватарки в сервисе, возвращается встроенная заглушка. Ручки по `avatar_id` (`GET /avatars/{id}`, metadata, delete) работают с конкретной записью и отдают **404**, если её нет.

Ответы `GET /avatars/{id}` и `GET /users/{user_id}/avatar` включают `Cache-Control: max-age=86400` и `ETag` (SHA-256). При совпадении `If-None-Match` возвращается **304 Not Modified**.

Пример загрузки:

```bash
curl -X POST http://localhost:8080/api/v1/avatars \
  -H "X-User-ID: user@example.com" \
  -F "file=@photo.jpg"
```

## Веб-интерфейс

SPA на базе [шаблона Yandex Practicum](https://github.com/Yandex-Practicum/go-avatar-service-template), адаптирован под API GophProfile (поле `file`, галерея).

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/` | редирект на `/web/upload` |
| GET | `/web/upload` | форма загрузки (preview, вызов REST API) |
| POST | `/web/upload` | загрузка через HTML-форму (`userId`, `file`) |
| GET | `/web/gallery/{user_id}` | галерея аватарок пользователя |

Статика вшита в бинарь (`embed`), отдельный каталог в Docker не нужен.

## Тесты

Unit-тесты:

```bash
go test ./...
```

Интеграционные тесты (нужен Docker):

```bash
go test -tags=integration ./tests/integration/...
```

Линтер ([golangci-lint](https://golangci-lint.run/)):

```bash
golangci-lint run ./...
```

Конфигурация — `.golangci.yml`.

## Observability

Стек поднимается вместе с `docker compose up --build`. Приложение экспортирует traces, logs и metrics через **OTLP** в **OpenTelemetry Collector**, который маршрутизирует данные дальше.

```
server / worker ──OTLP──→ otel-collector ──→ Jaeger (traces)
                                        ──→ OpenSearch (logs)
                                        ──→ Prometheus (metrics)
```

| UI | URL | Логин |
|----|-----|-------|
| Grafana | http://localhost:3000 | admin / admin |
| Jaeger | http://localhost:16686 | — |
| Prometheus | http://localhost:9090 | — |
| OpenSearch Dashboards | http://localhost:5601 | — |

Grafana dashboards (папка **Avatar Service**):
- **Avatar Service Overview** — метрики (HTTP, uploads, processing, DB, RabbitMQ, S3)
- **Avatar Service Logs** — логи из OpenSearch, фильтр по Trace ID

### Переменные окружения

См. `.env.example`. По умолчанию observability **включена** (`OTEL_*_ENABLED=true`).

Для worker задайте отдельное имя сервиса:

```bash
OTEL_SERVICE_NAME=avatar-service-worker go run ./cmd/worker
```

В `docker-compose.yml` для worker уже указано `OTEL_SERVICE_NAME=avatar-service-worker`.

### Поиск логов

**Grafana → Avatar Service Logs** — введите Trace ID из Jaeger в переменную шаблона.

**OpenSearch Dashboards** — index pattern `avatar-logs*`, примеры запросов:

```
trace.id: "<trace_id из Jaeger>"
service.name: "avatar-service"
log.level: "error"
```

### Проверка end-to-end

1. Upload аватара через API или `/web/upload`
2. **Grafana** — растут панели Uploads / Request Rate
3. **Jaeger** — trace от HTTP до worker `process_upload`
4. **Grafana Logs** или **OpenSearch** — лог с тем же `trace.id`
