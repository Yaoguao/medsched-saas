import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Activity,
  ArrowRight,
  AlertTriangle,
  BarChart3,
  Bell,
  Brain,
  CalendarDays,
  CheckCircle2,
  ClipboardPlus,
  Database,
  HeartPulse,
  LogOut,
  Search,
  Shield,
  Stethoscope,
  UserRound,
  UsersRound,
} from 'lucide-react'
import { lazy, Suspense, type ReactNode } from 'react'
import { useCallback, useDeferredValue, useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { authApi } from '../shared/api/auth'
import { coreApi } from '../shared/api/core'
import type { Appointment, RescheduleProposal, TimeSlot, User, UserRole } from '../shared/api/types'
import { clearAccessToken, getAccessToken } from '../shared/lib/auth'
import { appointmentStatusLabel, riskLevelLabel, userStatusLabel, visitTypeLabel } from '../shared/lib/labels'
import { isBookableTimeSlot } from '../shared/lib/timeSlots'
import { Button } from '../shared/ui/Button'
import { ContentState } from '../shared/ui/ContentState'
import { MetricCard } from '../shared/ui/MetricCard'

type RoleAction = {
  title: string
  text: string
  icon: ReactNode
  to?: string
}

const DoctorCalendar = lazy(() =>
  import('../features/calendar/DoctorCalendar').then((module) => ({ default: module.DoctorCalendar })),
)

const roleTitles: Record<UserRole, string> = {
  ADMIN: 'Администратор',
  REGISTRATOR: 'Регистратор',
  DOCTOR: 'Врач',
  PATIENT: 'Пациент',
  MANAGER: 'Менеджер',
  IT_SUPPORT: 'Техподдержка',
}

const roleDescriptions: Record<UserRole, string> = {
  ADMIN: 'Полный доступ к пользователям, врачам, пациентам, расписаниям, записям и аналитике.',
  REGISTRATOR: 'Работа с пациентами, слотами и записями на прием.',
  DOCTOR: 'Просмотр своего профиля, записей и клинических инструментов.',
  PATIENT: 'Личный кабинет пациента с профилем и записями.',
  MANAGER: 'Операционная аналитика, отчеты и контроль загрузки.',
  IT_SUPPORT: 'Контроль интеграций, уведомлений, ML-сервиса и сервисных данных.',
}

export function HomePage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const token = getAccessToken()
  const [patientDoctorId, setPatientDoctorId] = useState('')
  const [patientDoctorSearch, setPatientDoctorSearch] = useState('')
  const [loadedAt] = useState(() => Date.now())
  const deferredPatientDoctorSearch = useDeferredValue(patientDoctorSearch.trim())

  const meQuery = useQuery({
    queryKey: ['me'],
    queryFn: authApi.me,
    enabled: Boolean(token),
  })
  const patientsQuery = useQuery({
    queryKey: ['home-patients'],
    queryFn: () => coreApi.patients({ limit: 20 }),
    enabled: Boolean(token && canSeeOperations(meQuery.data?.role)),
  })
  const doctorsQuery = useQuery({
    queryKey: ['home-doctors', meQuery.data?.role === 'PATIENT' ? deferredPatientDoctorSearch : 'operations'],
    queryFn: () => coreApi.doctors({
      limit: meQuery.data?.role === 'PATIENT' ? 8 : 20,
      search: meQuery.data?.role === 'PATIENT' ? deferredPatientDoctorSearch : '',
      sort: 'name_asc',
      active_only: meQuery.data?.role === 'PATIENT',
    }),
    enabled: Boolean(token && (canSeeOperations(meQuery.data?.role) || meQuery.data?.role === 'PATIENT')),
  })
  const appointmentsQuery = useQuery({
    queryKey: ['home-appointments'],
    queryFn: () => coreApi.appointments({ limit: 50 }),
    enabled: Boolean(token && (canSeeOperations(meQuery.data?.role) || meQuery.data?.role === 'DOCTOR')),
  })
  const patientQuery = useQuery({
    queryKey: ['home-patient', meQuery.data?.patient_id],
    queryFn: () => coreApi.patient(meQuery.data?.patient_id ?? ''),
    enabled: Boolean(token && meQuery.data?.patient_id),
  })
  const doctorQuery = useQuery({
    queryKey: ['home-doctor', meQuery.data?.doctor_id],
    queryFn: () => coreApi.doctor(meQuery.data?.doctor_id ?? ''),
    enabled: Boolean(token && meQuery.data?.doctor_id),
  })
  const patientAppointmentsQuery = useQuery({
    queryKey: ['home-patient-appointments', meQuery.data?.patient_id],
    queryFn: () => coreApi.patientAppointments(meQuery.data?.patient_id ?? '', { limit: 10 }),
    enabled: Boolean(token && meQuery.data?.patient_id),
  })
  const patientRescheduleProposalsQuery = useQuery({
    queryKey: ['home-patient-reschedule-proposals', meQuery.data?.patient_id],
    queryFn: () => coreApi.patientRescheduleProposals(meQuery.data?.patient_id ?? '', { limit: 20 }),
    enabled: Boolean(token && meQuery.data?.role === 'PATIENT' && meQuery.data?.patient_id),
  })
  const patientBookingSchedulesQuery = useQuery({
    queryKey: ['patient-booking-schedules', patientDoctorId],
    queryFn: () => coreApi.doctorSchedules(patientDoctorId, { limit: 50 }),
    enabled: Boolean(token && meQuery.data?.role === 'PATIENT' && patientDoctorId),
  })
  const patientActionMutation = useMutation({
    mutationFn: async (action: () => Promise<unknown>) => action(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['home-patient-appointments', meQuery.data?.patient_id] })
      queryClient.invalidateQueries({ queryKey: ['home-patient-reschedule-proposals', meQuery.data?.patient_id] })
      queryClient.invalidateQueries({ queryKey: ['patient-booking-schedules', patientDoctorId] })
    },
  })

  if (!token) {
    return <Navigate to="/login" replace />
  }

  const user = meQuery.data
  const role = user?.role ?? 'PATIENT'
  const actions = roleActions(role)
  const operationalAppointments = appointmentsQuery.data?.items ?? []
  const personalAppointments = role === 'PATIENT'
    ? patientAppointmentsQuery.data?.items ?? []
    : operationalAppointments.filter((appointment) => appointment.doctor_id === user?.doctor_id)

  return (
    <main className="home-shell">
      <header className="home-header">
        <div className="home-brand">
          <div className="brand-mark small">
            <Activity size={20} />
          </div>
          <div>
            <strong>MedSched</strong>
            <span>Медицинская информационная система</span>
          </div>
        </div>
        <div className="home-user-actions">
          <div className="home-user-card">
            <span>{roleTitles[role]}</span>
            <strong>{user?.email ?? 'Пользователь'}</strong>
          </div>
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

      <section className="home-hero">
        <div>
          <p className="eyebrow">{roleTitles[role]}</p>
          <h1>{homeTitle(role, user, patientQuery.data?.full_name, doctorQuery.data?.full_name)}</h1>
          <p className="lead">{roleDescriptions[role]}</p>
        </div>
        <div className="home-status">
          <span>Статус профиля</span>
          <strong>{userStatusLabel(user?.status ?? 'ACTIVE')}</strong>
          <small>{roleTitles[role]}</small>
        </div>
      </section>

      {role === 'DOCTOR' ? (
        <DoctorMetrics appointments={personalAppointments} />
      ) : role === 'PATIENT' ? (
        <PatientMetrics appointments={personalAppointments} />
      ) : (
        <section className="metrics-grid">
          <MetricCard
            label="Пациенты"
            value={patientsQuery.data?.items.length ?? (patientQuery.data ? 1 : 0)}
            hint={canSeeOperations(role) ? 'доступно в панели' : 'ваш профиль'}
            tone="green"
            icon={<UsersRound size={22} />}
          />
          <MetricCard
            label="Врачи"
            value={doctorsQuery.data?.items.length ?? (doctorQuery.data ? 1 : 0)}
            hint={canSeeOperations(role) ? 'активный справочник' : 'связанный профиль'}
            tone="violet"
            icon={<Stethoscope size={22} />}
          />
          <MetricCard
            label="Записи"
            value={canSeeOperations(role) ? operationalAppointments.length : personalAppointments.length}
            hint="последние записи"
            tone="blue"
            icon={<CalendarDays size={22} />}
          />
          <MetricCard
            label="Инструменты"
            value={actions.length}
            hint="доступно по роли"
            tone="orange"
            icon={<Shield size={22} />}
          />
        </section>
      )}

      {role === 'DOCTOR' ? (
        <DoctorDashboard
          appointments={personalAppointments}
          doctorName={doctorQuery.data?.full_name}
        />
      ) : role === 'PATIENT' ? (
        <PatientDashboard
          patientId={user?.patient_id}
          patientName={patientQuery.data?.full_name}
          appointments={personalAppointments}
          proposals={patientRescheduleProposalsQuery.data?.items ?? []}
          doctors={doctorsQuery.data?.items ?? []}
          selectedDoctorId={patientDoctorId}
          doctorSearch={patientDoctorSearch}
          freeSlots={(patientBookingSchedulesQuery.data?.items ?? [])
            .flatMap((schedule) => schedule.time_slots ?? [])
            .filter((slot) => isBookableTimeSlot(slot, loadedAt))}
          isPending={patientActionMutation.isPending}
          onSelectDoctor={setPatientDoctorId}
          onSearchDoctor={(value) => {
            setPatientDoctorSearch(value)
            setPatientDoctorId('')
          }}
          onCreateAppointment={(slotId, visitType) => {
            if (!user?.patient_id) {
              return
            }
            patientActionMutation.mutate(() =>
              coreApi.createAppointment({
                patient_id: user.patient_id ?? '',
                time_slot_id: slotId,
                booking_channel: 'WEB',
                visit_type: visitType,
              }),
            )
          }}
        />
      ) : (
        <OperationalDashboard
          role={role}
          actions={actions}
          appointments={canSeeOperations(role) ? operationalAppointments : personalAppointments}
        />
      )}
    </main>
  )
}

