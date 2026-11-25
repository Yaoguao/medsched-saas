import { apiRequest } from './http'
import type {
  Appointment,
  AppointmentReport,
  BlockTimeSlotPayload,
  ChangeAppointmentStatusPayload,
  CreateAppointmentPayload,
  CreateDoctorPayload,
  CreateNotificationPayload,
  CreateMLTrainingJobPayload,
  CreatePatientPayload,
  CreatePredictionPayload,
  CreateRescheduleProposalPayload,
  CreateSchedulePayload,
  CreateStaffUserPayload,
  CreateTimeSlotPayload,
  Doctor,
  DoctorListParams,
  DashboardReport,
  EHRSyncResult,
  EHRPatientPreview,
  ExternalEntityLink,
  IntegrationEvent,
  IntegrationEventListParams,
  IntegrationProvider,
  LinkDoctorUserPayload,
  LinkPatientUserPayload,
  LinkPatientEHRPayload,
  ListParams,
  ListResponse,
  MLModelInfo,
  MLModelVersion,
  MLTrainingJob,
  Notification,
  NoShowReport,
  Patient,
  PredictionResult,
  RespondRescheduleProposalPayload,
  RescheduleAppointmentPayload,
  RescheduleProposal,
  LoadReport,
  Schedule,
  ScheduleImportPayload,
  ScheduleImportResult,
  TimeSlot,
  SyncPatientEHRPayload,
  UpdateDoctorPayload,
  UpdateMedicalProfilePayload,
  UpdatePatientPayload,
  UpdateSchedulePayload,
  UpdateTimeSlotPayload,
  UpdateUserPayload,
  User,
} from './types'

function listQuery(params: ListParams = {}) {
  const limit = params.limit ?? 20
  const offset = params.offset ?? 0
  return `limit=${limit}&offset=${offset}`
}

