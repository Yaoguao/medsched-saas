import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  ArrowRight,
  Brain,
  Check,
  Eye,
  HeartPulse,
  Microscope,
  Search,
  Shield,
  Stethoscope,
  UserRound,
} from 'lucide-react'
import type { ReactNode } from 'react'
import { useDeferredValue, useMemo, useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { authApi } from '../shared/api/auth'
import { coreApi } from '../shared/api/core'
import type { Doctor, TimeSlot } from '../shared/api/types'
import { getAccessToken } from '../shared/lib/auth'
import { isBookableTimeSlot } from '../shared/lib/timeSlots'
import { Button } from '../shared/ui/Button'

type BookingStep = 1 | 2 | 3 | 4 | 5
type VisitType = 'PRIMARY' | 'REPEAT' | 'DIAGNOSTIC' | 'CONSULTATION'

const bookingSteps = [
  { id: 1, title: 'Специализация', caption: 'Выбор специалиста' },
  { id: 2, title: 'Врач', caption: 'Выбор доктора' },
  { id: 3, title: 'Дата и время', caption: 'Выбор времени' },
  { id: 4, title: 'Контакты', caption: 'Ваши данные' },
  { id: 5, title: 'Подтверждение', caption: 'Проверка данных' },
] as const

const specialtyIcons: Record<string, ReactNode> = {
  терапевт: <Stethoscope size={22} />,
  кардиолог: <HeartPulse size={22} />,
  невролог: <Brain size={22} />,
  эндокринолог: <Shield size={22} />,
  офтальмолог: <Eye size={22} />,
  дерматолог: <Microscope size={22} />,
}

const specialtyDescriptions: Record<string, string> = {
  терапевт: 'Общая диагностика и лечение',
  кардиолог: 'Заболевания сердца и сосудов',
  невролог: 'Нервная система и мозг',
  эндокринолог: 'Гормоны и обмен веществ',
  офтальмолог: 'Лечение глаз и зрения',
  дерматолог: 'Кожные заболевания',
}

export function BookingPage() {
  const token = getAccessToken()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [step, setStep] = useState<BookingStep>(1)
  const [specialty, setSpecialty] = useState('')
  const [doctorId, setDoctorId] = useState('')
  const [doctorSearch, setDoctorSearch] = useState('')
  const [doctorSort, setDoctorSort] = useState<'name_asc' | 'experience_desc' | 'experience_asc'>('name_asc')
  const [doctorPage, setDoctorPage] = useState(0)
  const [dateKey, setDateKey] = useState('')
  const [slotId, setSlotId] = useState('')
  const [visitType, setVisitType] = useState<VisitType>('CONSULTATION')
  const [contactDraft, setContactDraft] = useState<{ phone?: string; email?: string }>({})
  const [createdAppointmentId, setCreatedAppointmentId] = useState('')
  const [openedAt] = useState(() => Date.now())

  const meQuery = useQuery({
    queryKey: ['me'],
    queryFn: authApi.me,
    enabled: Boolean(token),
  })
  const deferredDoctorSearch = useDeferredValue(doctorSearch.trim())
  const specialtiesQuery = useQuery({
    queryKey: ['booking-specialties'],
    queryFn: () => coreApi.doctorSpecialties(true),
    enabled: Boolean(token),
  })
  const doctorsQuery = useQuery({
    queryKey: ['booking-doctors', specialty, deferredDoctorSearch, doctorSort, doctorPage],
    queryFn: () => coreApi.doctors({
      limit: 13,
      offset: doctorPage * 12,
      search: deferredDoctorSearch,
      specialty,
      sort: doctorSort,
      active_only: true,
    }),
    enabled: Boolean(token && specialty),
  })
  const selectedDoctorQuery = useQuery({
    queryKey: ['doctor', doctorId],
    queryFn: () => coreApi.doctor(doctorId),
    enabled: Boolean(token && doctorId),
  })
  const patientQuery = useQuery({
    queryKey: ['booking-patient', meQuery.data?.patient_id],
    queryFn: () => coreApi.patient(meQuery.data?.patient_id ?? ''),
    enabled: Boolean(token && meQuery.data?.patient_id),
  })
  const schedulesQuery = useQuery({
    queryKey: ['booking-schedules', doctorId],
    queryFn: () => coreApi.doctorSchedules(doctorId, { limit: 100 }),
    enabled: Boolean(token && doctorId),
  })

  const doctorResults = useMemo(() => doctorsQuery.data?.items ?? [], [doctorsQuery.data?.items])
  const doctors = doctorResults.slice(0, 12)
  const specialties = specialtiesQuery.data ?? []
  const selectedDoctor = selectedDoctorQuery.data
  const availableSlots = useMemo(
    () => (schedulesQuery.data?.items ?? [])
      .flatMap((schedule) => schedule.time_slots ?? [])
      .filter((slot) => isBookableTimeSlot(slot, openedAt))
      .sort((left, right) => new Date(left.start_at).getTime() - new Date(right.start_at).getTime()),
    [openedAt, schedulesQuery.data?.items],
  )
  const dates = collectSlotDates(availableSlots)
  const slotsByDate = availableSlots.filter((slot) => slotDateKey(slot.start_at) === dateKey)
  const selectedSlot = availableSlots.find((slot) => slot.id === slotId)
  const contactPhone = contactDraft.phone ?? patientQuery.data?.phone ?? meQuery.data?.phone ?? ''
  const contactEmail = contactDraft.email ?? patientQuery.data?.email ?? meQuery.data?.email ?? ''

  const createMutation = useMutation({
    mutationFn: async () => {
      const patientId = meQuery.data?.patient_id
      if (!patientId || !slotId) {
        throw new Error('Не выбран пациент или слот')
      }
      if (patientQuery.data && (contactPhone !== patientQuery.data.phone || contactEmail !== patientQuery.data.email)) {
        await coreApi.updatePatient(patientId, {
          phone: contactPhone,
          email: contactEmail,
        })
      }
      return coreApi.createAppointment({
        patient_id: patientId,
        time_slot_id: slotId,
        booking_channel: 'WEB',
        visit_type: visitType,
      })
    },
    onSuccess: (appointment) => {
      setCreatedAppointmentId(appointment.id)
      queryClient.invalidateQueries({ queryKey: ['home-patient-appointments', meQuery.data?.patient_id] })
      queryClient.invalidateQueries({ queryKey: ['booking-schedules', doctorId] })
      queryClient.invalidateQueries({ queryKey: ['booking-patient', meQuery.data?.patient_id] })
    },
  })

  if (!token) {
    return <Navigate to="/login" replace />
  }

  const canGoNext = stepReady(step, {
    specialty,
    doctorId,
    slotId,
    contactPhone,
    contactEmail,
  })
  const isPatient = meQuery.data?.role === 'PATIENT'
  const missingPatientProfile = Boolean(meQuery.data && !meQuery.data.patient_id)

  function selectSpecialty(nextSpecialty: string) {
    setSpecialty(nextSpecialty)
    setDoctorId('')
    setDoctorSearch('')
    setDoctorPage(0)
    setDateKey('')
    setSlotId('')
  }

  function selectDoctor(nextDoctorId: string) {
    setDoctorId(nextDoctorId)
    setDateKey('')
    setSlotId('')
  }

  function nextStep() {
    if (step < 5 && canGoNext) {
      setStep((step + 1) as BookingStep)
    }
  }

  function prevStep() {
    if (step > 1) {
      setStep((step - 1) as BookingStep)
    } else {
      navigate('/home')
    }
  }

  return (
    <main className="booking-shell">
      <header className="booking-topbar">
        <Link className="booking-back-link" to="/home">
          <ArrowLeft size={18} />
          <span>Назад в кабинет</span>
        </Link>
        <div className="doctor-chip">
          <UserRound size={18} />
          <span>{meQuery.data?.email ?? 'Пациент'}</span>
        </div>
      </header>

      <section className="booking-hero">
        <h1>Онлайн-запись к врачу</h1>
        <p>Запишитесь на прием к специалисту за несколько простых шагов</p>
      </section>

      <section className="booking-progress" aria-label="Шаги онлайн-записи">
        {bookingSteps.map((item) => (
          <button
            className={item.id === step ? 'active' : item.id < step ? 'done' : ''}
            type="button"
            key={item.id}
            onClick={() => {
              if (item.id < step) {
                setStep(item.id)
              }
            }}
          >
            <span>{item.id < step ? <Check size={16} /> : item.id}</span>
            <strong>{item.title}</strong>
            <small>{item.caption}</small>
          </button>
        ))}
      </section>

      {!isPatient ? (
        <section className="booking-card">
          <p className="form-error">Онлайн-запись доступна для роли пациента.</p>
        </section>
      ) : missingPatientProfile ? (
        <section className="booking-card">
          <p className="form-error">К текущему user не привязан patient profile. Сначала создайте или привяжите пациента.</p>
        </section>
      ) : (
        <section className="booking-card">
          {step === 1 ? (
            <SpecialtyStep
              specialties={specialties}
              selectedSpecialty={specialty}
              onSelect={selectSpecialty}
            />
          ) : null}

          {step === 2 ? (
            <DoctorStep
              doctors={doctors}
              selectedDoctorId={doctorId}
              search={doctorSearch}
              sort={doctorSort}
              page={doctorPage}
              isLoading={doctorsQuery.isFetching}
              canLoadNext={doctorResults.length > 12}
              onSelect={selectDoctor}
              onSearch={(value) => {
                setDoctorSearch(value)
                setDoctorPage(0)
              }}
              onSort={(value) => {
                setDoctorSort(value)
                setDoctorPage(0)
              }}
              onPage={setDoctorPage}
            />
          ) : null}

          {step === 3 ? (
            <TimeStep
              dates={dates}
              dateKey={dateKey}
              slots={slotsByDate}
              slotId={slotId}
              isLoading={schedulesQuery.isLoading}
              onSelectDate={(nextDate) => {
                setDateKey(nextDate)
                setSlotId('')
              }}
              onSelectSlot={setSlotId}
            />
          ) : null}

          {step === 4 ? (
            <ContactStep
              patientName={patientQuery.data?.full_name}
              phone={contactPhone}
              email={contactEmail}
              visitType={visitType}
              onChangePhone={(phone) => setContactDraft((draft) => ({ ...draft, phone }))}
              onChangeEmail={(email) => setContactDraft((draft) => ({ ...draft, email }))}
              onChangeVisitType={setVisitType}
            />
          ) : null}

          {step === 5 ? (
            <ConfirmStep
              doctor={selectedDoctor}
              slot={selectedSlot}
              patientName={patientQuery.data?.full_name}
              phone={contactPhone}
              email={contactEmail}
              visitType={visitType}
              createdAppointmentId={createdAppointmentId}
            />
          ) : null}
        </section>
      )}

      {isPatient && !missingPatientProfile ? (
        <footer className="booking-actions">
          <Button variant="secondary" icon={<ArrowLeft size={16} />} onClick={prevStep}>
            Назад
          </Button>
          {step < 5 ? (
            <Button disabled={!canGoNext} icon={<ArrowRight size={16} />} onClick={nextStep}>
              Далее
            </Button>
          ) : createdAppointmentId ? (
            <Button onClick={() => navigate('/home')}>Готово</Button>
          ) : (
            <Button disabled={!canGoNext || createMutation.isPending} onClick={() => createMutation.mutate()}>
              Подтвердить запись
            </Button>
          )}
        </footer>
      ) : null}

      {createMutation.isError ? (
        <p className="booking-error">{createMutation.error.message}</p>
      ) : null}
    </main>
  )
}

function SpecialtyStep({
  specialties,
  selectedSpecialty,
  onSelect,
}: {
  specialties: string[]
  selectedSpecialty: string
  onSelect: (specialty: string) => void
}) {
  const [search, setSearch] = useState('')
  const filteredSpecialties = specialties
    .filter((item) => normalize(item).includes(normalize(search)))
    .sort((left, right) => left.localeCompare(right, 'ru'))

  return (
    <>
      <div className="booking-card-header">
        <h2>Специализация</h2>
        <p>Выбор специалиста</p>
      </div>
      <label className="booking-search-field">
        <Search size={18} />
        <input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Найти специализацию"
        />
      </label>
      <div className="specialty-grid">
        {filteredSpecialties.map((item) => {
          const key = normalize(item)
          return (
            <button
              className={selectedSpecialty === item ? 'specialty-card selected' : 'specialty-card'}
              type="button"
              key={item}
              onClick={() => onSelect(item)}
            >
              {specialtyIcons[key] ?? <Stethoscope size={22} />}
              <strong>{item}</strong>
              <span>{specialtyDescriptions[key] ?? 'Консультация специалиста'}</span>
            </button>
          )
        })}
        {!filteredSpecialties.length ? <p className="empty-state">Специализации не найдены.</p> : null}
      </div>
    </>
  )
}

function DoctorStep({
  doctors,
  selectedDoctorId,
  search,
  sort,
  page,
  isLoading,
  canLoadNext,
  onSelect,
  onSearch,
  onSort,
  onPage,
}: {
  doctors: Doctor[]
  selectedDoctorId: string
  search: string
  sort: 'name_asc' | 'experience_desc' | 'experience_asc'
  page: number
  isLoading: boolean
  canLoadNext: boolean
  onSelect: (doctorId: string) => void
  onSearch: (value: string) => void
  onSort: (value: 'name_asc' | 'experience_desc' | 'experience_asc') => void
  onPage: (page: number) => void
}) {
  return (
    <>
      <div className="booking-card-header">
        <h2>Выберите врача</h2>
        <p>Доступные специалисты выбранного направления</p>
      </div>
      <div className="booking-doctor-toolbar">
        <label className="booking-search-field">
          <Search size={18} />
          <input
            value={search}
            onChange={(event) => onSearch(event.target.value)}
            placeholder="ФИО, специальность или кабинет"
          />
        </label>
        <select
          aria-label="Сортировка врачей"
          value={sort}
          onChange={(event) => onSort(event.target.value as typeof sort)}
        >
          <option value="name_asc">По фамилии</option>
          <option value="experience_desc">Сначала опытные</option>
          <option value="experience_asc">Сначала меньший стаж</option>
        </select>
      </div>
      {isLoading ? <p className="empty-state">Ищем врачей...</p> : null}
      <div className="booking-doctor-list">
        {doctors.map((doctor) => (
          <button
            className={selectedDoctorId === doctor.id ? 'booking-doctor-card selected' : 'booking-doctor-card'}
            type="button"
            key={doctor.id}
            onClick={() => onSelect(doctor.id)}
          >
            <div className="doctor-avatar"><UserRound size={22} /></div>
            <div>
              <strong>{doctor.full_name}</strong>
              <span>{doctor.specialty}</span>
              <small>Кабинет {doctor.cabinet || 'не указан'} · стаж {doctor.experience_years} лет</small>
            </div>
          </button>
        ))}
        {!doctors.length && !isLoading ? <p className="empty-state">Подходящие врачи не найдены.</p> : null}
      </div>
      <div className="booking-pagination">
        <Button
          type="button"
          variant="secondary"
          disabled={page === 0 || isLoading}
          onClick={() => onPage(Math.max(0, page - 1))}
        >
          Назад
        </Button>
        <span>Страница {page + 1}</span>
        <Button
          type="button"
          variant="secondary"
          disabled={!canLoadNext || isLoading}
          onClick={() => onPage(page + 1)}
        >
          Далее
        </Button>
      </div>
    </>
  )
}

function TimeStep({
  dates,
  dateKey,
  slots,
  slotId,
  isLoading,
  onSelectDate,
  onSelectSlot,
}: {
  dates: Array<{ key: string; label: string; weekday: string }>
  dateKey: string
  slots: TimeSlot[]
  slotId: string
  isLoading: boolean
  onSelectDate: (dateKey: string) => void
  onSelectSlot: (slotId: string) => void
}) {
  return (
    <>
      <div className="booking-card-header">
        <h2>Дата и время</h2>
        <p>Выберите удобный день и свободное время приема</p>
      </div>
      {isLoading ? <p className="empty-state">Загружаем расписание...</p> : null}
      <div className="booking-date-row">
        {dates.map((date) => (
          <button
            className={dateKey === date.key ? 'booking-date-card selected' : 'booking-date-card'}
            type="button"
            key={date.key}
            onClick={() => onSelectDate(date.key)}
          >
            <strong>{date.label}</strong>
            <span>{date.weekday}</span>
          </button>
        ))}
      </div>
      {dateKey ? (
        <div className="booking-slot-grid">
          {slots.map((slot) => (
            <button
              className={slotId === slot.id ? 'booking-slot selected' : 'booking-slot'}
              type="button"
              key={slot.id}
              onClick={() => onSelectSlot(slot.id)}
            >
              {new Date(slot.start_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
            </button>
          ))}
          {!slots.length ? <p className="empty-state">На выбранный день свободных слотов нет.</p> : null}
        </div>
      ) : !isLoading ? (
        <p className="empty-state">Выберите дату, чтобы увидеть время приема.</p>
      ) : null}
    </>
  )
}

function ContactStep({
  patientName,
  phone,
  email,
  visitType,
  onChangePhone,
  onChangeEmail,
  onChangeVisitType,
}: {
  patientName?: string
  phone: string
  email: string
  visitType: VisitType
  onChangePhone: (phone: string) => void
  onChangeEmail: (email: string) => void
  onChangeVisitType: (visitType: VisitType) => void
}) {
  return (
    <>
      <div className="booking-card-header">
        <h2>Контактные данные</h2>
        <p>Проверьте контакты для подтверждения и уведомлений</p>
      </div>
      <div className="booking-contact-grid">
        <label>
          <span>Пациент</span>
          <input value={patientName ?? 'Пациент'} disabled />
        </label>
        <label>
          <span>Телефон</span>
          <input value={phone} onChange={(event) => onChangePhone(event.target.value)} placeholder="+79990000000" />
        </label>
        <label>
          <span>Email</span>
          <input value={email} onChange={(event) => onChangeEmail(event.target.value)} placeholder="patient@example.com" />
        </label>
        <label>
          <span>Тип приема</span>
          <select value={visitType} onChange={(event) => onChangeVisitType(event.target.value as VisitType)}>
            <option value="CONSULTATION">Консультация</option>
            <option value="PRIMARY">Первичный прием</option>
            <option value="REPEAT">Повторный прием</option>
            <option value="DIAGNOSTIC">Диагностика</option>
          </select>
        </label>
      </div>
    </>
  )
}

function ConfirmStep({
  doctor,
  slot,
  patientName,
  phone,
  email,
  visitType,
  createdAppointmentId,
}: {
  doctor?: Doctor
  slot?: TimeSlot
  patientName?: string
  phone: string
  email: string
  visitType: VisitType
  createdAppointmentId: string
}) {
  return (
    <>
      <div className="booking-card-header">
        <h2>{createdAppointmentId ? 'Запись создана' : 'Подтверждение'}</h2>
        <p>{createdAppointmentId ? 'Данные приема сохранены' : 'Проверьте данные перед созданием записи'}</p>
      </div>
      <div className={createdAppointmentId ? 'booking-summary success' : 'booking-summary'}>
        {createdAppointmentId ? (
          <div className="booking-success-mark">
            <Check size={28} />
          </div>
        ) : null}
        <SummaryLine label="Пациент" value={patientName ?? 'Пациент'} />
        <SummaryLine label="Врач" value={doctor ? `${doctor.full_name} · ${doctor.specialty}` : 'Не выбран'} />
        <SummaryLine label="Кабинет" value={doctor?.cabinet || 'Не указан'} />
        <SummaryLine label="Дата и время" value={slot ? new Date(slot.start_at).toLocaleString('ru-RU') : 'Не выбрано'} />
        <SummaryLine label="Тип приема" value={visitType} />
        <SummaryLine label="Телефон" value={phone || 'Не указан'} />
        <SummaryLine label="Email" value={email || 'Не указан'} />
        {createdAppointmentId ? <SummaryLine label="ID записи" value={createdAppointmentId} /> : null}
      </div>
    </>
  )
}

function SummaryLine({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

function collectSlotDates(slots: TimeSlot[]) {
  const uniqueDates = Array.from(new Set(slots.map((slot) => slotDateKey(slot.start_at))))
  return uniqueDates.slice(0, 12).map((key) => {
    const date = new Date(key)
    return {
      key,
      label: date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' }),
      weekday: date.toLocaleDateString('ru-RU', { weekday: 'short' }),
    }
  })
}

function slotDateKey(value: string) {
  const date = new Date(value)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function normalize(value: string) {
  return value.trim().toLowerCase()
}

function stepReady(
  step: BookingStep,
  state: {
    specialty: string
    doctorId: string
    slotId: string
    contactPhone: string
    contactEmail: string
  },
) {
  if (step === 1) return Boolean(state.specialty)
  if (step === 2) return Boolean(state.doctorId)
  if (step === 3) return Boolean(state.slotId)
  if (step === 4) return Boolean(state.contactPhone.trim() && state.contactEmail.trim())
  return Boolean(state.specialty && state.doctorId && state.slotId && state.contactPhone.trim() && state.contactEmail.trim())
}