function canSeeOperations(role?: UserRole) {
  return role === 'ADMIN' || role === 'REGISTRATOR' || role === 'MANAGER' || role === 'IT_SUPPORT'
}

function homeTitle(role: UserRole, user?: User, patientName?: string, doctorName?: string) {
  if (!user) {
    return 'Загрузка кабинета'
  }
  if (role === 'DOCTOR') {
    return doctorName ? `Добро пожаловать, ${doctorName}` : 'Рабочий кабинет врача'
  }
  if (role === 'PATIENT') {
    return patientName ? `Личный кабинет: ${patientName}` : 'Личный кабинет пациента'
  }
  return workspaceHomeTitle(role)
}

function workspaceHomeTitle(role: UserRole) {
  switch (role) {
    case 'ADMIN':
      return 'Панель администратора'
    case 'REGISTRATOR':
      return 'Рабочее место регистратора'
    case 'MANAGER':
      return 'Операционная панель менеджера'
    case 'IT_SUPPORT':
      return 'Панель технической поддержки'
    default:
      return 'Рабочий кабинет'
  }
}

function appointmentsTitle(role: UserRole) {
  if (role === 'DOCTOR') return 'Ваши ближайшие записи'
  if (role === 'PATIENT') return 'Ваши приемы'
  return 'Последние записи'
}

