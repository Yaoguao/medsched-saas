export type UserRole =
  | 'ADMIN'
  | 'REGISTRATOR'
  | 'DOCTOR'
  | 'PATIENT'
  | 'MANAGER'
  | 'IT_SUPPORT'

export type User = {
  id: string
  patient_id?: string
  doctor_id?: string
  role: UserRole
  email: string
  phone: string
  status: string
  created_at: string
  updated_at: string
}

export type Patient = {
  id: string
  user_id?: string
  full_name: string
  gender: string
  birth_date: string
  age: number
  phone: string
  email: string
  neighbourhood: string
  scholarship: boolean
  external_ehr_id: string
  last_ehr_sync_at?: string
  medical_profile?: MedicalProfile
}

export type IntegrationEvent = {
  id: string
  provider: string
  entity_type: string
  entity_id: string
  action: string
  status: string
  message: string
  actor_id?: string
  actor_email?: string
  applied_fields: string[]
  occurred_at: string
}

export type IntegrationEventListParams = ListParams & {
  provider?: string
  status?: 'SUCCESS' | 'FAILED' | ''
  entity_type?: string
  entity_id?: string
  from?: string
  to?: string
}

export type EHRSyncResult = {
  patient: Patient
  event: IntegrationEvent
}

export type IntegrationProvider = {
  id: string
  code: string
  name: string
  type: string
  base_url: string
  status: string
  enabled: boolean
  last_healthcheck_at?: string
  last_sync_at?: string
}

export type ExternalEntityLink = {
  id: string
  provider_id: string
  entity_type: string
  internal_id: string
  external_id: string
  last_synced_at?: string
}

export type EHRFieldChange = {
  field: string
  current: unknown
  external: unknown
  changed: boolean
}

export type EHRPatientPreview = {
  provider_id: string
  provider_name: string
  provider_type: string
  external_id: string
  changes: EHRFieldChange[]
}

export type LinkPatientEHRPayload = {
  provider_id: string
  external_id: string
}

export type SyncPatientEHRPayload = {
  provider_id: string
  fields: string[]
}

export type MedicalProfile = {
  id: string
  patient_id: string
  chronic_conditions: string
  risk_factors: string
  allergies: string
  notes: string
  has_hypertension: boolean
  has_diabetes: boolean
  has_alcoholism: boolean
  has_handicap: boolean
  updated_at_profile: string
}

export type Doctor = {
  id: string
  user_id?: string
  full_name: string
  specialty: string
  cabinet: string
  experience_years: number
  status: string
}

export type Appointment = {
  id: string
  patient_id: string
  doctor_id: string
  time_slot_id: string
  scheduled_at: string
  appointment_at: string
  status: string
  booking_channel: string
  visit_type: string
  reminder_sent: boolean
  showed_up?: boolean
  patient?: Patient
  doctor?: Doctor
  time_slot?: TimeSlot
  prediction_result?: PredictionResult
  status_history?: AppointmentStatusHistory[]
}

export type AppointmentStatusHistory = {
  id: string
  appointment_id: string
  old_status?: string
  new_status: string
  changed_by?: string
  reason: string
  changed_at: string
}

export type Notification = {
  id: string
  appointment_id: string
  patient_id: string
  channel: 'SMS' | 'EMAIL' | 'PUSH'
  type: 'CONFIRMATION' | 'REMINDER' | 'CANCELLATION'
  status: 'PENDING' | 'SENT' | 'FAILED' | 'DELIVERED'
  scheduled_at: string
  sent_at?: string
  provider: string
  external_id: string
  error_message: string
  created_at: string
  updated_at: string
}

export type RescheduleProposal = {
  id: string
  appointment_id: string
  proposed_time_slot_id: string
  requested_by?: string
  responded_by?: string
  reason: string
  response_reason?: string
  status: 'PENDING' | 'ACCEPTED' | 'DECLINED'
  responded_at?: string
  appointment: Appointment
  proposed_time_slot?: TimeSlot
  created_at: string
  updated_at: string
}

