import type { UseQueryResult } from '@tanstack/react-query'
import type { FormEvent } from 'react'
import { useMemo, useState } from 'react'
import { CalendarDays, ChevronLeft, ChevronRight, Clock3, Plus } from 'lucide-react'
import type {
  CreateSchedulePayload,
  CreateTimeSlotPayload,
  Doctor,
  Schedule,
  TimeSlot,
  UpdateSchedulePayload,
  UpdateTimeSlotPayload,
} from '../../shared/api/types'
import { scheduleStatusLabel, slotStatusLabel } from '../../shared/lib/labels'
import { isBookableTimeSlot } from '../../shared/lib/timeSlots'
import { Button } from '../../shared/ui/Button'
import { ContentState } from '../../shared/ui/ContentState'
import { OperationFeedback } from '../../shared/ui/OperationFeedback'
import { AdminTable } from './AdminTable'
import { ScheduleCsvImport } from './ScheduleCsvImport'
import type { ActionResult } from './types'
import { formBoolean, formDate, formDateTime, formString } from './form'

type SchedulesSectionProps = {
  doctors: Doctor[]
  selectedDoctorId: string
  isPending: boolean
  schedulesQuery: UseQueryResult<{ items: Schedule[] }>
  onSelectDoctor: (doctorId: string) => void
  onCreateSchedule: (payload: CreateSchedulePayload) => void
  onUpdateSchedule: (scheduleId: string, payload: UpdateSchedulePayload) => void
  onCreateTimeSlot: (scheduleId: string, payload: CreateTimeSlotPayload) => void
  onUpdateTimeSlot: (slotId: string, payload: UpdateTimeSlotPayload) => void
  onBlockTimeSlot: (slotId: string, reason: string) => void
  onUnblockTimeSlot: (slotId: string) => void
  onUseTimeSlot: (slotId: string) => void
  canUseTimeSlot: boolean
  actionResult?: ActionResult | null
  onDismissResult?: () => void
}