function roleActions(role: UserRole): RoleAction[] {
  const adminPanel = { title: 'Рабочая панель', text: 'Полное управление данными', icon: <Shield size={22} />, to: '/workspace' }

  if (role === 'ADMIN') {
    return [
      adminPanel,
      { title: 'Пользователи', text: 'Роли, статусы, привязки', icon: <UsersRound size={22} />, to: '/workspace' },
      { title: 'Отчеты', text: 'неявки и статусы записей', icon: <BarChart3 size={22} />, to: '/workspace' },
      { title: 'ML', text: 'Прогнозы и обучающая выборка', icon: <Brain size={22} />, to: '/workspace' },
    ]
  }
  if (role === 'REGISTRATOR') {
    return [
      adminPanel,
      { title: 'Пациенты', text: 'Создание и обновление карт', icon: <HeartPulse size={22} />, to: '/workspace' },
      { title: 'Записи', text: 'Подбор слотов и запись на прием', icon: <ClipboardPlus size={22} />, to: '/workspace' },
      { title: 'Уведомления', text: 'Напоминания по записям', icon: <Bell size={22} />, to: '/workspace' },
    ]
  }
  if (role === 'DOCTOR') {
    return [
      { title: 'Расписание', text: 'Смотрите ближайшие записи', icon: <CalendarDays size={22} /> },
      { title: 'Пациенты', text: 'Контекст приема и медицинские признаки', icon: <HeartPulse size={22} /> },
      { title: 'ML риск', text: 'Прогноз неявки по записи', icon: <Brain size={22} /> },
    ]
  }
  if (role === 'MANAGER') {
    return [
      adminPanel,
      { title: 'Отчеты', text: 'Статусы и доля неявок', icon: <BarChart3 size={22} />, to: '/workspace' },
      { title: 'Загрузка', text: 'Врачи, слоты, записи', icon: <CalendarDays size={22} />, to: '/workspace' },
    ]
  }
  if (role === 'IT_SUPPORT') {
    return [
      adminPanel,
      { title: 'API консоль', text: 'Проверка сервисных запросов', icon: <Database size={22} />, to: '/workspace' },
      { title: 'Интеграции', text: 'ML и сервис уведомлений', icon: <Activity size={22} />, to: '/workspace' },
    ]
  }
  return [
    { title: 'Мой профиль', text: 'Пациентские данные и контакты', icon: <HeartPulse size={22} /> },
    { title: 'Мои записи', text: 'Предстоящие приемы', icon: <CalendarDays size={22} /> },
    { title: 'Уведомления', text: 'Напоминания по приемам', icon: <Bell size={22} /> },
  ]
}

