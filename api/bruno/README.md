# MedSched-SaaS API requests

Core service base URL: `http://localhost:8081`.

Typical manual flow:

1. Run `00 Health`.
2. Start core with `bootstrap.admin.enabled: true`.
3. Run `02 Auth - login admin`, copy returned `access_token` into both `access_token` and `admin_token`.
4. Run `03 Users - me` to verify the admin token.
5. Run `04 Users - create staff` to create a clean doctor user, copy returned `id` into `user_id`.
6. Run `01 Users - create`, copy returned `user.id` into `patient_user_id` if needed and `patient.id` into `patient_id`.
7. Run `03 Auth - login patient`, copy returned `access_token` into `patient_token`.
8. Put `admin_token` back into `access_token` for admin-only requests.
9. Optionally run `07 Patients - create` only when you need an extra standalone patient.
10. Run `10 Patients - get medical profile` to check profile-only response.
11. Run `11 Patients - update medical profile` to test ML-related patient attributes.
12. Run `14 Doctors - create`, copy returned `id` into `doctor_id`.
13. Run `09 Doctors - link user` to attach the staff user to the doctor profile.
14. Run `16 Doctors - update` to test doctor editing.
15. Run `17 Schedules - create`, copy returned `id` into `schedule_id`.
16. Run `19 Slots - create`, copy returned `id` into `slot_id`.
17. Put `patient_token` into `access_token` if you want to create appointment as patient.
18. Run `22 Appointments - create`, copy returned `id` into `appointment_id`.
19. Put `admin_token` back into `access_token` for status, notification, prediction and report requests.
20. For notificator, run `39 Notifier - health`, then `40 Notifier - template upsert`, then `42 Notifier - create notification`, copy returned `id` into `notifier_notification_id`, and test `44/45/46`.
21. For ML service, run `50 ML - health`, then `51 ML - predict no-show`.
22. To test core-to-ML integration, start core and ML service, create an appointment, then run `52 Core - appointment predict no-show`.
23. To inspect facts for retraining, run `53 Core - ML training dataset` after some appointments are completed or marked no-show.

Variables used by requests:

- `patient_id`
- `user_id`
- `access_token`
- `admin_token`
- `doctor_token`
- `patient_token`
- `doctor_id`
- `schedule_id`
- `slot_id`
- `appointment_id`
- `notification_id`
- `notifier_notification_id`
- `prediction_id`

`POST /api/v1/users` is a public registration request and always creates an active `PATIENT` user plus a linked patient row.

Local bootstrap admin credentials from `core/config/local.yml`:

```text
email: admin@example.com
password: admin12345
```
