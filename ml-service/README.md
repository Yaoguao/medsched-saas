# MedSched ML Service

ML-сервис прогнозирует риск no-show для записи пациента.

Стек:

- Python + FastAPI для HTTP API;
- pandas + scikit-learn для подготовки признаков;
- XGBoost как основной алгоритм;
- LogisticRegression как fallback, если XGBoost не установлен.

## Как работает

1. `train.py` читает `../healthcare_noshows.csv`.
2. Признаки приводятся к контракту MedSched: пациент, запись, история визитов.
3. Обучается бинарный классификатор `no_show = 1`.
4. Артефакт сохраняется в `artifacts/no_show_model.joblib`.
5. API `POST /predict/no-show` возвращает вероятность no-show, уровень риска и основные факторы.

Если артефакт модели отсутствует, сервис использует простую эвристику `heuristic-001`. Это позволяет проверить интеграцию до обучения модели.

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
python train.py --source core --core-url http://localhost:8081 --core-limit 5000
```

Для такого обучения в core уже должны быть завершенные записи с известным исходом: `COMPLETED` или `NO_SHOW`.

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
