# notificator

Отдельный сервис управления уведомлениями MedSched-SaaS.

## Возможности

- создание, получение, список, отмена и ручная отправка уведомлений;
- хранение уведомлений в Redis;
- планирование отправки через RabbitMQ;
- шаблоны уведомлений с переменными вида `{{patient_name}}`;
- каналы `email`, `telegram`, `sms` со stub-отправителями;
- прием команд из очереди в формате envelope.

## HTTP

Base URL для локального конфига: `http://localhost:8082`.

- `POST /notify` - создать уведомление;
- `GET /notify` - список уведомлений;
- `GET /notify/{id}` - получить уведомление;
- `POST /notify/{id}/send` - отправить сразу;
- `DELETE /notify/{id}` - отменить;
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
