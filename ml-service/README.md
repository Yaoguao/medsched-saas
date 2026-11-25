# MedSched ML Service

ML-сервис прогнозирует риск no-show для записи пациента.

Стек:

- Python + FastAPI для HTTP API;
- pandas + scikit-learn для подготовки признаков;
- XGBoost как основной алгоритм;
- LogisticRegression как явно выбираемый baseline.

## Как работает

1. `train.py` читает `../healthcare_noshows.csv`.
2. Признаки приводятся к контракту MedSched: пациент, запись, история визитов.
3. Данные делятся по времени на train, validation и test.
4. Пациенты из validation/test исключаются из train для защиты от утечки.
5. На validation подбираются пороги `MEDIUM` и `HIGH`.
6. Итоговые метрики рассчитываются только на test.
7. Версионированный артефакт сохраняется в `artifacts/models`.
8. API `POST /predict/no-show` возвращает вероятность no-show, уровень риска и основные факторы.

Если артефакт модели отсутствует, сервис использует простую эвристику `heuristic-001`. Это позволяет проверить интеграцию до обучения модели.

Случайное разделение оставлено только для разработки:

```bash
python train.py --split-mode random
```

Основной режим `temporal_group` требует не менее 30 завершенных приемов и наличия
обоих классов в train, validation и test. При недостатке реальных данных модель
не должна переобучаться.

## Запуск

```bash
cd ml-service
python3 -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt
python train.py
uvicorn app.main:app --host 0.0.0.0 --port 8000
```

Обучение на фактах из core:

```bash
CORE_SERVICE_TOKEN=dev-ml-dataset-token python train.py \
  --source core \
  --core-url http://localhost:8081 \
  --core-limit 5000
```

Для такого обучения в core уже должны быть завершенные записи с известным исходом: `COMPLETED` или `NO_SHOW`.
Токен должен совпадать с `MEDCORE_ML_DATASET_SERVICE_TOKEN` в Core.

В интерфейсе Core доступны два независимых варианта обучения:

- `Базовый healthcare_noshows.csv` — основная модель на полном исходном наборе;
- `Накопленные данные Core` — локальная модель на фактических посещениях клиники.

Обе версии сохраняются в общем реестре. Пользователь с ролью `ADMIN` или
`IT_SUPPORT` может активировать любую candidate или archived-версию. Метрики
моделей разных источников не сравниваются напрямую, потому что рассчитаны на
разных test-наборах.

Каждое обучение создает отдельный каталог:

```text
artifacts/models/{model_version}/
```

По умолчанию новая модель остается кандидатом. Для назначения ее активной:

```bash
python train.py --activate
```

Команда атомарно обновит `artifacts/active_model.json`. Запущенный ML Service
продолжит использовать прежнюю модель, пока не будет выполнена защищенная
перезагрузка:

```bash
ML_ADMIN_TOKEN=dev-ml-admin-token \
uvicorn app.main:app --host 0.0.0.0 --port 8000

curl -X POST http://localhost:8000/model/reload \
  -H "Authorization: Bearer dev-ml-admin-token"
```

Активация конкретной candidate-версии:

```bash
curl -X POST http://localhost:8000/model/activate \
  -H "Authorization: Bearer dev-ml-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"model_version":"no-show-xgboost-20260609-153000"}'
```

Обычно этот endpoint вызывает Core после проверки метрик candidate-модели.

Если manifest активной модели отсутствует, сервис продолжает поддерживать старый
артефакт `artifacts/no_show_model.joblib`.

## API

Healthcheck:

```http
GET /health
```

Prediction:

```http
POST /predict/no-show
```

```json
{
  "patient": {
    "age": 36,
    "gender": "M",
    "neighbourhood": "CENTER",
    "scholarship": false,
    "has_hypertension": true,
    "has_diabetes": false,
    "has_alcoholism": false,
    "has_handicap": false
  },
  "appointment": {
    "scheduled_at": "2026-05-13T15:00:00Z",
    "appointment_at": "2026-05-20T10:00:00Z",
    "booking_channel": "WEB",
    "visit_type": "PRIMARY",
    "doctor_specialty": "THERAPIST",
    "sms_received": true
  },
  "history": {
    "previous_appointments_count": 5,
    "previous_no_show_count": 2
  }
}
```

Response:

```json
{
  "no_show_probability": 0.42,
  "risk_level": "MEDIUM",
  "model_version": "xgboost-001",
  "algorithm": "xgboost",
  "top_factors": ["previous_no_show_rate", "lead_time_days", "sms_received"]
}
```
