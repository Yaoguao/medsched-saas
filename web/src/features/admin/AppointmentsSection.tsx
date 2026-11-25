import type { FormEvent } from 'react'
import { useMemo, useState } from 'react'
import { CalendarDays, CheckCircle2, Clock3, TriangleAlert } from 'lucide-react'
import { Link } from 'react-router-dom'
import type {
  Appointment,
  CreateAppointmentPayload,
  Doctor,
  Patient,
  Schedule,
  TimeSlot,
  UserRole,
} from '../../shared/api/types'
import { appointmentStatusLabel, bookingChannelLabel, riskLevelLabel, slotStatusLabel, visitTypeLabel } from '../../shared/lib/labels'
import { isBookableTimeSlot } from '../../shared/lib/timeSlots'
import { Button } from '../../shared/ui/Button'
import { MetricCard } from '../../shared/ui/MetricCard'
import { AdminTable } from './AdminTable'
import { formString } from './form'

type AppointmentsSectionProps = {
  role?: UserRole
  preselectedPatientId?: string
  appointments: Appointment[]
  patients: Patient[]
  doctors: Doctor[]
  schedules: Schedule[]
  selectedDoctorId: string
  selectedTimeSlotId: string
  isPending: boolean
  isSchedulesLoading: boolean
  onSelectDoctor: (doctorId: string) => void
  onSelectTimeSlot: (slotId: string) => void
  onCreateAppointment: (payload: CreateAppointmentPayload) => void
  onPredict: (appointmentId: string) => void
}

