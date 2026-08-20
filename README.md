# GophProfile

Микросервис для управления аватарками: загрузка, хранение, обработка миниатюр и раздача через REST API.

## Стек

- Go, Chi
- PostgreSQL — метаданные
- MinIO / S3 — изображения
- RabbitMQ — асинхронная обработка (worker)
- Kubernetes + Helm — деплой

## Архитектура

```mermaid
flowchart LR
  Client --> Ingress
  subgraph k8s [Kubernetes namespace avatar]
    Ingress["Ingress nginx<br/>avatars.example.com"] --> SVC["Service avatar-service<br/>:80 http, :9090 metrics"]
    SVC --> Server["Deployment avatar-service<br/>HPA 2..10<br/>:8080 http, :9090 metrics"]
    Worker["Deployment avatar-worker<br/>:9090 metrics"]
    WSVC["Service avatar-worker-metrics<br/>headless, :9090"] --> Worker
    Migrate["Helm hook Job migrate"]
    Server --> PG[(PostgreSQL)]
    Server --> S3[(MinIO / S3)]
    Server --> MQ[[RabbitMQ]]
    Worker --> PG
    Worker --> S3
    Worker --> MQ
    Migrate --> PG
    SM["ServiceMonitor server + worker"] --> SVC
    SM --> WSVC
  end
  SM --> Prom[Prometheus]
  Prom --> Grafana
```

Server принимает HTTP, пишет метаданные в Postgres и оригинал в S3, публикует событие в RabbitMQ. Worker режет миниатюры и удаляет объекты. Миграции — embedded SQL: либо процесс применяет их сам (`POSTGRES_AUTO_MIGRATE=true`), либо владелец миграций — `cmd/migrate` (Helm hook, `POSTGRES_AUTO_MIGRATE=false`).

`/metrics` слушает отдельный порт `9090` (`METRICS_ADDR`) и не публикуется через Ingress: скрейпить его может только Prometheus из namespace мониторинга (NetworkPolicy). Публичный роутер отдаёт только API, `/health`, `/health/live` и документацию.

В Kubernetes: liveness на `/health/live` (без зависимостей, чтобы падение S3 не рестартовало поды), readiness на `/health` (с зависимостями), секреты через `secretKeyRef`, NetworkPolicy, non-root `65532`, HPA по CPU/RAM. ServiceAccount'ы без Role/RoleBinding — приложению не нужен доступ к API кластера, токен не монтируется. PSP нет (снят в 1.25) — вместо него Pod Security Admission и SecurityContext.

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

| URL | Назначение |
|-----|------------|
| http://localhost:8080 | API |
| http://localhost:8080/health | Readiness (зависимости) |
| http://localhost:8080/health/live | Liveness |
| http://localhost:9090/metrics | Prometheus |
| http://localhost:8080/docs | Swagger UI |
| http://localhost:8080/openapi.yaml | OpenAPI spec |

## Docker Compose

Поднимает PostgreSQL, MinIO, RabbitMQ, server, worker и observability:

```bash
docker compose up --build
```

| Сервис | URL |
|--------|-----|
| API | http://localhost:8081 |
| Веб | http://localhost:8081/web/upload |
| Swagger | http://localhost:8081/docs |
| Метрики server | http://localhost:9091/metrics |
| Метрики worker | http://localhost:9092/metrics |
| MinIO Console | http://localhost:9001 (minioadmin / minioadmin) |
| RabbitMQ | http://localhost:15672 (guest / guest) |
| Grafana | http://localhost:3001 (admin / admin) |
| Prometheus | http://localhost:19090 |
| Jaeger | http://localhost:16686 |
| OpenSearch Dashboards | http://localhost:5601 |

Переменные — `.env.example`. В compose `RABBITMQ_ENABLED=true`.

Локально с worker:

```bash
export RABBITMQ_ENABLED=true
go run ./cmd/worker
go run ./cmd/server
```

```bash
docker compose down
```

## Kubernetes

Образ содержит `server`, `worker` и `migrate`:

```bash
docker build -t avatar-service:latest .
```

Сырые манифесты (Rancher Desktop, nginx ingress, metrics-server). Namespace здесь создаётся вместе с метками Pod Security Admission, миграции применяет само приложение (`POSTGRES_AUTO_MIGRATE=true`), количество реплик server отдано HPA:

```bash
kubectl apply -k deploy/k8s
# после установки Prometheus Operator:
kubectl apply -k deploy/k8s/monitoring
```

Helm (предпочтительно) — здесь миграции отданы Job-хуку, поэтому в ConfigMap `POSTGRES_AUTO_MIGRATE=false`:

```bash
helm upgrade --install gophprofile deploy/helm/gophprofile \
  -n avatar --create-namespace \
  -f deploy/helm/gophprofile/values-local.yaml

kubectl label namespace avatar --overwrite \
  pod-security.kubernetes.io/enforce=baseline \
  pod-security.kubernetes.io/audit=restricted \
  pod-security.kubernetes.io/warn=restricted
```