export function SchedulesSection({
  doctors,
  selectedDoctorId,
  isPending,
  schedulesQuery,
  onSelectDoctor,
  onCreateSchedule,
  onUpdateSchedule,
  onCreateTimeSlot,
  onUpdateTimeSlot,
  onBlockTimeSlot,
  onUnblockTimeSlot,
  onUseTimeSlot,
  canUseTimeSlot,
  actionResult = null,
  onDismissResult = () => undefined,
}: SchedulesSectionProps) {
  const [today] = useState(() => new Date())
  const [calendarAnchor, setCalendarAnchor] = useState(() => startOfWeek(today))
  const [selectedSlotId, setSelectedSlotId] = useState('')
  const [slotDraftDate, setSlotDraftDate] = useState(() => localDateInput(today))
  const [blockTargetId, setBlockTargetId] = useState('')
  const defaultStartDate = localDateInput(new Date())
  const defaultEndDate = localDateInput(addDays(new Date(), 14))

  function createSchedule(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreateSchedule({
      doctor_id: formString(form, 'doctor_id', selectedDoctorId),
      start_date: formDate(form, 'start_date'),
      end_date: formDate(form, 'end_date'),
      working_days: formString(form, 'working_days', '1,2,3,4,5'),
      work_start_time: formString(form, 'work_start_time', '09:00'),
      work_end_time: formString(form, 'work_end_time', '18:00'),
      status: formString(form, 'status', 'ACTIVE'),
    })
  }

  function updateSchedule(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onUpdateSchedule(formString(form, 'schedule_id'), {
      start_date: formDate(form, 'start_date'),
      end_date: formDate(form, 'end_date'),
      working_days: formString(form, 'working_days'),
      work_start_time: formString(form, 'work_start_time'),
      work_end_time: formString(form, 'work_end_time'),
      status: formString(form, 'status'),
    })
  }

  const schedules = schedulesQuery.data?.items ?? []
  const slots = schedules.flatMap((schedule) => schedule.time_slots ?? [])
  const selectedDoctor = doctors.find((doctor) => doctor.id === selectedDoctorId)
  const weekDays = useMemo(
    () => createWeekDays(calendarAnchor, slots, today),
    [calendarAnchor, slots, today],
  )
  const weekSlots = weekDays.flatMap((day) => day.slots)
  const selectedSlot = slots.find((slot) => slot.id === selectedSlotId)
  const blockTarget = slots.find((slot) => slot.id === blockTargetId)

  function confirmBlock(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!blockTargetId) return
    const form = new FormData(event.currentTarget)
    const reason = formString(form, 'reason')
    if (!reason) return
    onBlockTimeSlot(blockTargetId, reason)
    setBlockTargetId('')
  }

  return (
    <>
    <OperationFeedback result={actionResult} onDismiss={onDismissResult} />
    <WeeklyScheduleWorkspace
      doctors={doctors}
      doctor={selectedDoctor}
      selectedDoctorId={selectedDoctorId}
      schedules={schedules}
      weekDays={weekDays}
      weekSlots={weekSlots}
      calendarAnchor={calendarAnchor}
      bookingCutoff={today}
      isLoading={schedulesQuery.isFetching}
      isPending={isPending}
      canUseTimeSlot={canUseTimeSlot}
      onSelectDoctor={onSelectDoctor}
      onMoveWeek={(days) => setCalendarAnchor((current) => addDays(current, days))}
      onCurrentWeek={() => setCalendarAnchor(startOfWeek(today))}
      onUseTimeSlot={onUseTimeSlot}
      onBlockTimeSlot={setBlockTargetId}
      onUnblockTimeSlot={onUnblockTimeSlot}
      onSelectSlot={setSelectedSlotId}
      onCreateForDay={(date) => {
        setSelectedSlotId('')
        setSlotDraftDate(localDateInput(date))
      }}
    />

    {blockTarget ? (
      <form className="slot-block-confirm" onSubmit={confirmBlock}>
        <div>
          <strong>Заблокировать слот {formatSlot(blockTarget)}?</strong>
          <span>Он станет недоступен для новых записей до ручной разблокировки.</span>
        </div>
        <textarea name="reason" placeholder="Причина блокировки" required autoFocus />
        <div className="form-actions">
          <Button type="submit" variant="danger" disabled={isPending}>Заблокировать</Button>
          <Button type="button" variant="ghost" disabled={isPending} onClick={() => setBlockTargetId('')}>
            Назад
          </Button>
        </div>
      </form>
    ) : null}

    <SlotWorkbench
      schedules={schedules}
      slot={selectedSlot}
      draftDate={slotDraftDate}
      bookingCutoff={today}
      isPending={isPending}
      canUseTimeSlot={canUseTimeSlot}
      onSelectSlot={setSelectedSlotId}
      onCreateTimeSlot={onCreateTimeSlot}
      onUpdateTimeSlot={onUpdateTimeSlot}
      onBlockTimeSlot={setBlockTargetId}
      onUnblockTimeSlot={onUnblockTimeSlot}
      onUseTimeSlot={onUseTimeSlot}
    />

    <ScheduleCsvImport />

    <section className="admin-two-column">
      <details className="panel workspace-form-disclosure">
        <summary>
          <Plus size={18} />
          <div>
            <p className="eyebrow">Расписание</p>
            <strong>Добавить расписание</strong>
          </div>
        </summary>
        <form className="admin-form" onSubmit={createSchedule}>
          <select
            name="doctor_id"
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
          <input name="start_date" type="date" defaultValue={defaultStartDate} required />
          <input name="end_date" type="date" defaultValue={defaultEndDate} required />
          <input name="working_days" placeholder="Рабочие дни: 1,2,3,4,5" defaultValue="1,2,3,4,5" />
          <input name="work_start_time" type="time" defaultValue="09:00" />
          <input name="work_end_time" type="time" defaultValue="18:00" />
          <select name="status" defaultValue="ACTIVE">
            <option value="ACTIVE">{scheduleStatusLabel('ACTIVE')}</option>
            <option value="INACTIVE">{scheduleStatusLabel('INACTIVE')}</option>
          </select>
          <Button type="submit" disabled={isPending || !selectedDoctorId}>Создать расписание</Button>
        </form>
      </details>

      <details className="panel workspace-form-disclosure">
        <summary>
          <CalendarDays size={18} />
          <div>
            <p className="eyebrow">Расписание</p>
            <strong>Изменить расписание</strong>
          </div>
        </summary>
        <form className="admin-form" onSubmit={updateSchedule}>
          <select name="schedule_id" required>
            <option value="">Выбери расписание</option>
            {schedules.map((schedule) => (
              <option key={schedule.id} value={schedule.id}>
                {formatDate(schedule.start_date)} - {formatDate(schedule.end_date)}
              </option>
            ))}
          </select>
          <input name="start_date" type="date" required />
          <input name="end_date" type="date" required />
          <input name="working_days" placeholder="Рабочие дни" defaultValue="1,2,3,4,5" />
          <input name="work_start_time" type="time" defaultValue="09:00" />
          <input name="work_end_time" type="time" defaultValue="18:00" />
          <select name="status" defaultValue="ACTIVE">
            <option value="ACTIVE">{scheduleStatusLabel('ACTIVE')}</option>
            <option value="INACTIVE">{scheduleStatusLabel('INACTIVE')}</option>
          </select>
          <Button type="submit" disabled={isPending}>Сохранить расписание</Button>
        </form>
      </details>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Список</p>
            <h2>Расписания врача</h2>
          </div>
        </div>
        {schedulesQuery.isFetching ? <ContentState variant="loading" title="Загружаю расписания" compact /> : null}
        {schedulesQuery.isError ? (
          <ContentState variant="error" title="Не удалось загрузить расписания" description={schedulesQuery.error.message} compact />
        ) : null}
        <AdminTable
          items={schedules}
          empty={selectedDoctorId ? 'Расписаний пока нет.' : 'Выбери врача, чтобы загрузить расписания.'}
          columns={[
            { key: 'dates', title: 'Период', render: (schedule) => `${new Date(schedule.start_date).toLocaleDateString('ru-RU')} - ${new Date(schedule.end_date).toLocaleDateString('ru-RU')}` },
            { key: 'work', title: 'Время', render: (schedule) => `${schedule.work_start_time} - ${schedule.work_end_time}` },
            { key: 'status', title: 'Статус', render: (schedule) => scheduleStatusLabel(schedule.status) },
          ]}
        />
      </article>

    </section>
    </>
  )
}