export function AppointmentsSection({
  role,
  preselectedPatientId = '',
  appointments,
  patients,
  doctors,
  schedules,
  selectedDoctorId,
  selectedTimeSlotId,
  isPending,
  isSchedulesLoading,
  onSelectDoctor,
  onSelectTimeSlot,
  onCreateAppointment,
  onPredict,
}: AppointmentsSectionProps) {
  const [query, setQuery] = useState('')
  const [selectedPatientId, setSelectedPatientId] = useState(preselectedPatientId)
  const [selectedAppointmentId, setSelectedAppointmentId] = useState('')
  const [openedAt] = useState(() => Date.now())

  function createAppointment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreateAppointment({
      patient_id: formString(form, 'patient_id', selectedPatientId),
      time_slot_id: formString(form, 'time_slot_id', selectedTimeSlotId),
      booking_channel: formString(form, 'booking_channel', 'WEB') as CreateAppointmentPayload['booking_channel'],
      visit_type: formString(form, 'visit_type', 'PRIMARY') as CreateAppointmentPayload['visit_type'],
    })
  }

  const freeSlots = schedules
    .flatMap((schedule) => schedule.time_slots ?? [])
    .filter((slot) => isBookableTimeSlot(slot, openedAt))
  const todayCount = appointments.filter((appointment) => isToday(appointment.appointment_at)).length
  const confirmedCount = appointments.filter((appointment) => appointment.status === 'CONFIRMED').length
  const highRiskCount = appointments.filter((appointment) => appointment.prediction_result?.risk_level === 'HIGH').length
  const selectedPatient = patients.find((patient) => patient.id === selectedPatientId)
  const selectedDoctor = doctors.find((doctor) => doctor.id === selectedDoctorId)
  const selectedSlot = freeSlots.find((slot) => slot.id === selectedTimeSlotId)
  const selectedAppointment = appointments.find((appointment) => appointment.id === selectedAppointmentId)
  const isRegistrarWorkspace = role === 'REGISTRATOR'
  const filteredAppointments = useMemo(() => {
    const search = query.trim().toLowerCase()
    if (!search) {
      return appointments
    }
    return appointments.filter((appointment) =>
      [
        appointment.id,
        appointment.patient_id,
        appointment.patient?.full_name,
        appointment.patient?.phone,
        appointment.doctor_id,
        appointment.doctor?.full_name,
        appointment.doctor?.specialty,
        appointment.time_slot_id,
        appointment.status,
        appointment.prediction_result?.risk_level,
        appointment.visit_type,
        appointment.booking_channel,
      ]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)),
    )
  }, [appointments, query])

  return (
    <>
    <RegistrarBookingFlow
      patients={patients}
      doctors={doctors}
      freeSlots={freeSlots}
      selectedPatient={selectedPatient}
      selectedDoctor={selectedDoctor}
      selectedSlot={selectedSlot}
      selectedPatientId={selectedPatientId}
      selectedDoctorId={selectedDoctorId}
      selectedTimeSlotId={selectedTimeSlotId}
      isRegistrarWorkspace={isRegistrarWorkspace}
      isSchedulesLoading={isSchedulesLoading}
      onSelectPatient={setSelectedPatientId}
      onSelectDoctor={onSelectDoctor}
      onSelectTimeSlot={onSelectTimeSlot}
    />

    <section className="metrics-grid">
      <MetricCard label="Всего записей" value={appointments.length} hint="доступны текущей роли" tone="blue" icon={<CalendarDays size={22} />} />
      <MetricCard label="Сегодня" value={todayCount} hint="приемы на текущую дату" tone="green" icon={<Clock3 size={22} />} />
      <MetricCard label="Подтверждены" value={confirmedCount} hint="готовы к приему" tone="violet" icon={<CheckCircle2 size={22} />} />
      <MetricCard label="Высокий риск" value={highRiskCount} hint="по ML-прогнозу" tone="orange" icon={<TriangleAlert size={22} />} />
    </section>

    <section className="admin-two-column">
      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Записи</p>
            <h2>Создать запись</h2>
          </div>
        </div>
        <form id="appointment-create-form" className="admin-form" onSubmit={createAppointment}>
          <select
            value={selectedDoctorId}
            onChange={(event) => onSelectDoctor(event.target.value)}
            required
          >
            <option value="">Выбери врача</option>
            {doctors.map((doctor) => (
              <option key={doctor.id} value={doctor.id}>
                {doctor.full_name} · {doctor.specialty}
              </option>
            ))}
          </select>
          <select name="patient_id" value={selectedPatientId} onChange={(event) => setSelectedPatientId(event.target.value)} required>
            <option value="">Выбери пациента</option>
            {patients.map((patient) => (
              <option key={patient.id} value={patient.id}>
                {patient.full_name}
              </option>
            ))}
          </select>
          <select
            name="time_slot_id"
            value={selectedTimeSlotId}
            onChange={(event) => onSelectTimeSlot(event.target.value)}
            required
          >
            <option value="">
              {selectedDoctorId ? 'Выбери свободный слот' : 'Сначала выбери врача'}
            </option>
            {freeSlots.map((slot) => (
              <option key={slot.id} value={slot.id}>
                {formatSlot(slot)}
              </option>
            ))}
          </select>
          <select name="booking_channel" defaultValue="WEB">
            <option value="WEB">{bookingChannelLabel('WEB')}</option>
            <option value="REGISTRY">{bookingChannelLabel('REGISTRY')}</option>
            <option value="PHONE">{bookingChannelLabel('PHONE')}</option>
          </select>
          <select name="visit_type" defaultValue="PRIMARY">
            <option value="PRIMARY">{visitTypeLabel('PRIMARY')}</option>
            <option value="REPEAT">{visitTypeLabel('REPEAT')}</option>
            <option value="DIAGNOSTIC">{visitTypeLabel('DIAGNOSTIC')}</option>
            <option value="CONSULTATION">{visitTypeLabel('CONSULTATION')}</option>
          </select>
          <Button type="submit" disabled={isPending || !selectedTimeSlotId}>Создать запись</Button>
        </form>
        <p className="section-note">
          {isSchedulesLoading
            ? 'Загружаю свободные слоты врача...'
            : 'Слот можно выбрать здесь или нажать "В запись" в разделе расписаний.'}
        </p>
      </article>

      <AppointmentDetailsPanel
        appointments={filteredAppointments}
        appointment={selectedAppointment}
        selectedAppointmentId={selectedAppointmentId}
        isPending={isPending}
        onSelectAppointment={setSelectedAppointmentId}
        onPredict={onPredict}
      />

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Свободные окна</p>
            <h2>Доступные слоты</h2>
          </div>
        </div>
        {!selectedDoctorId ? (
          <p className="empty-state">Выбери врача в форме выше, чтобы увидеть свободные слоты.</p>
        ) : null}
        {selectedDoctorId && !freeSlots.length && !isSchedulesLoading ? (
          <p className="empty-state">Свободных слотов нет. Создай слот в разделе расписаний.</p>
        ) : null}
        {freeSlots.length ? (
          <div className="slot-grid">
            {freeSlots.map((slot) => (
              <button
                key={slot.id}
                type="button"
                className={slot.id === selectedTimeSlotId ? 'slot-card selected' : 'slot-card'}
                onClick={() => onSelectTimeSlot(slot.id)}
              >
                <strong>{formatSlot(slot)}</strong>
                <span>{slotStatusLabel(slot.status)}</span>
              </button>
            ))}
          </div>
        ) : null}
      </article>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Список</p>
            <h2>Записи</h2>
          </div>
          <input
            className="table-filter"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Поиск по пациенту, врачу или статусу"
          />
        </div>
        <AdminTable
          items={filteredAppointments}
          empty="Записей пока нет."
          columns={[
            { key: 'time', title: 'Дата', render: (appointment) => new Date(appointment.appointment_at).toLocaleString('ru-RU') },
            { key: 'patient', title: 'Пациент', render: (appointment) => appointment.patient?.full_name ?? <code>{appointment.patient_id}</code> },
            { key: 'doctor', title: 'Врач', render: (appointment) => appointment.doctor ? `${appointment.doctor.full_name} · ${appointment.doctor.specialty}` : <code>{appointment.doctor_id}</code> },
            { key: 'status', title: 'Статус', render: (appointment) => <AppointmentStatusBadge status={appointment.status} /> },
            { key: 'risk', title: 'ML риск', render: (appointment) => appointment.prediction_result ? <RiskMeter level={appointment.prediction_result.risk_level} probability={appointment.prediction_result.probability} /> : '-' },
            { key: 'visit', title: 'Тип', render: (appointment) => visitTypeLabel(appointment.visit_type) },
            {
              key: 'actions',
              title: 'Действия',
              render: (appointment) => (
                <div className="table-actions">
                  <Link className="button button-secondary" to={`/workspace/appointments/${appointment.id}`}>
                    Открыть
                  </Link>
                </div>
              ),
            },
          ]}
        />
      </article>
    </section>
    </>
  )
}

