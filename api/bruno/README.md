# MedSched-SaaS API requests

Core service base URL: `http://localhost:8081`.

Typical manual flow:

1. Run `00 Health`.
2. Run `01 Users - create`, copy returned `user.id` into `user_id` and `patient.id` into `patient_id`.
3. Run `02 Auth - login`, copy returned `access_token`.
4. Run `03 Users - me` with `access_token`.
5. Optionally run `07 Patients - create` only when you need an extra standalone patient.
6. Run `10 Patients - get medical profile` to check profile-only response.
7. Run `11 Patients - update medical profile` to test ML-related patient attributes.
8. Run `14 Doctors - create`, copy returned `id` into `doctor_id`.
9. Run `16 Doctors - update` to test doctor editing.
10. Run `17 Schedules - create`, copy returned `id` into `schedule_id`.
11. Run `19 Slots - create`, copy returned `id` into `slot_id`.
12. Run `22 Appointments - create`, copy returned `id` into `appointment_id`.
13. Run status, notification, prediction and report requests.
14. For notificator, run `39 Notifier - health`, then `40 Notifier - template upsert`, then `42 Notifier - create notification`, copy returned `id` into `notifier_notification_id`, and test `44/45/46`.
15. For ML service, run `50 ML - health`, then `51 ML - predict no-show`.
16. To test core-to-ML integration, start core and ML service, create an appointment, then run `52 Core - appointment predict no-show`.
17. To inspect facts for retraining, run `53 Core - ML training dataset` after some appointments are completed or marked no-show.

Variables used by requests:

- `patient_id`
- `user_id`
- `access_token`
- `doctor_id`
- `schedule_id`
- `slot_id`
- `appointment_id`
- `notification_id`
- `notifier_notification_id`
- `prediction_id`

`POST /api/v1/users` is a public registration request and always creates an active `PATIENT` user plus a linked patient row. Admin/staff role management should be tested later through a separate RBAC-protected endpoint.
