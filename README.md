# MedSched-SaaS

MedSched-SaaS — прототип системы управления медицинским расписанием клиники:
пользователи и роли, пациенты, врачи, расписания, слоты, запись на приём,
уведомления, ML-прогноз no-show, отчёты и демонстрационная интеграция с EHR/EMR.

## Состав проекта

- `core` — Go API: Gin, GORM, PostgreSQL, JWT, RBAC, расписания, записи, отчёты,
  ML-management и интеграции.
- `notificator` — Go-сервис уведомлений: Redis, RabbitMQ, тестовые отправители
  Email/SMS/Telegram и callback в Core.
- `ml-service` — Python FastAPI-сервис прогнозирования no-show: pandas,
  scikit-learn, XGBoost.
- `web` — React/Vite/TypeScript веб-интерфейс.
- `diagrams` — PlantUML-диаграммы для диплома и презентации.

## Порты

| Компонент | Адрес |
|---|---|
| Core API | `http://localhost:8081` |
| Notification Service | `http://localhost:8082` |
| ML Service | `http://localhost:8000` |
| Web UI | адрес из вывода `npm run dev`, обычно `http://localhost:5173` |
| PostgreSQL | `localhost:5432` |
| RabbitMQ | `localhost:5672`, management UI `http://localhost:15672` |
| Redis | `localhost:6379` |

## Быстрый запуск для демонстрации

Команды запускать из корня проекта `/home/slava/go/src/diplom`, если не указано
иное.

### 1. Запустить инфраструктуру

```bash
docker compose up -d medcore-db rabbitmq redis
docker compose ps
```

PostgreSQL, RabbitMQ и Redis берут параметры из `.env`.

### 2. Обучить или проверить ML-модель

Если `ml-service/artifacts/active_model.json` и модели уже есть, этот шаг можно
пропустить. Для полной подготовки:

```bash
cd ml-service
python3 -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt
python train.py --activate
```

После обучения оставить терминал в `ml-service` или открыть новый.

### 3. Запустить Core

В отдельном терминале:

```bash
cd core
`CONFIG_PATH=./config/local.yml go run ./cmd/core`
```

Core должен подняться на `http://localhost:8081`.

### 4. Заполнить seed-данные

В новом терминале:

```bash
cd core
CONFIG_PATH=./config/local.yml go run ./cmd/seed
```

Seed рассчитывает даты относительно момента запуска. Если выполнить его
17 июня 2026 года, будущие слоты будут созданы на 18, 19 и 22 июня.

Для чистой демонстрации лучше запускать seed на свежей базе. Повторный seed без
очистки БД не удаляет старые слоты и записи.

### 5. Запустить ML Service

В отдельном терминале:

```bash
cd ml-service
. .venv/bin/activate
ML_ADMIN_TOKEN=dev-ml-admin-token uvicorn app.main:app --host 0.0.0.0 --port 8000
```

Если активной модели нет, сервис использует fallback-эвристику. Для защиты лучше
использовать обученную XGBoost-модель.

### 6. Запустить Notification Service

В отдельном терминале:

```bash
cd notificator
CONFIG_PATH=./config/local.yml go run ./cmd/notificator
```

В локальном режиме отправители тестовые. Сервис планирует уведомления через
RabbitMQ/Redis и отправляет callback статуса доставки в Core.

### 7. Запустить Frontend

В отдельном терминале:

```bash
cd web
npm install
npm run dev
```

Открыть адрес из вывода Vite.

## Проверка после запуска

Проверить, что сервисы отвечают:

```bash
curl http://localhost:8081/api/v1/health
curl http://localhost:8082/health
curl http://localhost:8000/health
```

Если `curl` недоступен, открыть эти адреса в браузере.

## Тестовые аккаунты

Общий пароль seed-пользователей:

```text
password123
```

| Роль | Email |
|---|---|
| Администратор | `admin@example.com` |
| Регистратор | `registrator@example.com` |
| Регистратор 2 | `registrator2@example.com` |
| Менеджер | `manager@example.com` |
| IT-поддержка | `it@example.com` |
| Врач-терапевт | `doctor.therapist@example.com` |
| Пациент | `patient01@example.com` |
| Свободный patient-аккаунт | `available.patient@example.com` |

Подробнее о seed-данных: [SEED_DATA.md](SEED_DATA.md).

## Что создаёт seed