type ScheduleWeekDay = {
  key: string
  date: Date
  label: string
  isToday: boolean
  slots: TimeSlot[]
}

function WeeklyScheduleWorkspace({
  doctors,
  doctor,
  selectedDoctorId,
  schedules,
  weekDays,
  weekSlots,
  calendarAnchor,
  bookingCutoff,
  isLoading,
  isPending,
  canUseTimeSlot,
  onSelectDoctor,
  onMoveWeek,
  onCurrentWeek,
  onUseTimeSlot,
  onBlockTimeSlot,
  onUnblockTimeSlot,
  onSelectSlot,
  onCreateForDay,
}: {
  doctors: Doctor[]
  doctor?: Doctor
  selectedDoctorId: string
  schedules: Schedule[]
  weekDays: ScheduleWeekDay[]
  weekSlots: TimeSlot[]
  calendarAnchor: Date
  bookingCutoff: Date
  isLoading: boolean
  isPending: boolean
  canUseTimeSlot: boolean
  onSelectDoctor: (doctorId: string) => void
  onMoveWeek: (days: number) => void
  onCurrentWeek: () => void
  onUseTimeSlot: (slotId: string) => void
  onBlockTimeSlot: (slotId: string) => void
  onUnblockTimeSlot: (slotId: string) => void
  onSelectSlot: (slotId: string) => void
  onCreateForDay: (date: Date) => void
}) {
  const freeCount = weekSlots.filter((slot) => slot.status === 'FREE' && !slot.is_blocked).length
  const bookedCount = weekSlots.filter((slot) => slot.status === 'BOOKED').length
  const unavailableCount = weekSlots.filter((slot) => slot.is_blocked || slot.status === 'BLOCKED' || slot.status === 'BREAK').length

  return (
    <section className="schedule-workspace">
      <div className="schedule-workspace-header">
        <div>
          <p className="eyebrow">Рабочее расписание</p>
          <h2>{doctor ? `${doctor.full_name} · ${doctor.specialty}` : 'Недельная сетка врача'}</h2>
        </div>
        <div className="schedule-week-controls">
          <Button
            type="button"
            variant="secondary"
            icon={<ChevronLeft size={16} />}
            onClick={() => onMoveWeek(-7)}
          >
            Назад
          </Button>
          <Button type="button" variant="secondary" onClick={onCurrentWeek}>
            Текущая неделя
          </Button>
          <Button
            type="button"
            variant="secondary"
            icon={<ChevronRight size={16} />}
            onClick={() => onMoveWeek(7)}
          >
            Вперед
          </Button>
        </div>
      </div>

      <div className="schedule-workspace-toolbar">
        <select value={selectedDoctorId} onChange={(event) => onSelectDoctor(event.target.value)}>
          <option value="">Выбери врача</option>
          {doctors.map((item) => (
            <option key={item.id} value={item.id}>
              {item.full_name} · {item.specialty}
            </option>
          ))}
        </select>
        <div className="schedule-period">
          <CalendarDays size={17} />
          <strong>{formatWeekRange(calendarAnchor)}</strong>
        </div>
      </div>

      <div className="schedule-summary">
        <ScheduleSummary label="Расписаний" value={schedules.length} />
        <ScheduleSummary label="Слотов на неделе" value={weekSlots.length} />
        <ScheduleSummary label="Свободно" value={freeCount} tone="success" />
        <ScheduleSummary label="Занято" value={bookedCount} tone="info" />
        <ScheduleSummary label="Недоступно" value={unavailableCount} tone="warning" />
      </div>

      {!selectedDoctorId ? (
        <ContentState variant="empty" title="Врач не выбран" description="Выбери врача, чтобы загрузить расписание и слоты." />
      ) : isLoading ? (
        <ContentState variant="loading" title="Загружаю расписание врача" />
      ) : (
        <div className="schedule-week-grid">
          {weekDays.map((day) => (
            <section className={day.isToday ? 'schedule-week-day today' : 'schedule-week-day'} key={day.key}>
              <header>
                <span>{day.label}</span>
                <strong>{day.date.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })}</strong>
                <small>{day.slots.length ? `${day.slots.length} слотов` : 'Нет слотов'}</small>
              </header>
              <div className="schedule-week-slots">
                {day.slots.map((slot) => (
                  <article className={`schedule-slot ${slotClass(slot)}`} key={slot.id}>
                    <div>
                      <strong>{formatSlotTime(slot)}</strong>
                      <span>{slotStatusLabel(slot.status)}{slot.is_blocked ? ' · заблокирован' : ''}</span>
                    </div>
                    <div className="slot-card-actions">
                      <Button type="button" variant="ghost" onClick={() => onSelectSlot(slot.id)}>
                        Детали
                      </Button>
                      {canUseTimeSlot && isBookableTimeSlot(slot, bookingCutoff) ? (
                        <Button type="button" variant="secondary" onClick={() => onUseTimeSlot(slot.id)}>
                          В запись
                        </Button>
                      ) : null}
                      {slot.is_blocked ? (
                        <Button type="button" variant="secondary" disabled={isPending} onClick={() => onUnblockTimeSlot(slot.id)}>
                          Разблокировать
                        </Button>
                      ) : slot.status !== 'BOOKED' ? (
                        <Button type="button" variant="secondary" disabled={isPending} onClick={() => onBlockTimeSlot(slot.id)}>
                          Блокировать
                        </Button>
                      ) : null}
                    </div>
                  </article>
                ))}
                {!day.slots.length ? (
                  <div className="schedule-day-empty">
                    <Clock3 size={17} />
                    <span>Приемов нет</span>
                    <Button type="button" variant="ghost" onClick={() => onCreateForDay(day.date)}>
                      Добавить слот
                    </Button>
                  </div>
                ) : (
                  <Button type="button" variant="ghost" onClick={() => onCreateForDay(day.date)}>
                    Добавить время
                  </Button>
                )}
              </div>
            </section>
          ))}
        </div>
      )}
    </section>
  )
}