export type Schedule = {
  id: string
  doctor_id: string
  start_date: string
  end_date: string
  working_days: string
  work_start_time: string
  work_end_time: string
  status: string
  time_slots?: TimeSlot[]
  created_at: string
  updated_at: string
}

export type TimeSlot = {
  id: string
  schedule_id: string
  doctor_id: string
  start_at: string
  end_at: string
  status: 'FREE' | 'BOOKED' | 'BLOCKED' | 'BREAK'
  is_blocked: boolean
  block_reason: string
  created_at: string
  updated_at: string
}

export type ListResponse<T> = {
  items: T[]
  pagination: {
    limit: number
    offset: number
  }
}

export type ListParams = {
  limit?: number
  offset?: number
}

export type DoctorListParams = ListParams & {
  search?: string
  specialty?: string
  sort?: 'name_asc' | 'name_desc' | 'experience_asc' | 'experience_desc'
  active_only?: boolean
}

export type BlockTimeSlotPayload = {
  reason: string
}

export type LoginResponse = {
  access_token: string
  token_type: string
  expires_at: string
  user: User
}

export type RegisterResponse = User

export type PredictionResult = {
  id: string
  appointment_id: string
  probability: number
  risk_level: 'LOW' | 'MEDIUM' | 'HIGH'
  model_version: string
  requested_at: string
  received_at: string
}

export type MLModelInfo = {
  loaded: boolean
  model_version: string
  algorithm: string
  artifact_path: string
  metrics: {
    trained_at?: string
    rows?: number
    roc_auc?: number
    average_precision?: number
    precision?: number
    recall?: number
    f1?: number
    decision_threshold?: number
    risk_thresholds?: {
      medium: number
      high: number
    }
    [key: string]: unknown
  }
}

export type MLTrainingJobStatus = 'QUEUED' | 'RUNNING' | 'SUCCEEDED' | 'FAILED' | 'CANCELLED'

export type MLTrainingJob = {
  id: string
  status: MLTrainingJobStatus
  source: string
  algorithm: string
  model_version: string
  core_limit: number
  requested_by?: string
  started_at?: string
  finished_at?: string
  dataset_rows: number
  error_message?: string
  created_at: string
  updated_at: string
}

export type MLModelStatus = 'CANDIDATE' | 'ACTIVE' | 'REJECTED' | 'ARCHIVED'

export type MLModelVersion = {
  id: string
  version: string
  algorithm: string
  dataset_source: 'csv' | 'core'
  artifact_uri: string
  checksum: string
  feature_schema_version: string
  dataset_rows: number
  positive_rate: number
  metrics: Record<string, unknown>
  thresholds: Record<string, unknown>
  status: MLModelStatus
  training_job_id: string
  activated_by?: string
  activated_at?: string
  created_at: string
  updated_at: string
}

export type CreateMLTrainingJobPayload = {
  source?: 'core' | 'csv'
  algorithm?: 'xgboost' | 'logistic_regression'
  core_limit?: number
}

export type ReportRange = {
  from: string
  to: string
}

export type AppointmentReport = {
  range: ReportRange
  status_counts: Record<string, number>
  total: number
  no_show: number
  cancelled: number
  completed: number
  no_show_rate: number
}

export type DoctorNoShowReportRow = {
  doctor_id: string
  doctor_name: string
  specialty: string
  total: number
  no_show: number
  cancelled: number
  completed: number
  no_show_rate: number
}

export type NoShowReport = {
  range: ReportRange
  summary: AppointmentReport
  doctors: DoctorNoShowReportRow[]
}

export type DoctorLoadReportRow = {
  doctor_id: string
  doctor_name: string
  specialty: string
  slots_total: number
  free_slots: number
  booked_slots: number
  blocked_slots: number
  load_rate: number
}

export type LoadReport = {
  range: ReportRange
  doctors: DoctorLoadReportRow[]
  slots_total: number
  booked_slots: number
  load_rate: number
}

export type DashboardReport = {
  range: ReportRange
  appointments: AppointmentReport
  no_show: NoShowReport
  load: LoadReport
}