function AppointmentPreview({ appointments }: { appointments: Appointment[] }) {
  if (!appointments.length) {
    return <p className="empty-state">Записей пока нет.</p>
  }

  return (
    <div className="appointment-list">
      {appointments.slice(0, 6).map((appointment) => (
        <div className="appointment-row" key={appointment.id}>
          <span>{new Date(appointment.appointment_at).toLocaleString('ru-RU')}</span>
          <strong>{appointmentStatusLabel(appointment.status)}</strong>
          <small>{visitTypeLabel(appointment.visit_type)}</small>
        </div>
      ))}
    </div>
  )
}

function OperationalDashboard({
  role,
  actions,
  appointments,
}: {
  role: UserRole
  actions: RoleAction[]
  appointments: Appointment[]
}) {
  const focusItems = roleFocusItems(role)
  return (
    <section className="home-grid">
      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Быстрый доступ</p>
            <h2>Рабочие действия</h2>
          </div>
        </div>
        <div className="home-actions">
          {actions.map((action) => (
            action.to ? (
              <Link className="home-action" key={action.title} to={action.to}>
                {action.icon}
                <strong>{action.title}</strong>
                <span>{action.text}</span>
              </Link>
            ) : (
              <div className="home-action" key={action.title}>
                {action.icon}
                <strong>{action.title}</strong>
                <span>{action.text}</span>
              </div>
            )
          ))}
        </div>
      </article>

      <article className="panel role-focus-panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">{roleTitles[role]}</p>
            <h2>Фокус рабочего дня</h2>
          </div>
        </div>
        <div className="role-focus-list">
          {focusItems.map((item) => (
            <div key={item.title}>
              {item.icon}
              <strong>{item.title}</strong>
              <span>{item.text}</span>
            </div>
          ))}
        </div>
      </article>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Записи</p>
            <h2>{appointmentsTitle(role)}</h2>
          </div>
        </div>
        <AppointmentPreview appointments={appointments} />
      </article>
    </section>
  )
}

function roleFocusItems(role: UserRole): RoleAction[] {
  if (role === 'ADMIN') {
    return [
      { title: 'Контроль доступа', text: 'Пользователи, роли, блокировки и привязки аккаунтов.', icon: <Shield size={20} /> },
      { title: 'Операционные данные', text: 'Пациенты, врачи, расписания и записи в единой панели.', icon: <Database size={20} /> },
      { title: 'Аналитика и ML', text: 'Отчеты, прогнозы риска и обучающая выборка.', icon: <Brain size={20} /> },
    ]
  }
  if (role === 'REGISTRATOR') {
    return [
      { title: 'Поиск пациента', text: 'Быстро найти карту, проверить контакты и аккаунт.', icon: <Search size={20} /> },
      { title: 'Запись на прием', text: 'Выбрать врача, свободный слот и тип визита.', icon: <ClipboardPlus size={20} /> },
      { title: 'Статусы приема', text: 'Подтверждение, явка, неявка и отмена записи.', icon: <CheckCircle2 size={20} /> },
    ]
  }
  if (role === 'MANAGER') {
    return [
      { title: 'Расписания', text: 'Планирование рабочих дней врачей и слотов.', icon: <CalendarDays size={20} /> },
      { title: 'Загрузка', text: 'Свободные и занятые слоты, эффективность расписаний.', icon: <BarChart3 size={20} /> },
      { title: 'Неявки', text: 'Контроль no-show и операционных потерь.', icon: <AlertTriangle size={20} /> },
    ]
  }
  if (role === 'IT_SUPPORT') {
    return [
      { title: 'Core API', text: 'Проверка доступности и сервисных ответов.', icon: <Activity size={20} /> },
      { title: 'ML Service', text: 'Состояние модели, прогнозы и данные обучения.', icon: <Brain size={20} /> },
      { title: 'Уведомления', text: 'Ошибки доставки и статусы внешнего сервиса.', icon: <Bell size={20} /> },
    ]
  }
  return []
}