function SlotWorkbench({
  schedules,
  slot,
  draftDate,
  bookingCutoff,
  isPending,
  canUseTimeSlot,
  onSelectSlot,
  onCreateTimeSlot,
  onUpdateTimeSlot,
  onBlockTimeSlot,
  onUnblockTimeSlot,
  onUseTimeSlot,
}: {
  schedules: Schedule[]
  slot?: TimeSlot
  draftDate: string
  bookingCutoff: Date
  isPending: boolean
  canUseTimeSlot: boolean
  onSelectSlot: (slotId: string) => void
  onCreateTimeSlot: (scheduleId: string, payload: CreateTimeSlotPayload) => void
  onUpdateTimeSlot: (slotId: string, payload: UpdateTimeSlotPayload) => void
  onBlockTimeSlot: (slotId: string) => void
  onUnblockTimeSlot: (slotId: string) => void
  onUseTimeSlot: (slotId: string) => void
}) {
  const slots = schedules.flatMap((schedule) => schedule.time_slots ?? [])
  const defaultStart = `${draftDate}T09:00`
  const defaultEnd = `${draftDate}T09:30`

  function createSlot(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreateTimeSlot(formString(form, 'schedule_id'), {
      start_at: formDateTime(form, 'start_at'),
      end_at: formDateTime(form, 'end_at'),
      status: formString(form, 'status', 'FREE') as CreateTimeSlotPayload['status'],
      is_blocked: formBoolean(form, 'is_blocked'),
    })
  }

  function updateSlot(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!slot) {
      return
    }
    const form = new FormData(event.currentTarget)
    onUpdateTimeSlot(slot.id, {
      start_at: formDateTime(form, 'start_at'),
      end_at: formDateTime(form, 'end_at'),
      status: formString(form, 'status', slot.status) as UpdateTimeSlotPayload['status'],
      is_blocked: slot.is_blocked,
    })
  }

  return (
    <section className="slot-workbench">
      <details className="workspace-form-disclosure slot-create-disclosure">
        <summary>
          <Plus size={18} />
          <div>
            <p className="eyebrow">Новое время</p>
            <strong>Добавить слот</strong>
          </div>
        </summary>
        <form className="admin-form" key={draftDate} onSubmit={createSlot}>
          <select name="schedule_id" required>
            <option value="">Выбери расписание</option>
            {schedules.map((schedule) => (
              <option key={schedule.id} value={schedule.id}>
                {formatDate(schedule.start_date)} - {formatDate(schedule.end_date)} · {schedule.work_start_time}-{schedule.work_end_time}
              </option>
            ))}
          </select>
          <div className="slot-time-fields">
            <input name="start_at" type="datetime-local" defaultValue={defaultStart} required />
            <input name="end_at" type="datetime-local" defaultValue={defaultEnd} required />
          </div>
          <select name="status" defaultValue="FREE">
            <option value="FREE">{slotStatusLabel('FREE')}</option>
            <option value="BREAK">{slotStatusLabel('BREAK')}</option>
          </select>
          <Button type="submit" disabled={isPending || !schedules.length}>Добавить в расписание</Button>
        </form>
      </details>

      <article>
        <div className="panel-header">
          <div>
            <p className="eyebrow">Выбранный слот</p>
            <h2>{slot ? formatSlot(slot) : 'Детали времени'}</h2>
          </div>
          {slot ? <span className={`workspace-badge ${slotBadgeTone(slot)}`}>{slotStatusLabel(slot.status)}</span> : null}
        </div>
        <select value={slot?.id ?? ''} onChange={(event) => onSelectSlot(event.target.value)}>
          <option value="">Выбери слот из сетки</option>
          {slots.map((item) => (
            <option key={item.id} value={item.id}>
              {formatSlot(item)} · {slotStatusLabel(item.status)}
            </option>
          ))}
        </select>
        {!slot ? (
          <ContentState variant="empty" title="Слот не выбран" description="Нажми «Детали» на слоте в недельной сетке." compact />
        ) : (
          <form className="admin-form" key={slot.id} onSubmit={updateSlot}>
            <div className="entity-id-line">
              <span>ID слота</span>
              <code>{slot.id}</code>
            </div>
            <div className="slot-time-fields">
              <input name="start_at" type="datetime-local" defaultValue={localDateTimeInput(new Date(slot.start_at))} required />
              <input name="end_at" type="datetime-local" defaultValue={localDateTimeInput(new Date(slot.end_at))} required />
            </div>
            <select name="status" defaultValue={slot.status}>
              <option value="FREE">{slotStatusLabel('FREE')}</option>
              <option value="BOOKED">{slotStatusLabel('BOOKED')}</option>
              <option value="BREAK">{slotStatusLabel('BREAK')}</option>
              {slot.status === 'BLOCKED' ? (
                <option value="BLOCKED">{slotStatusLabel('BLOCKED')}</option>
              ) : null}
            </select>
            {slot.is_blocked && slot.block_reason ? (
              <p className="slot-block-reason"><strong>Причина блокировки:</strong> {slot.block_reason}</p>
            ) : null}
            <div className="form-actions">
              <Button type="submit" disabled={isPending}>Сохранить</Button>
              {slot.is_blocked ? (
                <Button type="button" variant="secondary" disabled={isPending} onClick={() => onUnblockTimeSlot(slot.id)}>
                  Разблокировать
                </Button>
              ) : slot.status !== 'BOOKED' ? (
                <Button type="button" variant="secondary" disabled={isPending} onClick={() => onBlockTimeSlot(slot.id)}>
                  Заблокировать
                </Button>
              ) : null}
              {canUseTimeSlot && isBookableTimeSlot(slot, bookingCutoff) ? (
                <Button type="button" variant="secondary" onClick={() => onUseTimeSlot(slot.id)}>
                  Оформить запись
                </Button>
              ) : null}
            </div>
          </form>
        )}
      </article>
    </section>
  )
}

