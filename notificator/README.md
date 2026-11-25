# notificator

Отдельный сервис управления уведомлениями MedSched-SaaS.

## Возможности

- создание, получение, список, отмена и ручная отправка уведомлений;
- хранение уведомлений в Redis;
- планирование отправки через RabbitMQ;
- шаблоны уведомлений с переменными вида `{{patient_name}}`;
- каналы `email`, `telegram`, `sms` со stub-отправителями;
- retry policy с exponential backoff;
- история попыток отправки в поле `attempt_history`;
- dead-letter список после исчерпания retry;
- callback в core со статусами `SENT`, `FAILED`, `DELIVERED`;
- прием команд из очереди в формате envelope.

## HTTP

Base URL для локального конфига: `http://localhost:8082`.

- `POST /notify` - создать уведомление;
- `GET /notify` - список уведомлений;
- `GET /notify/{id}` - получить уведомление;
- `POST /notify/{id}/send` - отправить сразу;
- `DELETE /notify/{id}` - отменить;
- `GET /dead-letters` - список уведомлений, провалившихся после всех retry;
- `GET /dead-letters/{id}` - dead-letter по id уведомления;
- `POST /templates` - создать или обновить шаблон;
- `GET /templates` - список шаблонов;
- `GET /templates/{id}` - получить шаблон;
- `PUT /templates/{id}` - обновить шаблон;
- `DELETE /templates/{id}` - удалить шаблон.

## RabbitMQ Command Envelope

```json
{
  "type": "notification.schedule",
  "notification": {
    "channel": "email",
    "recipient": "patient@example.com",
    "template_id": "appointment_reminder",
    "variables": {
      "patient_name": "Пациент Тестовый"
    },
    "send_at": "2030-01-15T09:00:00Z"
  }
}
```

Поддерживаемые `type`: `notification.create`, `notification.schedule`, `notification.cancel`, `notification.send_now`, `template.upsert`.

## Core callback

После отправки notificator вызывает:

```text
POST http://localhost:8081/api/v1/notifications/{id}/delivery-status
```

Заголовок:

```text
X-Notifier-Secret: dev-notifier-callback-secret
```

## Demo failure

Чтобы показать retry и dead-letter без реального провайдера, укажи recipient со словом `fail`, например `fail@example.com` или `+7999fail`. Stub sender вернет ошибку, notificator выполнит retry и после лимита положит уведомление в `/dead-letters`.