export const coreApi = {
  health() {
    return apiRequest<{ status: string }>('/health', { auth: false })
  },
  users(params?: ListParams) {
    return apiRequest<ListResponse<User>>(`/users?${listQuery(params)}`)
  },
  createStaffUser(payload: CreateStaffUserPayload) {
    return apiRequest<User>('/users/staff', { method: 'POST', body: payload })
  },
  updateUser(userId: string, payload: UpdateUserPayload) {
    return apiRequest<User>(`/users/${userId}`, { method: 'PATCH', body: payload })
  },
  patients(params?: ListParams) {
    return apiRequest<ListResponse<Patient>>(`/patients?${listQuery(params)}`)
  },
  patient(patientId: string) {
    return apiRequest<Patient>(`/patients/${patientId}`)
  },
  patientAppointments(patientId: string, params?: ListParams) {
    return apiRequest<ListResponse<Appointment>>(
      `/patients/${patientId}/appointments?${listQuery(params)}`,
    )
  },
  createPatient(payload: CreatePatientPayload) {
    return apiRequest<Patient>('/patients', { method: 'POST', body: payload })
  },
  updatePatient(patientId: string, payload: UpdatePatientPayload) {
    return apiRequest<Patient>(`/patients/${patientId}`, { method: 'PATCH', body: payload })
  },
  linkPatientUser(patientId: string, payload: LinkPatientUserPayload) {
    return apiRequest<Patient>(`/patients/${patientId}/user`, {
      method: 'PATCH',
      body: payload,
    })
  },
  integrationProviders() {
    return apiRequest<IntegrationProvider[]>('/integrations/providers')
  },
  testIntegrationProvider(providerId: string) {
    return apiRequest<IntegrationProvider>(`/integrations/providers/${providerId}/test`, {
      method: 'POST',
    })
  },
  integrationEvents(params: IntegrationEventListParams = {}) {
    const query = new URLSearchParams(listQuery(params))
    if (params.provider) query.set('provider', params.provider)
    if (params.status) query.set('status', params.status)
    if (params.entity_type) query.set('entity_type', params.entity_type)
    if (params.entity_id) query.set('entity_id', params.entity_id)
    if (params.from) query.set('from', params.from)
    if (params.to) query.set('to', params.to)
    return apiRequest<ListResponse<IntegrationEvent>>(
      `/integrations/events?${query}`,
    )
  },
  retryIntegrationEvent(eventId: string) {
    return apiRequest<IntegrationEvent>(`/integrations/events/${eventId}/retry`, {
      method: 'POST',
    })
  },
  linkPatientEHR(patientId: string, payload: LinkPatientEHRPayload) {
    return apiRequest<ExternalEntityLink>(`/patients/${patientId}/ehr-link`, {
      method: 'POST',
      body: payload,
    })
  },
  previewPatientEHR(patientId: string, providerId: string) {
    const query = new URLSearchParams({ provider_id: providerId })
    return apiRequest<EHRPatientPreview>(`/patients/${patientId}/ehr-preview?${query}`)
  },
  syncPatientEHR(patientId: string, payload: SyncPatientEHRPayload) {
    return apiRequest<EHRSyncResult>(`/patients/${patientId}/ehr-sync`, {
      method: 'POST',
      body: payload,
    })
  },
  updatePatientProfile(patientId: string, payload: UpdateMedicalProfilePayload) {
    return apiRequest<unknown>(`/patients/${patientId}/profile`, {
      method: 'PATCH',
      body: payload,
    })
  },
  doctors(params: DoctorListParams = {}) {
    const query = new URLSearchParams(listQuery(params))
    if (params.search) query.set('search', params.search)
    if (params.specialty) query.set('specialty', params.specialty)
    if (params.sort) query.set('sort', params.sort)
    if (params.active_only) query.set('active_only', 'true')
    return apiRequest<ListResponse<Doctor>>(`/doctors?${query}`)
  },
  doctorSpecialties(activeOnly = false) {
    const query = new URLSearchParams()
    if (activeOnly) query.set('active_only', 'true')
    return apiRequest<string[]>(`/doctors/specialties?${query}`)
  },
  doctor(doctorId: string) {
    return apiRequest<Doctor>(`/doctors/${doctorId}`)
  },
  createDoctor(payload: CreateDoctorPayload) {
    return apiRequest<Doctor>('/doctors', { method: 'POST', body: payload })
  },
  updateDoctor(doctorId: string, payload: UpdateDoctorPayload) {
    return apiRequest<Doctor>(`/doctors/${doctorId}`, { method: 'PATCH', body: payload })
  },
  linkDoctorUser(doctorId: string, payload: LinkDoctorUserPayload) {
    return apiRequest<Doctor>(`/doctors/${doctorId}/user`, {
      method: 'PATCH',
      body: payload,
    })
  },
  createSchedule(payload: CreateSchedulePayload) {
    return apiRequest<Schedule>('/schedules', { method: 'POST', body: payload })
  },
  updateSchedule(scheduleId: string, payload: UpdateSchedulePayload) {
    return apiRequest<Schedule>(`/schedules/${scheduleId}`, {
      method: 'PATCH',
      body: payload,
    })
  },
  doctorSchedules(doctorId: string, params?: ListParams) {
    return apiRequest<ListResponse<Schedule>>(`/schedules/${doctorId}?${listQuery(params)}`)
  },
  previewScheduleImport(payload: ScheduleImportPayload) {
    return apiRequest<ScheduleImportResult>('/schedules/import/preview', {
      method: 'POST',
      body: payload,
    })
  },
  applyScheduleImport(payload: ScheduleImportPayload) {
    return apiRequest<ScheduleImportResult>('/schedules/import/apply', {
      method: 'POST',
      body: payload,
    })
  },
  createTimeSlot(scheduleId: string, payload: CreateTimeSlotPayload) {
    return apiRequest<TimeSlot>(`/schedules/${scheduleId}/slots`, {
      method: 'POST',
      body: payload,
    })
  },
  updateTimeSlot(slotId: string, payload: UpdateTimeSlotPayload) {
    return apiRequest<TimeSlot>(`/slots/${slotId}`, { method: 'PATCH', body: payload })
  },
  blockTimeSlot(slotId: string, payload: BlockTimeSlotPayload) {
    return apiRequest<TimeSlot>(`/slots/${slotId}/block`, { method: 'POST', body: payload })
  },
  unblockTimeSlot(slotId: string) {
    return apiRequest<TimeSlot>(`/slots/${slotId}/unblock`, { method: 'POST' })
  },
  appointments(params?: ListParams) {
    return apiRequest<ListResponse<Appointment>>(`/appointments?${listQuery(params)}`)
  },
  appointment(appointmentId: string) {
    return apiRequest<Appointment>(`/appointments/${appointmentId}`)
  },
  createAppointment(payload: CreateAppointmentPayload) {
    return apiRequest<Appointment>('/appointments', { method: 'POST', body: payload })
  },
  confirmAppointment(appointmentId: string, payload?: ChangeAppointmentStatusPayload) {
    return apiRequest<Appointment>(`/appointments/${appointmentId}/confirm`, {
      method: 'POST',
      body: payload,
    })
  },
  cancelAppointment(appointmentId: string, payload?: ChangeAppointmentStatusPayload) {
    return apiRequest<Appointment>(`/appointments/${appointmentId}/cancel`, {
      method: 'POST',
      body: payload,
    })
  },
  checkInAppointment(appointmentId: string, payload?: ChangeAppointmentStatusPayload) {
    return apiRequest<Appointment>(`/appointments/${appointmentId}/check-in`, {
      method: 'POST',
      body: payload,
    })
  },
  noShowAppointment(appointmentId: string, payload?: ChangeAppointmentStatusPayload) {
    return apiRequest<Appointment>(`/appointments/${appointmentId}/no-show`, {
      method: 'POST',
      body: payload,
    })
  },
  rescheduleAppointment(appointmentId: string, payload: RescheduleAppointmentPayload) {
    return apiRequest<Appointment>(`/appointments/${appointmentId}/reschedule`, {
      method: 'POST',
      body: payload,
    })
  },
  createRescheduleProposal(appointmentId: string, payload: CreateRescheduleProposalPayload) {
    return apiRequest<RescheduleProposal>(`/appointments/${appointmentId}/reschedule-proposals`, {
      method: 'POST',
      body: payload,
    })
  },
  appointmentRescheduleProposals(appointmentId: string, params?: ListParams) {
    return apiRequest<ListResponse<RescheduleProposal>>(
      `/appointments/${appointmentId}/reschedule-proposals?${listQuery(params)}`,
    )
  },
  patientRescheduleProposals(patientId: string, params?: ListParams) {
    return apiRequest<ListResponse<RescheduleProposal>>(
      `/patients/${patientId}/reschedule-proposals?${listQuery(params)}`,
    )
  },
  acceptRescheduleProposal(proposalId: string, payload: RespondRescheduleProposalPayload) {
    return apiRequest<RescheduleProposal>(`/reschedule-proposals/${proposalId}/accept`, {
      method: 'POST',
      body: payload,
    })
  },
  declineRescheduleProposal(proposalId: string, payload: RespondRescheduleProposalPayload) {
    return apiRequest<RescheduleProposal>(`/reschedule-proposals/${proposalId}/decline`, {
      method: 'POST',
      body: payload,
    })
  },
  predictAppointmentNoShow(appointmentId: string) {
    return apiRequest<PredictionResult>(
      `/appointments/${appointmentId}/predict-no-show`,
      { method: 'POST' },
    )
  },
  createPrediction(payload: CreatePredictionPayload) {
    return apiRequest<PredictionResult>('/ml/predictions', {
      method: 'POST',
      body: payload,
    })
  },
  createNotification(payload: CreateNotificationPayload) {
    return apiRequest<Notification>('/notifications/send', {
      method: 'POST',
      body: payload,
    })
  },
  appointmentNotifications(appointmentId: string, params?: ListParams) {
    return apiRequest<ListResponse<Notification>>(
      `/appointments/${appointmentId}/notifications?${listQuery(params)}`,
    )
  },
  appointmentReport(from?: string, to?: string) {
    return apiRequest<AppointmentReport>(`/reports/appointments${reportQuery(from, to)}`)
  },
  noShowReport(from?: string, to?: string) {
    return apiRequest<NoShowReport>(`/reports/no-show${reportQuery(from, to)}`)
  },
  loadReport(from?: string, to?: string) {
    return apiRequest<LoadReport>(`/reports/load${reportQuery(from, to)}`)
  },
  dashboardReport(from?: string, to?: string) {
    return apiRequest<DashboardReport>(`/reports/dashboard${reportQuery(from, to)}`)
  },
  trainingDataset() {
    return apiRequest<{ items: unknown[]; count: number }>(
      '/ml/training/no-show-dataset?limit=100',
    )
  },
  modelInfo() {
    return apiRequest<MLModelInfo>('/ml/model-info')
  },
  mlTrainingJobs(params?: ListParams) {
    return apiRequest<ListResponse<MLTrainingJob>>(`/ml/training-jobs?${listQuery(params)}`)
  },
  createMLTrainingJob(payload: CreateMLTrainingJobPayload) {
    return apiRequest<MLTrainingJob>('/ml/training-jobs', { method: 'POST', body: payload })
  },
  mlModels(params?: ListParams) {
    return apiRequest<ListResponse<MLModelVersion>>(`/ml/models?${listQuery(params)}`)
  },
  activateMLModel(version: string) {
    return apiRequest<MLModelVersion>(`/ml/models/${encodeURIComponent(version)}/activate`, {
      method: 'POST',
    })
  },
}

function reportQuery(from?: string, to?: string) {
  const params = new URLSearchParams()
  if (from) params.set('from', from)
  if (to) params.set('to', to)
  const query = params.toString()
  return query ? `?${query}` : ''
}