function ScheduleSummary({
  label,
  value,
  tone = 'neutral',
}: {
  label: string
  value: number
  tone?: 'neutral' | 'success' | 'info' | 'warning'
}) {
  return (
    <div>
      <span>{label}</span>
      <strong className={tone}>{value}</strong>
    </div>
  )
}

function formatSlot(slot: TimeSlot) {
  return `${formatDateTime(slot.start_at)} - ${new Date(slot.end_at).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })}`
}

function formatSlotTime(slot: TimeSlot) {
  return `${new Date(slot.start_at).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })} - ${new Date(slot.end_at).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })}`
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString('ru-RU')
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function addDays(date: Date, days: number) {
  const next = new Date(date)
  next.setDate(next.getDate() + days)
  return next
}

function createWeekDays(anchor: Date, slots: TimeSlot[], today: Date): ScheduleWeekDay[] {
  return Array.from({ length: 7 }, (_, index) => {
    const date = addDays(anchor, index)
    const key = localDateInput(date)
    return {
      key,
      date,
      label: date.toLocaleDateString('ru-RU', { weekday: 'short' }),
      isToday: isSameLocalDate(date, today),
      slots: slots
        .filter((slot) => localDateInput(new Date(slot.start_at)) === key)
        .sort((left, right) => new Date(left.start_at).getTime() - new Date(right.start_at).getTime()),
    }
  })
}

function startOfWeek(date: Date) {
  const start = new Date(date)
  const day = start.getDay()
  const distanceFromMonday = day === 0 ? 6 : day - 1
  start.setDate(start.getDate() - distanceFromMonday)
  start.setHours(0, 0, 0, 0)
  return start
}

function isSameLocalDate(left: Date, right: Date) {
  return left.getFullYear() === right.getFullYear()
    && left.getMonth() === right.getMonth()
    && left.getDate() === right.getDate()
}

function formatWeekRange(anchor: Date) {
  const end = addDays(anchor, 6)
  return `${anchor.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short' })} - ${end.toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })}`
}

function slotClass(slot: TimeSlot) {
  if (slot.is_blocked) {
    return 'blocked'
  }
  return slot.status.toLowerCase()
}

function slotBadgeTone(slot: TimeSlot) {
  if (slot.is_blocked || slot.status === 'BLOCKED') {
    return 'danger'
  }
  if (slot.status === 'FREE') {
    return 'success'
  }
  if (slot.status === 'BOOKED') {
    return 'info'
  }
  return 'warning'
}

function localDateInput(date: Date) {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}

function localDateTimeInput(date: Date) {
  return `${localDateInput(date)}T${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}