function formatSlot(slot: TimeSlot) {
  return `${new Date(slot.start_at).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })} - ${new Date(slot.end_at).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })}`
}

function AppointmentDetailsPanel({
  appointments,
  appointment,
  selectedAppointmentId,
  isPending,
  onSelectAppointment,
  onPredict,
}: {
  appointments: Appointment[]
  appointment?: Appointment
  selectedAppointmentId: string
  isPending: boolean
  onSelectAppointment: (appointmentId: string) => void
  onPredict: (appointmentId: string) => void
}) {
  return (
    <article className="panel appointment-workbench">
      <div className="panel-header">
        <div>
          <p className="eyebrow">Запись</p>
          <h2>Карточка выбранного приема</h2>
        </div>
        {appointment ? <AppointmentStatusBadge status={appointment.status} /> : null}
      </div>

      <select value={selectedAppointmentId} onChange={(event) => onSelectAppointment(event.target.value)}>
        <option value="">Выбери запись из списка</option>
        {appointments.map((item) => (
          <option key={item.id} value={item.id}>
            {appointmentOptionLabel(item)}
          </option>
        ))}
      </select>

      {!appointment ? (
        <p className="empty-state">Открой запись из списка ниже, чтобы увидеть детали и доступные действия.</p>
      ) : (
        <>
          <div className="appointment-detail-grid">
            <DetailLine label="Пациент" value={appointment.patient?.full_name ?? appointment.patient_id} />
            <DetailLine label="Контакт" value={appointment.patient?.phone || appointment.patient?.email || 'Не указан'} />
            <DetailLine label="Врач" value={appointment.doctor ? `${appointment.doctor.full_name} · ${appointment.doctor.specialty}` : appointment.doctor_id} />
            <DetailLine label="Кабинет" value={appointment.doctor?.cabinet || 'Не указан'} />
            <DetailLine label="Дата и время" value={new Date(appointment.appointment_at).toLocaleString('ru-RU')} />
            <DetailLine label="Тип приема" value={visitTypeLabel(appointment.visit_type)} />
            <DetailLine label="Канал" value={bookingChannelLabel(appointment.booking_channel)} />
            <DetailLine label="Слот" value={appointment.time_slot ? formatSlot(appointment.time_slot) : appointment.time_slot_id} />
          </div>

          <div className="appointment-workbench-row">
            <div className="appointment-risk-card">
              <p className="eyebrow">ML риск</p>
              {appointment.prediction_result ? (
                <RiskMeter
                  level={appointment.prediction_result.risk_level}
                  probability={appointment.prediction_result.probability}
                />
              ) : (
                <span>Прогноз еще не рассчитан.</span>
              )}
            </div>
            <div className="status-actions">
              <Link className="button button-primary" to={`/workspace/appointments/${appointment.id}`}>
                Открыть полную карточку
              </Link>
              <Button type="button" variant="secondary" disabled={isPending} onClick={() => onPredict(appointment.id)}>
                Рассчитать ML
              </Button>
            </div>
          </div>

          <div className="appointment-history">
            <p className="eyebrow">История статусов</p>
            {appointment.status_history?.length ? (
              appointment.status_history.map((history) => (
                <div key={history.id}>
                  <strong>
                    {history.old_status
                      ? `${appointmentStatusLabel(history.old_status)} → ${appointmentStatusLabel(history.new_status)}`
                      : appointmentStatusLabel(history.new_status)}
                  </strong>
                  <span>{new Date(history.changed_at).toLocaleString('ru-RU')} · {history.reason || 'Без комментария'}</span>
                </div>
              ))
            ) : (
              <span>История пока пустая.</span>
            )}
          </div>
        </>
      )}
    </article>
  )
}