function DoctorMetrics({ appointments }: { appointments: Appointment[] }) {
  const todayCount = appointments.filter(isTodayAppointment).length
  const waitingCount = appointments.filter((item) => item.status === 'CREATED' || item.status === 'CONFIRMED').length
  const noShowCount = appointments.filter((item) => item.status === 'NO_SHOW').length
  const completedCount = appointments.filter((item) => item.status === 'COMPLETED').length
  const attendance = appointments.length ? Math.round((completedCount / appointments.length) * 100) : 0

  return (
    <section className="metrics-grid">
      <MetricCard label="Записи сегодня" value={todayCount} hint="по календарю врача" tone="blue" icon={<CalendarDays size={22} />} />
      <MetricCard label="Пациентов в очереди" value={waitingCount} hint="созданы и подтверждены" tone="green" icon={<UsersRound size={22} />} />
      <MetricCard label="Неявки" value={noShowCount} hint="пропущенные приемы" tone="orange" icon={<Bell size={22} />} />
      <MetricCard label="Посещаемость" value={`${attendance}%`} hint="завершенные приемы" tone="violet" icon={<BarChart3 size={22} />} />
    </section>
  )
}

function PatientMetrics({ appointments }: { appointments: Appointment[] }) {
  const upcoming = appointments.filter(isFutureActiveAppointment).length
  const completed = appointments.filter((item) => item.status === 'COMPLETED').length
  const cancelled = appointments.filter((item) => item.status === 'CANCELLED').length
  const reminders = appointments.filter((item) => item.reminder_sent).length

  return (
    <section className="metrics-grid">
      <MetricCard label="Будущие приемы" value={upcoming} hint="активные записи" tone="blue" icon={<CalendarDays size={22} />} />
      <MetricCard label="Завершены" value={completed} hint="история посещений" tone="green" icon={<HeartPulse size={22} />} />
      <MetricCard label="Отменены" value={cancelled} hint="отмененные записи" tone="orange" icon={<Bell size={22} />} />
      <MetricCard label="Напоминания" value={reminders} hint="напоминание отправлено" tone="violet" icon={<Shield size={22} />} />
    </section>
  )
}