Прод: `-f deploy/helm/gophprofile/values-prod.yaml`, секрет создать заранее (`existingSecret`). Так как зависимости внешние (`dependencies.enabled=false`), egress наружу описывается явными правилами `networkPolicy.externalEgress` (CIDR + порты Postgres, RabbitMQ, S3) — дырки `0.0.0.0/0` нет; CIDR в `values-prod.yaml` нужно заменить на адреса своей инфраструктуры.

Ingress: `127.0.0.1 avatars.example.com` в `/etc/hosts`.

Миграции: Job-хук вызывает `/app/migrate` (по умолчанию `post-install,pre-upgrade`). Повторный прогон безопасен (advisory lock). `pre-install` допустим только с внешним секретом: чарт-секрет на этой фазе ещё не существует, поэтому такая комбинация останавливает установку с явной ошибкой.

Метрики: ServiceMonitor'ы для server и worker смотрят на порт `metrics` (9090). Для worker создаётся headless Service `avatar-worker-metrics` — без него метрики воркера в кластере не собирались бы.

### Проверка деплоя

Образ собирается локально и тянется с `pullPolicy: IfNotPresent`, поэтому он должен лежать в том же образном хранилище, что использует k3s. При движке dockerd достаточно `docker build`, при containerd нужен namespace `k8s.io`:

```bash
docker build -t avatar-service:latest .
# либо, если в Rancher Desktop выбран containerd:
nerdctl --namespace k8s.io build -t avatar-service:latest .
```

Установка и ожидание готовности всех подов:

```bash
helm upgrade --install gophprofile deploy/helm/gophprofile \
  -n avatar --create-namespace \
  -f deploy/helm/gophprofile/values-local.yaml --wait --timeout 10m

kubectl -n avatar get pods,svc,ingress,hpa
kubectl -n avatar get jobs        # хук миграций
kubectl -n avatar logs job/avatar-migrate
```

Пробы и доступность через Ingress (`127.0.0.1 avatars.example.com` в `/etc/hosts`):

```bash
kubectl -n avatar rollout status deploy/avatar-service
kubectl -n avatar describe pod -l app=avatar-service | grep -A2 -E 'Liveness|Readiness'
curl -i http://avatars.example.com/health
curl -i http://avatars.example.com/health/live
curl -s http://avatars.example.com/api/v1/users/user@example.com/avatar -o /dev/null -w '%{http_code}\n'
```

Метрики: `/metrics` наружу не смотрит, поэтому проверяется через port-forward, а сбор — по таргетам Prometheus.

```bash
kubectl -n avatar port-forward svc/avatar-service 9090:9090 &
curl -s localhost:9090/metrics | grep -c http_server_requests
kubectl -n avatar port-forward svc/avatar-worker-metrics 9091:9090 &
curl -s localhost:9091/metrics | grep -c avatar_
kubectl -n avatar get servicemonitor   # включается --set serviceMonitor.enabled=true
```

В `values-local.yaml` ServiceMonitor выключен, потому что без установленного Prometheus Operator его CRD в кластере нет. После установки оператора включить флагом и проверить, что оба таргета (`avatar-service`, `avatar-worker-metrics`) в состоянии `up` на странице Targets.

HPA и балансировка: под нагрузкой растёт число реплик, а запросы расходятся по подам (видно по логам с префиксом пода).

```bash
kubectl -n avatar get hpa avatar-service-hpa -w
hey -z 60s -c 50 http://avatars.example.com/health/live   # или ab -t 60 -c 50
kubectl -n avatar logs -l app=avatar-service --prefix --tail=20
```

Graceful shutdown: при удалении пода в логах должно быть завершение работы, а нагрузка не должна получить 5xx.

```bash
kubectl -n avatar delete pod -l app=avatar-service --wait=false
kubectl -n avatar logs -l app=avatar-service --prefix --tail=50 | grep -i shut
```

NetworkPolicy: сервис принимает HTTP только от ingress-контроллера, поэтому запрос из постороннего пода того же namespace должен упасть в таймаут (`000`), тогда как тот же путь через Ingress отвечает `200`.

```bash
kubectl -n avatar run netcheck --rm -it --image=curlimages/curl --restart=Never -- \
  curl -m 3 -s -o /dev/null -w '%{http_code}\n' http://avatar-service/health
```

## API