- 5 сотрудников разных ролей;
- 5 врачей разных специальностей;
- 20 пациентов с медицинскими профилями;
- 16 привязанных patient-аккаунтов;
- один свободный patient-аккаунт для демонстрации привязки;
- расписания и будущие слоты;
- 60 завершённых исторических приёмов для ML-выборки;
- будущие, подтверждённые и отменённые записи;
- уведомления в разных статусах;
- демонстрационные прогнозы `LOW`, `MEDIUM`, `HIGH`;
- предложение переноса записи;
- Demo EHR-провайдер, связи пациентов и журнал синхронизации.

## Основной сценарий демонстрации

1. Войти как `registrator@example.com`.
2. Открыть пациента **Иванов Иван Иванович**.
3. Показать карту пациента и медицинский профиль.
4. Перейти к записи на приём.
5. Выбрать врача, будущий свободный слот и создать запись.
6. Войти как `doctor.therapist@example.com`.
7. Показать календарь врача и созданную запись.
8. Открыть профиль пациента из записи.
9. Рассчитать или показать сохранённый ML-прогноз no-show.
10. Показать уведомления и, если есть высокий риск, предложение переноса.
11. Войти как `patient01@example.com` и показать, что пациент видит только свои
    записи, профиль и пациентские действия.
12. При наличии времени войти как `manager@example.com` и показать отчёты.

Подробный сценарий защиты: [DEFENSE_SCRIPT.md](DEFENSE_SCRIPT.md).

## ML Service

Обычное обучение на публичном датасете:

```bash
cd ml-service
. .venv/bin/activate
python train.py --activate
```

Обучение на фактах, накопленных Core:

```bash
cd ml-service
. .venv/bin/activate
CORE_SERVICE_TOKEN=dev-ml-dataset-token python train.py \
  --source core \
  --core-url http://localhost:8081 \
  --core-limit 5000
```

Модели сохраняются в:

```text
ml-service/artifacts/models/{model_version}/
```

Активная модель задаётся в:

```text
ml-service/artifacts/active_model.json
```

После активации модели через Core или API можно перезагрузить ML Service:

```bash
curl -X POST http://localhost:8000/model/reload \
  -H "Authorization: Bearer dev-ml-admin-token"
```

Подробнее: [ml-service/README.md](ml-service/README.md).

## EHR/EMR

В проекте есть демонстрационный провайдер `Demo EHR`. Он позволяет:

- привязать пациента к внешнему идентификатору;
- загрузить предварительные изменения;
- выбрать поля для синхронизации;
- применить изменения в медицинский профиль;
- сохранить событие в журнал интеграций.

Также есть базовый FHIR R4 adapter. Он включается переменными окружения:

```bash
MEDCORE_FHIR_ENABLED=true
MEDCORE_FHIR_BASE_URL=https://ehr.example.org/fhir
MEDCORE_FHIR_BEARER_TOKEN=<token>
MEDCORE_FHIR_CODE=clinic-fhir
MEDCORE_FHIR_NAME="Clinic FHIR R4"
```

После перезапуска Core провайдер появится в разделе «Интеграции».

## Проверки

Core:

```bash
cd core
go test ./...
```

Интеграционный demo-flow тест с реальной PostgreSQL и fake ML server:

```bash
cd core
RUN_INTEGRATION=1 TEST_DATABASE_DSN="postgres://medcoredb:medcoredb@localhost:5432/medcoredb?sslmode=disable" go test -v ./internal/integration
```

ML:

```bash
cd ml-service
python3 -m compileall app train.py
```

Frontend:

```bash
cd web
npm run lint
npm run build
```

## Частые проблемы

### `connect: connection refused` при прогнозе

ML Service не запущен или запущен не на `localhost:8000`.

### `503 Service Unavailable` при активации модели

Проверить, что в `ml-service/artifacts/models` есть выбранная версия модели, а
ML Service запущен с `ML_ADMIN_TOKEN=dev-ml-admin-token`.

### В онлайн-записи нет свободных слотов

Скорее всего, seed был запущен давно, и будущие слоты стали прошедшими.
Запустить seed на свежей базе перед демонстрацией.

### Уведомления не отправляются

Проверить, что запущены RabbitMQ, Redis и `notificator`. В прототипе реальные
Email/SMS не отправляются, используются тестовые адаптеры.

## Дополнительные материалы

- [DEFENSE_SCRIPT.md](DEFENSE_SCRIPT.md) — текст защиты и демонстрации.
- [ML_DEFENSE_CHEATSHEET.md](ML_DEFENSE_CHEATSHEET.md) — ответы по ML-слайду.
- [MANUAL_DEMO_CHECKLIST.md](MANUAL_DEMO_CHECKLIST.md) — чеклист ручной проверки.
- [WORK_SUMMARY.md](WORK_SUMMARY.md) — срез проделанной работы.
- [diagrams/README.md](diagrams/README.md) — структура диаграмм.