function RegistrarBookingFlow({
  patients,
  doctors,
  freeSlots,
  selectedPatient,
  selectedDoctor,
  selectedSlot,
  selectedPatientId,
  selectedDoctorId,
  selectedTimeSlotId,
  isRegistrarWorkspace,
  isSchedulesLoading,
  onSelectPatient,
  onSelectDoctor,
  onSelectTimeSlot,
}: {
  patients: Patient[]
  doctors: Doctor[]
  freeSlots: TimeSlot[]
  selectedPatient?: Patient
  selectedDoctor?: Doctor
  selectedSlot?: TimeSlot
  selectedPatientId: string
  selectedDoctorId: string
  selectedTimeSlotId: string
  isRegistrarWorkspace: boolean
  isSchedulesLoading: boolean
  onSelectPatient: (patientId: string) => void
  onSelectDoctor: (doctorId: string) => void
  onSelectTimeSlot: (slotId: string) => void
}) {
  const canCreateAppointment = Boolean(selectedPatientId && selectedDoctorId && selectedTimeSlotId)

  return (
    <section className={isRegistrarWorkspace ? 'registrar-flow primary' : 'registrar-flow'}>
      <div className="registrar-flow-header">
        <div>
          <p className="eyebrow">Оформление записи</p>
          <h2>Пациент → врач → свободное время → прием</h2>
        </div>
        <span className={`workspace-badge ${canCreateAppointment ? 'success' : 'info'}`}>
          {canCreateAppointment ? 'Готово к созданию' : 'Заполни шаги'}
        </span>
      </div>

      <div className="registrar-flow-grid">
        <article className={selectedPatient ? 'flow-step complete' : 'flow-step'}>
          <span>1</span>
          <div>
            <strong>Пациент</strong>
            <select value={selectedPatientId} onChange={(event) => onSelectPatient(event.target.value)}>
              <option value="">Выбери пациента</option>
              {patients.map((patient) => (
                <option key={patient.id} value={patient.id}>
                  {patient.full_name} · {patient.phone || patient.email || patient.id}
                </option>
              ))}
            </select>
            {selectedPatient ? (
              <small>{selectedPatient.phone || 'Телефон не указан'} · {selectedPatient.email || 'email не указан'}</small>
            ) : (
              <small>Если пациента нет в списке, сначала создай карту в разделе пациентов.</small>
            )}
          </div>
        </article>

        <article className={selectedDoctor ? 'flow-step complete' : 'flow-step'}>
          <span>2</span>
          <div>
            <strong>Врач</strong>
            <select value={selectedDoctorId} onChange={(event) => onSelectDoctor(event.target.value)}>
              <option value="">Выбери врача</option>
              {doctors.map((doctor) => (
                <option key={doctor.id} value={doctor.id}>
                  {doctor.full_name} · {doctor.specialty}
                </option>
              ))}
            </select>
            {selectedDoctor ? (
              <small>{selectedDoctor.specialty} · каб. {selectedDoctor.cabinet || 'не указан'}</small>
            ) : (
              <small>После выбора врача загрузятся свободные слоты.</small>
            )}
          </div>
        </article>

        <article className={selectedSlot ? 'flow-step complete' : 'flow-step'}>
          <span>3</span>
          <div>
            <strong>Свободный слот</strong>
            {!selectedDoctorId ? (
              <small>Сначала выбери врача.</small>
            ) : isSchedulesLoading ? (
              <small>Загружаю расписание врача...</small>
            ) : freeSlots.length ? (
              <div className="flow-slot-list">
                {freeSlots.slice(0, 5).map((slot) => (
                  <button
                    key={slot.id}
                    type="button"
                    className={slot.id === selectedTimeSlotId ? 'selected' : undefined}
                    onClick={() => onSelectTimeSlot(slot.id)}
                  >
                    {formatSlot(slot)}
                  </button>
                ))}
              </div>
            ) : (
              <small>Свободных слотов нет. Менеджер должен добавить расписание.</small>
            )}
          </div>
        </article>

        <article className={canCreateAppointment ? 'flow-step flow-summary complete' : 'flow-step flow-summary'}>
          <span>4</span>
          <div>
            <strong>Итог записи</strong>
            <small>{selectedPatient?.full_name ?? 'Пациент не выбран'}</small>
            <small>{selectedDoctor ? `${selectedDoctor.full_name} · ${selectedDoctor.specialty}` : 'Врач не выбран'}</small>
            <small>{selectedSlot ? formatSlot(selectedSlot) : 'Слот не выбран'}</small>
            <Button form="appointment-create-form" type="submit" disabled={!canCreateAppointment}>
              Создать запись
            </Button>
          </div>
        </article>
      </div>
    </section>
  )
}