function PatientDashboard({
  patientId,
  patientName,
  appointments,
  proposals,
  doctors,
  selectedDoctorId,
  doctorSearch,
  freeSlots,
  isPending,
  onSelectDoctor,
  onSearchDoctor,
  onCreateAppointment,
}: {
  patientId?: string
  patientName?: string
  appointments: Appointment[]
  proposals: RescheduleProposal[]
  doctors: Array<{ id: string; full_name: string; specialty: string }>
  selectedDoctorId: string
  doctorSearch: string
  freeSlots: TimeSlot[]
  isPending: boolean
  onSelectDoctor: (doctorId: string) => void
  onSearchDoctor: (value: string) => void
  onCreateAppointment: (slotId: string, visitType: 'PRIMARY' | 'REPEAT' | 'DIAGNOSTIC' | 'CONSULTATION') => void
}) {
  const [bookingSlotId, setBookingSlotId] = useState('')
  const [visitType, setVisitType] = useState<'PRIMARY' | 'REPEAT' | 'DIAGNOSTIC' | 'CONSULTATION'>('CONSULTATION')
  const upcoming = appointments
    .filter(isFutureActiveAppointment)
    .sort((left, right) => new Date(left.appointment_at).getTime() - new Date(right.appointment_at).getTime())
  const history = appointments
    .filter((appointment) => !isFutureActiveAppointment(appointment))
    .sort((left, right) => new Date(right.appointment_at).getTime() - new Date(left.appointment_at).getTime())
  const nextAppointment = upcoming[0]
  const pendingProposals = proposals.filter((proposal) => proposal.status === 'PENDING')

  function submitBooking() {
    if (!bookingSlotId) {
      return
    }
    onCreateAppointment(bookingSlotId, visitType)
    setBookingSlotId('')
  }

  return (
    <section className="patient-dashboard">
      <article className="patient-next-card">
        <div>
          <p className="eyebrow">Ближайший прием</p>
          <h2>{patientName ?? 'Пациент'}</h2>
          {nextAppointment ? (
            <p>
              {new Date(nextAppointment.appointment_at).toLocaleString('ru-RU')} · {doctorDisplayName(nextAppointment)} · {visitTypeLabel(nextAppointment.visit_type)}
            </p>
          ) : (
            <p>Активных записей пока нет.</p>
          )}
        </div>
        {nextAppointment ? (
          <Link className="button button-secondary" to={`/workspace/appointments/${nextAppointment.id}`}>
            Открыть запись
          </Link>
        ) : null}
      </article>

      <section className="patient-main">
        {pendingProposals.length ? (
          <article className="panel admin-wide patient-proposal-panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Согласование</p>
                <h2>Врач предложил перенос</h2>
              </div>
            </div>
            <div className="patient-appointment-list">
              {pendingProposals.map((proposal) => (
                <Link
                  className="patient-appointment-card"
                  key={proposal.id}
                  to={`/workspace/appointments/${proposal.appointment_id}`}
                >
                  <span>{patientDisplayName(proposal.appointment)}</span>
                  <strong>Нужно решение</strong>
                  <small>{formatProposalSlot(proposal)} · {proposal.reason}</small>
                  <span className="patient-card-link">Открыть и ответить</span>
                </Link>
              ))}
            </div>
          </article>
        ) : null}

        <article className="panel">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Онлайн-запись</p>
              <h2>Выбрать врача и свободный слот</h2>
            </div>
            <Link className="button button-primary" to="/booking">
              <span>Открыть форму</span>
            </Link>
          </div>
          {!patientId ? <p className="form-error">К user не привязан patient profile.</p> : null}
          <div className="admin-form">
            <label className="booking-search-field patient-doctor-search">
              <Search size={18} />
              <input
                value={doctorSearch}
                onChange={(event) => {
                  onSearchDoctor(event.target.value)
                  setBookingSlotId('')
                }}
                placeholder="Найти врача по ФИО или специальности"
              />
            </label>
            <div className="patient-doctor-results">
              {doctors.map((doctor) => (
                <button
                  className={selectedDoctorId === doctor.id ? 'selected' : ''}
                  type="button"
                  key={doctor.id}
                  onClick={() => {
                    onSelectDoctor(doctor.id)
                    setBookingSlotId('')
                  }}
                >
                  <strong>{doctor.full_name}</strong>
                  <span>{doctor.specialty}</span>
                </button>
              ))}
              {!doctors.length ? <p className="empty-state">Врачи не найдены.</p> : null}
            </div>
            <select value={bookingSlotId} onChange={(event) => setBookingSlotId(event.target.value)} disabled={!selectedDoctorId}>
              <option value="">{selectedDoctorId ? 'Выбери свободный слот' : 'Сначала выбери врача'}</option>
              {freeSlots.map((slot) => (
                <option key={slot.id} value={slot.id}>
                  {new Date(slot.start_at).toLocaleString('ru-RU')} - {new Date(slot.end_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                </option>
              ))}
            </select>
            <select value={visitType} onChange={(event) => setVisitType(event.target.value as typeof visitType)}>
              <option value="CONSULTATION">Консультация</option>
              <option value="PRIMARY">Первичный прием</option>
              <option value="REPEAT">Повторный прием</option>
              <option value="DIAGNOSTIC">Диагностика</option>
            </select>
            <Button disabled={!patientId || !bookingSlotId || isPending} onClick={submitBooking}>
              Записаться
            </Button>
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Мои приемы</p>
              <h2>Будущие записи</h2>
            </div>
          </div>
          <div className="patient-appointment-list">
            {upcoming.map((appointment) => (
              <PatientAppointmentCard
                key={appointment.id}
                appointment={appointment}
              />
            ))}
            {!upcoming.length ? <p className="empty-state">Будущих записей нет.</p> : null}
          </div>
        </article>

        <article className="panel admin-wide">
          <div className="panel-header">
            <div>
              <p className="eyebrow">История</p>
              <h2>Прошлые и закрытые записи</h2>
            </div>
          </div>
          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Врач</th>
                  <th>Статус</th>
                  <th>Тип</th>
                  <th>ID</th>
                </tr>
              </thead>
              <tbody>
                {history.map((appointment) => (
                  <tr key={appointment.id}>
                    <td>{new Date(appointment.appointment_at).toLocaleString('ru-RU')}</td>
                    <td>{doctorDetails(appointment)}</td>
                    <td>{appointmentStatusLabel(appointment.status)}</td>
                    <td>{visitTypeLabel(appointment.visit_type)}</td>
                    <td>
                      <Link className="button button-ghost" to={`/workspace/appointments/${appointment.id}`}>
                        Открыть
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {!history.length ? <p className="empty-state">История пока пустая.</p> : null}
        </article>
      </section>

    </section>
  )
}

function PatientAppointmentCard({
  appointment,
}: {
  appointment: Appointment
}) {
  return (
    <Link className="patient-appointment-card" to={`/workspace/appointments/${appointment.id}`}>
      <span>{new Date(appointment.appointment_at).toLocaleString('ru-RU')}</span>
      <strong>{appointmentStatusLabel(appointment.status)}</strong>
      <small>{doctorDetails(appointment)} · {visitTypeLabel(appointment.visit_type)}</small>
      <span className="patient-card-link">Открыть запись</span>
    </Link>
  )
}

function DoctorDashboard({
  appointments,
  doctorName,
}: {
  appointments: Appointment[]
  doctorName?: string
}) {
  const navigate = useNavigate()
  const [patientSearch, setPatientSearch] = useState('')
  const openAppointment = useCallback((appointment: Appointment) => {
    navigate(`/workspace/appointments/${appointment.id}`)
  }, [navigate])
  const warnings = appointments
    .filter((item) => item.status === 'CREATED' || item.status === 'NO_SHOW' || item.prediction_result?.risk_level === 'HIGH')
    .slice(0, 3)
  const queue = appointments
    .filter((item) => item.status === 'CREATED' || item.status === 'CONFIRMED')
    .sort((left, right) => new Date(left.appointment_at).getTime() - new Date(right.appointment_at).getTime())
    .slice(0, 5)
  const filteredAppointments = appointments.filter((appointment) =>
    [
      appointment.patient_id,
      appointment.patient?.full_name,
      appointment.patient?.phone,
      appointment.patient?.email,
      appointment.status,
      appointment.visit_type,
      appointment.id,
    ]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(patientSearch.trim().toLowerCase())),
  )

  return (
    <section className="clinic-dashboard">
      {warnings.length ? (
        <article className="doctor-alerts">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Требует внимания</p>
              <h2>Риски и неподтвержденные записи</h2>
            </div>
            <AlertTriangle size={20} />
          </div>
          {warnings.map((appointment) => (
            <div className="doctor-alert" key={appointment.id}>
              <div>
                <strong>{appointmentAlertTitle(appointment)}</strong>
                <span>{formatAppointmentDate(appointment)} · {patientDisplayName(appointment)}</span>
              </div>
              <small>{appointment.prediction_result ? riskBadgeText(appointment) : visitTypeLabel(appointment.visit_type)}</small>
            </div>
          ))}
        </article>
      ) : null}

      <section className="clinic-main">
        <article className="panel clinic-calendar-panel">
          <Suspense fallback={<ContentState variant="loading" title="Загружаю календарь" />}>
            <DoctorCalendar
              appointments={appointments}
              doctorName={doctorName}
              onSelectAppointment={openAppointment}
            />
          </Suspense>
        </article>

        <aside className="clinic-side">
          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Поиск пациента</p>
                <h2>Быстрый поиск</h2>
              </div>
            </div>
            <div className="search-field">
              <Search size={17} />
              <input
                value={patientSearch}
                onChange={(event) => setPatientSearch(event.target.value)}
                placeholder="ФИО, телефон или email"
              />
            </div>
            <div className="clinic-search-results">
              {filteredAppointments.slice(0, 4).map((appointment) => (
                <div key={appointment.id}>
                  <strong>{patientDisplayName(appointment)}</strong>
                  <span>{patientContactLine(appointment)} · {formatAppointmentDate(appointment)} · {appointmentStatusLabel(appointment.status)}</span>
                  <Link className="button button-ghost" to={`/workspace/appointments/${appointment.id}`}>Открыть</Link>
                </div>
              ))}
              {!filteredAppointments.length ? <p className="empty-state">Совпадений нет.</p> : null}
            </div>
          </article>

          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Очередь</p>
                <h2>На прием</h2>
              </div>
              <span className="queue-badge"><CheckCircle2 size={14} /> {queue.length} ожидают</span>
            </div>
            <div className="queue-list">
              {queue.map((appointment) => (
                <div className="queue-item" key={appointment.id}>
                  <span>{new Date(appointment.appointment_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}</span>
                  <strong>{patientDisplayName(appointment)}</strong>
                  <small>{appointmentStatusLabel(appointment.status)}</small>
                  <Link className="button button-ghost" to={`/workspace/appointments/${appointment.id}`}>Детали</Link>
                </div>
              ))}
              {!queue.length ? <p className="empty-state">Очередь пуста.</p> : null}
            </div>
          </article>
        </aside>
      </section>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Таблица</p>
            <h2>Все записи врача</h2>
          </div>
        </div>
        {!appointments.length ? (
          <p className="empty-state">Нет записей для отображения.</p>
        ) : (
          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Пациент</th>
                  <th>Контакты</th>
                  <th>Статус</th>
                  <th>Тип</th>
                  <th>ML риск</th>
                  <th>Кабинет</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {appointments.map((appointment) => (
                  <tr key={appointment.id}>
                    <td>{new Date(appointment.appointment_at).toLocaleString('ru-RU')}</td>
                    <td>{patientDisplayName(appointment)}</td>
                    <td>{patientContactLine(appointment)}</td>
                    <td>{appointmentStatusLabel(appointment.status)}</td>
                    <td>{visitTypeLabel(appointment.visit_type)}</td>
                    <td>{appointment.prediction_result ? riskBadgeText(appointment) : '-'}</td>
                    <td>{appointment.doctor?.cabinet || '-'}</td>
                    <td>
                      <Link className="icon-button" to={`/workspace/appointments/${appointment.id}`} title="Открыть запись">
                        <ArrowRight size={16} />
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </article>
    </section>
  )
}

function isTodayAppointment(appointment: Appointment) {
  const appointmentDate = new Date(appointment.appointment_at)
  const today = new Date()
  return appointmentDate.toDateString() === today.toDateString()
}

function patientDisplayName(appointment: Appointment) {
  return appointment.patient?.full_name || `Пациент ${shortId(appointment.patient_id)}`
}

function patientContactLine(appointment: Appointment) {
  const contacts = [appointment.patient?.phone, appointment.patient?.email].filter(Boolean)
  return contacts.length ? contacts.join(' · ') : 'Контакты не указаны'
}

function appointmentAlertTitle(appointment: Appointment) {
  if (appointment.prediction_result?.risk_level === 'HIGH') {
    return 'Высокий риск неявки'
  }
  if (appointment.status === 'NO_SHOW') {
    return 'Зафиксирована неявка'
  }
  return 'Запись ожидает подтверждения'
}

function riskBadgeText(appointment: Appointment) {
  const prediction = appointment.prediction_result
  if (!prediction) {
    return 'Риск не рассчитан'
  }
  return `${riskLevelLabel(prediction.risk_level)} · ${Math.round(prediction.probability * 100)}%`
}

function doctorDisplayName(appointment: Appointment) {
  return appointment.doctor?.full_name || `Врач ${shortId(appointment.doctor_id)}`
}

function doctorDetails(appointment: Appointment) {
  if (!appointment.doctor) {
    return doctorDisplayName(appointment)
  }
  const cabinet = appointment.doctor.cabinet ? `каб. ${appointment.doctor.cabinet}` : 'кабинет не указан'
  return `${appointment.doctor.full_name} · ${appointment.doctor.specialty} · ${cabinet}`
}

function formatProposalSlot(proposal: RescheduleProposal) {
  if (!proposal.proposed_time_slot) {
    return `Новый слот ${shortId(proposal.proposed_time_slot_id)}`
  }
  return `Новое время: ${new Date(proposal.proposed_time_slot.start_at).toLocaleString('ru-RU')} - ${new Date(proposal.proposed_time_slot.end_at).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })}`
}

function isFutureActiveAppointment(appointment: Appointment) {
  const appointmentTime = new Date(appointment.appointment_at).getTime()
  return appointmentTime >= Date.now() && !['CANCELLED', 'COMPLETED', 'NO_SHOW'].includes(appointment.status)
}

function formatAppointmentDate(appointment: Appointment) {
  return new Date(appointment.appointment_at).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function shortId(value: string) {
  return value.slice(0, 8)
}