Спецификация: [docs/openapi.yaml](docs/openapi.yaml), в рантайме — `/docs` и `/openapi.yaml`. Сама спецификация отдаётся сервисом, а оболочку Swagger UI страница `/docs` подгружает с `unpkg.com` силами браузера, поэтому в закрытом контуре смотреть спецификацию нужно через `/openapi.yaml`.

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/v1/avatars` | Загрузка (`X-User-ID`, поле `file`) |
| GET | `/api/v1/avatars/{id}` | Изображение (`size`, `format`) |
| GET | `/api/v1/avatars/{id}/metadata` | Метаданные |
| DELETE | `/api/v1/avatars/{id}` | Удаление (`X-User-ID`) |
| GET | `/api/v1/users/{user_id}/avatar` | Аватар или placeholder (200) |
| DELETE | `/api/v1/users/{user_id}/avatar` | Удалить аватар пользователя |
| GET | `/api/v1/users/{user_id}/avatars` | Список (пустой массив, если нет) |
| GET | `/health` | Readiness: PostgreSQL, S3, RabbitMQ |
| GET | `/health/live` | Liveness, без зависимостей |

`GET /users/{user_id}/avatar` без аватарки отдаёт заглушку (200). Ручки по `avatar_id` — **404**, если записи нет.

`GET` изображений: `Cache-Control: max-age=86400`, `ETag` (SHA-256), **304** при `If-None-Match`.

```bash
curl -X POST http://localhost:8080/api/v1/avatars \
  -H "X-User-ID: user@example.com" \
  -F "file=@photo.jpg"
```

Форматы: JPEG, PNG, WebP. Размеры: `original`, `100x100`, `300x300`.

## Веб-интерфейс

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/` | редирект на `/web/upload` |
| GET | `/web/upload` | форма загрузки |
| POST | `/web/upload` | загрузка (`userId`, `file`) |
| GET | `/web/gallery/{user_id}` | галерея |

Статика вшита в бинарь (`embed`).

## Тесты

```bash
go test ./...
go test -tags=integration ./tests/integration/...
golangci-lint run ./...
```

## Мониторинг и алерты

### Docker Compose

Трейсы и логи уходят push-моделью: OTLP → Collector → Jaeger / OpenSearch. Метрики — pull-модель: Prometheus скрейпит `server:9090` и `worker:9090` (job `avatar-service`), поэтому в Compose и в кластере используется один и тот же путь сбора и нет двойного учёта серий. Grafana dashboards в папке **Avatar Service**.

В Kubernetes метрики снимает ServiceMonitor с `/metrics` (порт `metrics`), OTLP по умолчанию выключен.

### Метрики

| Метрика | Смысл |
|---------|--------|
| `http_requests_total` / `http_request_duration_seconds` | HTTP |
| `avatars_uploads_total` / `avatars_deletes_total` | Загрузка и удаление |
| `avatars_processing_total` | Обработка во worker |
| `s3_operations_total` | S3 |
| `rabbitmq_messages_published_total` / `_consumed_total` | Очередь |
| `db_query_duration_seconds`, `db_connections_*` | Postgres |
| `circuit_breaker_state` | Состояние breaker'ов: 0 closed, 1 half-open, 2 open |

### Алерты

Правила: [deploy/prometheus/alerts.yml](deploy/prometheus/alerts.yml) (подключены в Compose).

| Алерт | Условие | Severity |
|-------|---------|----------|
| AvatarServiceDown | `up == 0` 1m | critical |
| AvatarServiceHighErrorRate | доля 5xx > 5% 5m | warning |
| AvatarServiceHighLatency | p95 > 1s 5m | warning |
| AvatarUploadFailures | ошибки upload > 0.1/s 5m | warning |
| AvatarCircuitBreakerOpen | `circuit_breaker_state == 2` 2m | warning |
| AvatarProcessingFailures | ошибки processing > 0.1/s 5m | warning |

В кластере с Prometheus Operator те же правила можно завести как `PrometheusRule`.

## Production

- Graceful shutdown по SIGTERM (HTTP и metrics-листенер), `terminationGracePeriodSeconds: 20` при таймауте остановки 10s
- Circuit breaker на S3 и RabbitMQ publish → 503; health-check `Ping` идёт мимо breaker'а, состояние breaker'ов видно в метрике `circuit_breaker_state`. Отмена запроса клиентом breaker не открывает (это не отказ зависимости), таймаут — открывает
- Rate limit (`go-chi/httprate`) по IP: `RATE_LIMIT_RPS` запросов в секунду → 429; middleware стоит после трейсинга, метрик и логирования, поэтому 429 попадают в наблюдаемость
- Адрес клиента берётся по явной модели доверия: `TRUSTED_PROXY_CIDRS` — сети reverse proxy, чьему `X-Forwarded-For` можно верить (в манифестах указан pod CIDR `10.42.0.0/16` для Rancher Desktop, в своём кластере замените). Пусто — заголовки игнорируются и берётся адрес соединения, поэтому клиент не может подделать ключ лимитера
- Resource requests/limits у приложения и у in-cluster зависимостей, HPA
- Non-root контейнеры, read-only root FS, drop ALL capabilities, токен ServiceAccount не монтируется
- Обновление ConfigMap/Secret перекатывает поды: в pod template проставляются аннотации `checksum/config` и `checksum/secret`
- Зависимости с RWO-томами (Postgres, MinIO) обновляются стратегией `Recreate`, образы зависимостей зафиксированы по версиям
- В обработке RabbitMQ гарантия at-least-once: сообщение, застигнутое shutdown'ом, переотправляется брокером (`nack` + redelivery)