function DetailLine({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value || '-'}</strong>
    </div>
  )
}

function appointmentOptionLabel(appointment: Appointment) {
  const patient = appointment.patient?.full_name ?? appointment.patient_id
  const doctor = appointment.doctor?.full_name ?? appointment.doctor_id
  return `${new Date(appointment.appointment_at).toLocaleString('ru-RU')} · ${patient} · ${doctor}`
}

function AppointmentStatusBadge({ status }: { status: string }) {
  return (
    <span className={`workspace-badge ${statusTone(status)}`}>
      {appointmentStatusLabel(status)}
    </span>
  )
}

function RiskMeter({ level, probability }: { level: string; probability: number }) {
  const percent = Math.round(probability * 100)
  return (
    <div className="risk-meter">
      <div>
        <span className={riskTone(level)} style={{ width: `${percent}%` }} />
      </div>
      <strong>{riskLevelLabel(level)} · {percent}%</strong>
    </div>
  )
}

function statusTone(status: string) {
  switch (status) {
    case 'CONFIRMED':
    case 'COMPLETED':
      return 'success'
    case 'CREATED':
      return 'info'
    case 'NO_SHOW':
      return 'warning'
    case 'CANCELLED':
      return 'danger'
    default:
      return 'neutral'
  }
}

function riskTone(level: string) {
  switch (level) {
    case 'HIGH':
      return 'danger'
    case 'MEDIUM':
      return 'warning'
    default:
      return 'success'
  }
}

function isToday(value: string) {
  const date = new Date(value)
  const today = new Date()
  return date.getFullYear() === today.getFullYear()
    && date.getMonth() === today.getMonth()
    && date.getDate() === today.getDate()
}
