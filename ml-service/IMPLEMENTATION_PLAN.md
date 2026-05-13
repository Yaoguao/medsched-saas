# ML Service Implementation Plan

## 1. MVP сервис

- Поднять отдельный Python/FastAPI сервис.
- Реализовать `GET /health`.
- Реализовать `POST /predict/no-show`.
- Вернуть `no_show_probability`, `risk_level`, `model_version`, `algorithm`, `top_factors`.
- Использовать fallback-эвристику, если обученный артефакт еще не создан.

## 2. Обучение модели

- Использовать `healthcare_noshows.csv` как стартовый датасет.
- Использовать core endpoint `GET /api/v1/ml/training/no-show-dataset` для переобучения на накопленных фактах.
- Целевая переменная: `no_show = not Showed_up`.
- Подготовить признаки пациента, приема и истории.
- Обучить baseline `LogisticRegression`.
- Основная модель: `XGBoostClassifier`.
- Сохранять артефакт в `artifacts/no_show_model.joblib`.
- Сохранять метрики в `artifacts/metrics.json`.

## 3. Контракт с core

- Core собирает признаки из `Patient`, `PatientMedicalProfile`, `Appointment`, `TimeSlot`, `Doctor`.
- Core вызывает `POST /predict/no-show`.
- Core сохраняет ответ в `PredictionResult`.
- Для отчетов использовать сохраненную вероятность и risk level.

## 4. Улучшение признаков

- Реальные исторические признаки из БД: количество прошлых записей, no-show count, no-show rate.
- Добавить признаки врача и специальности.
- Добавить признаки уведомлений: было ли отправлено напоминание, канал, время до приема.
- Добавить временные признаки: день недели, час приема, lead time.

## 5. Качество и эксплуатация

- Сравнить baseline и XGBoost по ROC-AUC, average precision, recall.
- Зафиксировать версию модели.
- Добавить endpoint с информацией о модели.
- Переобучать отдельной командой `python train.py --source core --core-url http://localhost:8081`.
