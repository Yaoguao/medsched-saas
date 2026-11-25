import type { QueryKey } from '@tanstack/react-query'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Activity,
  BarChart3,
  Bell,
  Brain,
  CalendarDays,
  ClipboardPlus,
  Database,
  LogOut,
  Plug,
  Shield,
  Stethoscope,
  UserRound,
  UsersRound,
} from 'lucide-react'
import type { ReactNode } from 'react'
import { useMemo, useState } from 'react'
import { Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { ApiConsoleSection } from '../features/admin/ApiConsoleSection'
import { AppointmentsSection } from '../features/admin/AppointmentsSection'
import { DoctorsSection } from '../features/admin/DoctorsSection'
import { IntegrationsSection } from '../features/admin/IntegrationsSection'
import { MlSection } from '../features/admin/MlSection'
import { NotificationsSection } from '../features/admin/NotificationsSection'
import { OverviewSection } from '../features/admin/OverviewSection'
import { PatientsSection } from '../features/admin/PatientsSection'
import { ReportsSection } from '../features/admin/ReportsSection'
import { SchedulesSection } from '../features/admin/SchedulesSection'
import { UsersSection } from '../features/admin/UsersSection'
import type { ActionResult, AdminSection } from '../features/admin/types'
import { authApi } from '../shared/api/auth'
import { coreApi } from '../shared/api/core'
import type {
  CreateAppointmentPayload,
  CreateDoctorPayload,
  CreateNotificationPayload,
  CreatePatientPayload,
  CreateMLTrainingJobPayload,
  CreateSchedulePayload,
  CreateStaffUserPayload,
  CreateTimeSlotPayload,
  LinkDoctorUserPayload,
  LinkPatientUserPayload,
  UpdateDoctorPayload,
  UpdateMedicalProfilePayload,
  UpdatePatientPayload,
  UpdateSchedulePayload,
  UpdateTimeSlotPayload,
  UpdateUserPayload,
  UserRole,
} from '../shared/api/types'
import { clearAccessToken, getAccessToken } from '../shared/lib/auth'
import { Button } from '../shared/ui/Button'
import { OperationFeedback } from '../shared/ui/OperationFeedback'

type AdminAction = {
  title: string
  action: () => Promise<unknown>
  invalidate?: QueryKey[]
  onSuccess?: (payload: unknown) => void
}

type AdminMutationResult = ActionResult & {
  invalidate?: QueryKey[]
  onSuccess?: (payload: unknown) => void
}

type NavItem = {
  id: AdminSection
  label: string
  icon: ReactNode
}

const navItems: NavItem[] = [
  { id: 'overview', label: 'Обзор', icon: <ClipboardPlus size={18} /> },
  { id: 'users', label: 'Пользователи', icon: <Shield size={18} /> },
  { id: 'patients', label: 'Пациенты', icon: <UsersRound size={18} /> },
  { id: 'doctors', label: 'Врачи', icon: <Stethoscope size={18} /> },
  { id: 'schedules', label: 'Расписания', icon: <CalendarDays size={18} /> },
  { id: 'appointments', label: 'Записи', icon: <ClipboardPlus size={18} /> },
  { id: 'notifications', label: 'Уведомления', icon: <Bell size={18} /> },
  { id: 'integrations', label: 'Интеграции', icon: <Plug size={18} /> },
  { id: 'reports', label: 'Отчеты', icon: <BarChart3 size={18} /> },
  { id: 'ml', label: 'ML', icon: <Brain size={18} /> },
  { id: 'api', label: 'API консоль', icon: <Database size={18} /> },
]

const sectionTitles: Record<AdminSection, string> = {
  overview: 'Рабочая панель',
  users: 'Пользователи и роли',
  patients: 'Пациенты',
  doctors: 'Врачи',
  schedules: 'Расписания и слоты',
  appointments: 'Записи на прием',
  notifications: 'Уведомления',
  integrations: 'Интеграции EHR/EMR',
  reports: 'Отчеты и аналитика',
  ml: 'ML прогнозы',
  api: 'API консоль',
}

export function AdminDashboardPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const queryClient = useQueryClient()
  const token = getAccessToken()
  const [activeSection, setActiveSection] = useState<AdminSection>(() => parseAdminSection(searchParams.get('section')))
  const [selectedDoctorId, setSelectedDoctorId] = useState(searchParams.get('doctor') ?? '')
  const [selectedTimeSlotId, setSelectedTimeSlotId] = useState('')
  const [selectedBookingPatientId, setSelectedBookingPatientId] = useState(searchParams.get('patient') ?? '')
  const [selectedNotificationAppointmentId, setSelectedNotificationAppointmentId] = useState(searchParams.get('appointment') ?? '')
  const [result, setResult] = useState<ActionResult | null>(null)

  const meQuery = useQuery({
    queryKey: ['me'],
    queryFn: authApi.me,
    enabled: Boolean(token),
  })
  const role = meQuery.data?.role
  const allowedNavItems = useMemo(() => navItems.filter((item) => canOpenAdminSection(role, item.id)), [role])
  const canLoadUsers = role === 'ADMIN' || role === 'REGISTRATOR'
  const canLoadPatients = role === 'ADMIN' || role === 'REGISTRATOR'
  const canLoadDoctors = role === 'ADMIN' || role === 'REGISTRATOR' || role === 'MANAGER'
  const canLoadAppointments = role === 'ADMIN' || role === 'REGISTRATOR' || role === 'IT_SUPPORT'
  const canLoadModelInfo = role === 'ADMIN' || role === 'MANAGER' || role === 'IT_SUPPORT'

  const usersQuery = useQuery({
    queryKey: ['users'],
    queryFn: () => coreApi.users({ limit: 50 }),
    enabled: Boolean(token && canLoadUsers),
  })
  const patientsQuery = useQuery({
    queryKey: ['patients'],
    queryFn: () => coreApi.patients({ limit: 50 }),
    enabled: Boolean(token && canLoadPatients),
  })
  const doctorsQuery = useQuery({
    queryKey: ['doctors'],
    queryFn: () => coreApi.doctors({ limit: 50 }),
    enabled: Boolean(token && canLoadDoctors),
  })
  const appointmentsQuery = useQuery({
    queryKey: ['appointments'],
    queryFn: () => coreApi.appointments({ limit: 50 }),
    enabled: Boolean(token && canLoadAppointments),
  })
  const schedulesQuery = useQuery({
    queryKey: ['doctor-schedules', selectedDoctorId],
    queryFn: () => coreApi.doctorSchedules(selectedDoctorId, { limit: 50 }),
    enabled: Boolean(token && selectedDoctorId),
  })
  const mlModelInfoQuery = useQuery({
    queryKey: ['ml-model-info'],
    queryFn: coreApi.modelInfo,
    enabled: Boolean(token && canLoadModelInfo),
  })
  const mlTrainingJobsQuery = useQuery({
    queryKey: ['ml-training-jobs'],
    queryFn: () => coreApi.mlTrainingJobs({ limit: 50 }),
    enabled: Boolean(token && canLoadModelInfo && activeSection === 'ml'),
    refetchInterval: (query) => {
      const jobs = query.state.data?.items ?? []
      return jobs.some((job) => job.status === 'QUEUED' || job.status === 'RUNNING')
        ? 3000
        : false
    },
  })
  const mlModelsQuery = useQuery({
    queryKey: ['ml-model-versions'],
    queryFn: () => coreApi.mlModels({ limit: 50 }),
    enabled: Boolean(token && canLoadModelInfo && activeSection === 'ml'),
    refetchInterval: () => {
      const jobs = mlTrainingJobsQuery.data?.items ?? []
      return jobs.some((job) => job.status === 'QUEUED' || job.status === 'RUNNING')
        ? 3000
        : false
    },
  })

  const actionMutation = useMutation<AdminMutationResult, Error, AdminAction>({
    mutationFn: async ({ title, action, invalidate, onSuccess }) => ({
      title,
      payload: await action(),
      invalidate,
      onSuccess,
    }),
    onSuccess: ({ title, payload, invalidate, onSuccess }) => {
      setResult({ title, payload })
      invalidate?.forEach((queryKey) => {
        queryClient.invalidateQueries({ queryKey })
      })
      onSuccess?.(payload)
    },
    onError: (error) => {
      setResult({ title: 'Ошибка запроса', payload: { message: error.message } })
    },
  })

  if (!token) {
    return <Navigate to="/login" replace />
  }
  if (role === 'PATIENT' || role === 'DOCTOR') {
    return <Navigate to="/home" replace />
  }

  function runAction(
    title: string,
    action: () => Promise<unknown>,
    invalidate: QueryKey[] = [],
    onSuccess?: (payload: unknown) => void,
  ) {
    setResult(null)
    actionMutation.mutate({ title, action, invalidate, onSuccess })
  }

  function selectDoctor(doctorId: string) {
    setSelectedDoctorId(doctorId)
    setSelectedTimeSlotId('')
  }

  function openSection(section: AdminSection, params?: Record<string, string>) {
    setResult(null)
    setActiveSection(section)
    setSearchParams({ section, ...params })
  }

  function useTimeSlotForAppointment(slotId: string) {
    setSelectedTimeSlotId(slotId)
    openSection('appointments')
  }

  const users = usersQuery.data?.items ?? []
  const patients = patientsQuery.data?.items ?? []
  const doctors = doctorsQuery.data?.items ?? []
  const appointments = appointmentsQuery.data?.items ?? []
  const currentSection = role && !canOpenAdminSection(role, activeSection)
    ? defaultAdminSection(role)
    : activeSection
  const roleLabel = workspaceLabel(role)

  return (
    <main className="dashboard-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-mark small">
            <Activity size={19} />
          </div>
          <div>
            <strong>MedSched Workspace</strong>
            <span>{roleLabel}</span>
          </div>
        </div>
        <nav className="sidebar-nav" aria-label="Навигация рабочей панели">
          {allowedNavItems.map((item) => (
            <button
              key={item.id}
              type="button"
              className={item.id === currentSection ? 'active' : undefined}
              onClick={() => openSection(item.id)}
            >
              {item.icon}
              {item.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-user">
          <div className="sidebar-avatar">
            {userInitials(meQuery.data?.email)}
          </div>
          <div>
            <strong>{roleLabel}</strong>
            <span>{meQuery.data?.email ?? 'Пользователь'}</span>
          </div>
        </div>
      </aside>

      <section className="dashboard-main">
        <header className="dashboard-header">
          <div>
            <p className="eyebrow">{roleLabel}</p>
            <h1>{sectionTitles[currentSection]}</h1>
            <span>{sectionDescription(currentSection)}</span>
          </div>
          <div className="workspace-actions">
            <Button
              variant="secondary"
              icon={<UserRound size={16} />}
              onClick={() => navigate('/profile')}
            >
              Профиль
            </Button>
            <Button
              variant="ghost"
              icon={<LogOut size={16} />}
              onClick={() => {
                clearAccessToken()
                navigate('/login')
              }}
            >
              Выйти
            </Button>
          </div>
        </header>

        {!['patients', 'doctors', 'schedules'].includes(currentSection) ? (
          <OperationFeedback result={result} onDismiss={() => setResult(null)} />
        ) : null}

        {currentSection === 'overview' ? (
          <OverviewSection
            role={role}
            appointments={appointments}
            patients={patients}
            doctors={doctors}
          />
        ) : null}

        {currentSection === 'users' ? (
          <UsersSection
            users={users}
            isPending={actionMutation.isPending}
            onUpdateUser={(userId: string, payload: UpdateUserPayload) =>
              runAction('User обновлен', () => coreApi.updateUser(userId, payload), [
                ['users'],
                ['me'],
              ])
            }
          />
        ) : null}

        {currentSection === 'patients' ? (
          <PatientsSection
            patients={patients}
            users={users}
            initialPatientId={searchParams.get('patient') ?? ''}
            actionResult={result}
            onDismissResult={() => setResult(null)}
            isPending={actionMutation.isPending}
            onCreatePatient={(payload: CreatePatientPayload) =>
              runAction('Пациент создан', () => coreApi.createPatient(payload), [['patients']])
            }
            onUpdatePatient={(patientId: string, payload: UpdatePatientPayload) =>
              runAction('Пациент обновлен', () => coreApi.updatePatient(patientId, payload), [
                ['patients'],
                ['patient', patientId],
              ])
            }
            onUpdateProfile={(patientId: string, payload: UpdateMedicalProfilePayload) =>
              runAction(
                'Медицинский профиль обновлен',
                () => coreApi.updatePatientProfile(patientId, payload),
                [
                  ['patients'],
                  ['patient', patientId],
                ],
              )
            }
            onLinkPatientUser={(patientId: string, payload: LinkPatientUserPayload) =>
              runAction('Аккаунт привязан к пациенту', () => coreApi.linkPatientUser(patientId, payload), [
                ['patients'],
                ['patient', patientId],
                ['users'],
              ])
            }
            onOpenIntegrations={(patientId: string) =>
              openSection('integrations', { patient: patientId })
            }
            onBookPatient={(patientId: string) => {
              setSelectedBookingPatientId(patientId)
              openSection('appointments', { patient: patientId })
            }}
          />
        ) : null}

        {currentSection === 'doctors' ? (
          <DoctorsSection
            doctors={doctors}
            users={users}
            initialDoctorId={searchParams.get('doctor') ?? ''}
            actionResult={result}
            onDismissResult={() => setResult(null)}
            isPending={actionMutation.isPending}
            canManageStaff={role === 'ADMIN'}
            onCreateStaffUser={(payload: CreateStaffUserPayload) =>
              runAction('Аккаунт сотрудника создан', () => coreApi.createStaffUser(payload), [['users']])
            }
            onCreateDoctor={(payload: CreateDoctorPayload) =>
              runAction('Врач создан', () => coreApi.createDoctor(payload), [['doctors']])
            }
            onUpdateDoctor={(doctorId: string, payload: UpdateDoctorPayload) =>
              runAction('Врач обновлен', () => coreApi.updateDoctor(doctorId, payload), [
                ['doctors'],
                ['doctor', doctorId],
              ])
            }
            onLinkDoctorUser={(doctorId: string, payload: LinkDoctorUserPayload) =>
              runAction('Аккаунт привязан к врачу', () => coreApi.linkDoctorUser(doctorId, payload), [
                ['doctors'],
                ['doctor', doctorId],
                ['users'],
                ['me'],
              ])
            }
            onOpenSchedule={(doctorId: string) => {
              selectDoctor(doctorId)
              openSection('schedules', { doctor: doctorId })
            }}
          />
        ) : null}

        {currentSection === 'schedules' ? (
          <SchedulesSection
            doctors={doctors}
            selectedDoctorId={selectedDoctorId}
            isPending={actionMutation.isPending}
            schedulesQuery={schedulesQuery}
            onSelectDoctor={selectDoctor}
            onCreateSchedule={(payload: CreateSchedulePayload) =>
              runAction('Расписание создано', () => coreApi.createSchedule(payload), [
                ['doctor-schedules', payload.doctor_id],
              ])
            }
            onUpdateSchedule={(scheduleId: string, payload: UpdateSchedulePayload) =>
              runAction('Расписание обновлено', () => coreApi.updateSchedule(scheduleId, payload), [
                ['doctor-schedules', selectedDoctorId],
              ])
            }
            onCreateTimeSlot={(scheduleId: string, payload: CreateTimeSlotPayload) =>
              runAction('Слот создан', () => coreApi.createTimeSlot(scheduleId, payload), [
                ['doctor-schedules', selectedDoctorId],
              ])
            }
            onUpdateTimeSlot={(slotId: string, payload: UpdateTimeSlotPayload) =>
              runAction('Слот обновлен', () => coreApi.updateTimeSlot(slotId, payload), [
                ['doctor-schedules', selectedDoctorId],
              ])
            }
            onBlockTimeSlot={(slotId: string, reason: string) =>
              runAction('Слот заблокирован', () => coreApi.blockTimeSlot(slotId, { reason }), [
                ['doctor-schedules', selectedDoctorId],
              ])
            }
            onUnblockTimeSlot={(slotId: string) =>
              runAction('Слот разблокирован', () => coreApi.unblockTimeSlot(slotId), [
                ['doctor-schedules', selectedDoctorId],
              ])
            }
            onUseTimeSlot={useTimeSlotForAppointment}
            canUseTimeSlot={canOpenAdminSection(role, 'appointments')}
            actionResult={result}
            onDismissResult={() => setResult(null)}
          />
        ) : null}

        {currentSection === 'appointments' ? (
          <AppointmentsSection
            role={role}
            preselectedPatientId={selectedBookingPatientId}
            appointments={appointments}
            patients={patients}
            doctors={doctors}
            schedules={schedulesQuery.data?.items ?? []}
            selectedDoctorId={selectedDoctorId}
            selectedTimeSlotId={selectedTimeSlotId}
            isPending={actionMutation.isPending}
            isSchedulesLoading={schedulesQuery.isFetching}
            onSelectDoctor={selectDoctor}
            onSelectTimeSlot={setSelectedTimeSlotId}
            onCreateAppointment={(payload: CreateAppointmentPayload) =>
              runAction(
                'Запись создана',
                async () => {
                  const appointment = await coreApi.createAppointment(payload)
                  setSelectedTimeSlotId('')
                  return appointment
                },
                [
                  ['appointments'],
                  ['doctor-schedules', selectedDoctorId],
                ],
                (created) => {
                  const appointment = created as { id?: string }
                  if (appointment.id) {
                    navigate(`/workspace/appointments/${appointment.id}`)
                  }
                },
              )
            }
            onPredict={(appointmentId: string) =>
              runAction('ML-прогноз по записи', () => coreApi.predictAppointmentNoShow(appointmentId))
            }
          />
        ) : null}

        {currentSection === 'notifications' ? (
          <NotificationsSection
            appointments={appointments}
            appointmentId={selectedNotificationAppointmentId}
            isPending={actionMutation.isPending}
            onSelectAppointment={(appointmentId) => {
              setSelectedNotificationAppointmentId(appointmentId)
              openSection('notifications', appointmentId ? { appointment: appointmentId } : undefined)
            }}
            onCreateNotification={(payload: CreateNotificationPayload) =>
              runAction('Уведомление создано', () => coreApi.createNotification(payload), [
                ['appointment-notifications', payload.appointment_id],
              ])
            }
          />
        ) : null}

        {currentSection === 'integrations' ? (
          <IntegrationsSection
            patients={patients}
            initialPatientId={searchParams.get('patient') ?? ''}
            canTestProvider={role === 'ADMIN' || role === 'IT_SUPPORT'}
            canManageLinks={role === 'ADMIN' || role === 'REGISTRATOR'}
            canRetryEvents={role === 'ADMIN' || role === 'REGISTRATOR'}
          />
        ) : null}

        {currentSection === 'reports' ? <ReportsSection /> : null}

        {currentSection === 'ml' ? (
          <MlSection
            appointments={appointments}
            isPending={actionMutation.isPending}
            modelInfo={mlModelInfoQuery.data}
            isModelInfoLoading={mlModelInfoQuery.isFetching}
            trainingJobs={mlTrainingJobsQuery.data?.items ?? []}
            models={mlModelsQuery.data?.items ?? []}
            isRegistryLoading={mlTrainingJobsQuery.isFetching || mlModelsQuery.isFetching}
            canManage={role === 'ADMIN' || role === 'IT_SUPPORT'}
            onPredict={(appointmentId: string) =>
              runAction('ML-прогноз по записи', () => coreApi.predictAppointmentNoShow(appointmentId))
            }
            onStartTraining={(payload: CreateMLTrainingJobPayload) =>
              runAction('Обучение поставлено в очередь', () => coreApi.createMLTrainingJob(payload), [
                ['ml-training-jobs'],
              ])
            }
            onActivateModel={(version: string) =>
              runAction('Модель активирована', () => coreApi.activateMLModel(version), [
                ['ml-model-versions'],
                ['ml-model-info'],
              ])
            }
          />
        ) : null}

        {currentSection === 'api' ? (
          <ApiConsoleSection
            result={result}
            isPending={actionMutation.isPending}
            canCheckUsers={role === 'ADMIN'}
            canCheckPatients={role === 'ADMIN' || role === 'REGISTRATOR'}
            canCheckAppointments={role === 'ADMIN' || role === 'REGISTRATOR' || role === 'IT_SUPPORT'}
            canCheckTrainingDataset={role === 'ADMIN' || role === 'MANAGER' || role === 'IT_SUPPORT'}
            onHealth={() => runAction('Проверка сервиса', coreApi.health)}
            onUsers={() => runAction('Пользователи', coreApi.users, [['users']])}
            onPatients={() => runAction('Пациенты', coreApi.patients, [['patients']])}
            onDoctors={() => runAction('Врачи', coreApi.doctors, [['doctors']])}
            onAppointments={() => runAction('Записи', coreApi.appointments, [['appointments']])}
            onTrainingDataset={() => runAction('Обучающая выборка', coreApi.trainingDataset)}
          />
        ) : null}
      </section>
    </main>
  )
}

function canOpenAdminSection(role: UserRole | undefined, section: AdminSection) {
  if (!role) {
    return section === 'overview'
  }
  if (role === 'ADMIN') {
    return true
  }
  if (role === 'REGISTRATOR') {
    return ['overview', 'patients', 'appointments', 'notifications', 'integrations'].includes(section)
  }
  if (role === 'MANAGER') {
    return ['overview', 'schedules', 'reports', 'integrations'].includes(section)
  }
  if (role === 'IT_SUPPORT') {
    return ['overview', 'notifications', 'integrations', 'ml', 'api'].includes(section)
  }
  return false
}

function parseAdminSection(value: string | null): AdminSection {
  const sections: AdminSection[] = [
    'overview',
    'users',
    'patients',
    'doctors',
    'schedules',
    'appointments',
    'notifications',
    'integrations',
    'reports',
    'ml',
    'api',
  ]
  return value && sections.includes(value as AdminSection) ? value as AdminSection : 'overview'
}

function defaultAdminSection(role: UserRole): AdminSection {
  if (role === 'MANAGER') {
    return 'reports'
  }
  return 'overview'
}

function workspaceLabel(role: UserRole | undefined) {
  switch (role) {
    case 'ADMIN':
      return 'Панель администратора'
    case 'REGISTRATOR':
      return 'Рабочее место регистратора'
    case 'MANAGER':
      return 'Панель менеджера'
    case 'IT_SUPPORT':
      return 'Техническая панель'
    default:
      return 'Рабочая панель'
  }
}

function sectionDescription(section: AdminSection) {
  switch (section) {
    case 'overview':
      return 'Сводка по операционной работе клиники и ближайшим приемам.'
    case 'users':
      return 'Роли, статусы и привязки аккаунтов к профилям.'
    case 'patients':
      return 'Медицинские карты, признаки риска и интеграция с EHR/EMR.'
    case 'doctors':
      return 'Профили врачей, специальности, кабинеты и связанные аккаунты.'
    case 'schedules':
      return 'Расписания врачей, слоты приема и блокировки времени.'
    case 'appointments':
      return 'Записи пациентов, статусы, действия и ML-риск неявки.'
    case 'notifications':
      return 'Уведомления по приемам, каналы и статусы доставки.'
    case 'integrations':
      return 'Подключения к внешним медицинским системам, привязки пациентов и журнал обмена.'
    case 'reports':
      return 'Аналитика по записям, неявкам и загрузке слотов.'
    case 'ml':
      return 'Модель риска неявки, прогнозы и обучающая выборка.'
    case 'api':
      return 'Быстрая проверка сервисных endpoint и диагностических ответов.'
  }
}

function userInitials(email?: string) {
  if (!email) {
    return 'MS'
  }
  const name = email.split('@')[0] || email
  return name
    .split(/[._-]/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('') || 'MS'
}