export type ReportStatusCounts = Record<string, number>

export type CreatePatientPayload = {
  full_name: string
  gender: string
  birth_date: string
  age: number
  phone: string
  email: string
  neighbourhood: string
  scholarship: boolean
  external_ehr_id?: string
  medical_profile?: {
    chronic_conditions: string
    risk_factors: string
    allergies: string
    notes: string
    has_hypertension: boolean
    has_diabetes: boolean
    has_alcoholism: boolean
    has_handicap: boolean
  }
}

export type UpdateMedicalProfilePayload = {
  chronic_conditions?: string
  risk_factors?: string
  allergies?: string
  notes?: string
  has_hypertension?: boolean
  has_diabetes?: boolean
  has_alcoholism?: boolean
  has_handicap?: boolean
}

export type UpdateUserPayload = {
  role?: UserRole
  status?: string
  email?: string
  phone?: string
}

export type CreateStaffUserPayload = {
  role: Exclude<UserRole, 'PATIENT'>
  email: string
  phone: string
  password: string
}

export type UpdatePatientPayload = Partial<Omit<CreatePatientPayload, 'medical_profile'>>

export type LinkPatientUserPayload = {
  user_id: string
}

export type CreateDoctorPayload = {
  full_name: string
  specialty: string
  cabinet: string
  experience_years: number
  status: string
}

export type UpdateDoctorPayload = Partial<CreateDoctorPayload & { user_id: string }>

export type LinkDoctorUserPayload = {
  user_id: string
}

export type CreateSchedulePayload = {
  doctor_id: string
  start_date: string
  end_date: string
  working_days: string
  work_start_time: string
  work_end_time: string
  status: string
}

export type UpdateSchedulePayload = Partial<CreateSchedulePayload>

export type CreateTimeSlotPayload = {
  start_at: string
  end_at: string
  status: 'FREE' | 'BOOKED' | 'BLOCKED' | 'BREAK'
  is_blocked: boolean
}

export type UpdateTimeSlotPayload = Partial<CreateTimeSlotPayload>

export type ScheduleImportConflictMode = 'SKIP' | 'ABORT'

export type ScheduleImportPayload = {
  csv: string
  conflict_mode: ScheduleImportConflictMode
  timezone: string
}

export type ScheduleImportRowResult = {
  row: number
  doctor_id?: string
  doctor_name?: string
  doctor_email: string
  date: string
  slots_planned: number
  slots_created: number
  slots_skipped: number
  schedule_state?: 'CREATE' | 'EXISTING'
  status: 'VALID' | 'WARNING' | 'ERROR'
  messages: string[]
}

export type ScheduleImportResult = {
  rows: number
  valid_rows: number
  slots_planned: number
  slots_created: number
  slots_skipped: number
  schedules_planned: number
  schedules_created: number
  can_apply: boolean
  applied: boolean
  row_results: ScheduleImportRowResult[]
}

export type CreateAppointmentPayload = {
  patient_id: string
  time_slot_id: string
  booking_channel: 'WEB' | 'REGISTRY' | 'PHONE'
  visit_type: 'PRIMARY' | 'REPEAT' | 'DIAGNOSTIC' | 'CONSULTATION'
}

export type ChangeAppointmentStatusPayload = {
  changed_by?: string
  reason?: string
}

export type RescheduleAppointmentPayload = {
  time_slot_id: string
  changed_by?: string
  reason?: string
}

export type CreateRescheduleProposalPayload = {
  time_slot_id: string
  requested_by?: string
  reason?: string
}

export type RespondRescheduleProposalPayload = {
  responded_by?: string
  reason?: string
}

export type CreateNotificationPayload = {
  appointment_id: string
  patient_id: string
  channel: Notification['channel']
  type: Notification['type']
  status: Notification['status']
  scheduled_at: string
  provider?: string
  external_id?: string
  error_message?: string
}

export type CreatePredictionPayload = {
  appointment_id: string
  probability: number
  risk_level: PredictionResult['risk_level']
  model_version: string
  requested_at: string
  received_at: string
}
