# MedSched-SaaS API requests

Core service base URL: `http://localhost:8081`.

Typical manual flow:

1. Run `00 Health`.
2. Run `02 Patients - create`, copy returned `id` into `patient_id`.
3. Run `05 Patients - get medical profile` to check profile-only response.
4. Run `07 Doctors - create`, copy returned `id` into `doctor_id`.
5. Run `09 Schedules - create`, copy returned `id` into `schedule_id`.
6. Run `11 Slots - create`, copy returned `id` into `slot_id`.
7. Run `14 Appointments - create`, copy returned `id` into `appointment_id`.
8. Run status, notification, prediction and report requests.

Variables used by requests:

- `patient_id`
- `doctor_id`
- `schedule_id`
- `slot_id`
- `appointment_id`
- `notification_id`
- `prediction_id`
