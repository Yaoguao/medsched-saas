import { useQuery } from '@tanstack/react-query'
import type { FormEvent } from 'react'
import { useMemo, useState } from 'react'
import { BriefcaseMedical, CalendarDays, Link2, Plane, Plus, Stethoscope } from 'lucide-react'
import type {
  CreateStaffUserPayload,
  CreateDoctorPayload,
  Doctor,
  LinkDoctorUserPayload,
  TimeSlot,
  UpdateDoctorPayload,
  User,
} from '../../shared/api/types'
import { coreApi } from '../../shared/api/core'
import { doctorStatusLabel, slotStatusLabel, userRoleLabel } from '../../shared/lib/labels'
import { Button } from '../../shared/ui/Button'
import { ContentState } from '../../shared/ui/ContentState'
import { MetricCard } from '../../shared/ui/MetricCard'
import { OperationFeedback } from '../../shared/ui/OperationFeedback'
import { AdminTable } from './AdminTable'
import type { ActionResult } from './types'
import { formNumber, formOptionalNumber, formOptionalString, formString } from './form'

type DoctorsSectionProps = {
  doctors: Doctor[]
  users: User[]
  initialDoctorId?: string
  actionResult?: ActionResult | null
  onDismissResult?: () => void
  isPending: boolean
  canManageStaff: boolean
  onCreateStaffUser: (payload: CreateStaffUserPayload) => void
  onCreateDoctor: (payload: CreateDoctorPayload) => void
  onUpdateDoctor: (doctorId: string, payload: UpdateDoctorPayload) => void
  onLinkDoctorUser: (doctorId: string, payload: LinkDoctorUserPayload) => void
  onOpenSchedule: (doctorId: string) => void
}

export function DoctorsSection({
  doctors,
  users,
  initialDoctorId = '',
  actionResult = null,
  onDismissResult = () => undefined,
  isPending,
  canManageStaff,
  onCreateStaffUser,
  onCreateDoctor,
  onUpdateDoctor,
  onLinkDoctorUser,
  onOpenSchedule,
}: DoctorsSectionProps) {
  const [query, setQuery] = useState('')
  const [selectedDoctorId, setSelectedDoctorId] = useState(initialDoctorId)
  const selectedDoctorQuery = useQuery({
    queryKey: ['doctor', selectedDoctorId],
    queryFn: () => coreApi.doctor(selectedDoctorId),
    enabled: Boolean(selectedDoctorId),
  })
  const selectedDoctorSchedulesQuery = useQuery({
    queryKey: ['doctor-schedules', selectedDoctorId],
    queryFn: () => coreApi.doctorSchedules(selectedDoctorId, { limit: 50 }),
    enabled: Boolean(selectedDoctorId),
  })

  const filteredDoctors = useMemo(() => {
    const search = query.trim().toLowerCase()
    if (!search) {
      return doctors
    }
    return doctors.filter((doctor) =>
      [doctor.full_name, doctor.specialty, doctor.cabinet, doctor.status, doctor.id, doctor.user_id]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)),
    )
  }, [doctors, query])

  function createDoctor(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreateDoctor({
      full_name: formString(form, 'full_name'),
      specialty: formString(form, 'specialty'),
      cabinet: formString(form, 'cabinet'),
      experience_years: formNumber(form, 'experience_years'),
      status: formString(form, 'status', 'ACTIVE'),
    })
  }

  function createStaffUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreateStaffUser({
      role: formString(form, 'role', 'DOCTOR') as CreateStaffUserPayload['role'],
      email: formString(form, 'email'),
      phone: formString(form, 'phone'),
      password: formString(form, 'password'),
    })
    event.currentTarget.reset()
  }

  function linkDoctorUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onLinkDoctorUser(formString(form, 'doctor_id', selectedDoctorId), {
      user_id: formString(form, 'user_id'),
    })
  }

  function updateDoctor(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onUpdateDoctor(formString(form, 'doctor_id', selectedDoctorId), {
      full_name: formOptionalString(form, 'full_name'),
      specialty: formOptionalString(form, 'specialty'),
      cabinet: formOptionalString(form, 'cabinet'),
      experience_years: formOptionalNumber(form, 'experience_years'),
      status: formOptionalString(form, 'status'),
    })
  }

  const unlinkedDoctors = doctors.filter((doctor) => !doctor.user_id)
  const linkableUsers = users.filter((user) => user.role !== 'PATIENT' && !user.doctor_id)
  const activeCount = doctors.filter((doctor) => doctor.status === 'ACTIVE').length
  const linkedCount = doctors.filter((doctor) => doctor.user_id).length
  const vacationCount = doctors.filter((doctor) => doctor.status === 'VACATION').length
  const selectedDoctor = selectedDoctorQuery.data ?? doctors.find((doctor) => doctor.id === selectedDoctorId)
  const selectedDoctorUser = users.find((user) => user.id === selectedDoctor?.user_id)
  const selectedDoctorSchedules = selectedDoctorSchedulesQuery.data?.items ?? []
  const selectedDoctorSlots = selectedDoctorSchedules.flatMap((schedule) => schedule.time_slots ?? [])
  const linkDoctorTargetId = selectedDoctor && !selectedDoctor.user_id ? selectedDoctor.id : ''

  return (
    <>
    <OperationFeedback result={actionResult} onDismiss={onDismissResult} />
    <DoctorWorkbench
      doctors={filteredDoctors}
      doctor={selectedDoctor}
      user={selectedDoctorUser}
      schedulesCount={selectedDoctorSchedules.length}
      slots={selectedDoctorSlots}
      selectedDoctorId={selectedDoctorId}
      isLoading={selectedDoctorQuery.isFetching || selectedDoctorSchedulesQuery.isFetching}
      onSelectDoctor={setSelectedDoctorId}
      onOpenSchedule={onOpenSchedule}
    />

    <section className="metrics-grid">
      <MetricCard label="Врачи" value={doctors.length} hint="в справочнике" tone="blue" icon={<Stethoscope size={22} />} />
      <MetricCard label="Активны" value={activeCount} hint="доступны для расписаний" tone="green" icon={<BriefcaseMedical size={22} />} />
      <MetricCard label="Аккаунты" value={linkedCount} hint="привязаны к врачам" tone="violet" icon={<Link2 size={22} />} />
      <MetricCard label="В отпуске" value={vacationCount} hint="недоступны для части расписаний" tone="orange" icon={<Plane size={22} />} />
    </section>

    <section className="admin-two-column">
      {canManageStaff ? (
        <details className="panel workspace-form-disclosure">
          <summary>
            <Plus size={18} />
            <div>
              <p className="eyebrow">Аккаунт</p>
              <strong>Добавить аккаунт сотрудника</strong>
            </div>
          </summary>
          <form className="admin-form" onSubmit={createStaffUser}>
            <select name="role" defaultValue="DOCTOR" required>
              <option value="DOCTOR">{userRoleLabel('DOCTOR')}</option>
              <option value="REGISTRATOR">{userRoleLabel('REGISTRATOR')}</option>
              <option value="MANAGER">{userRoleLabel('MANAGER')}</option>
              <option value="IT_SUPPORT">{userRoleLabel('IT_SUPPORT')}</option>
              <option value="ADMIN">{userRoleLabel('ADMIN')}</option>
            </select>
            <input name="email" type="email" placeholder="doctor1@example.com" required />
            <input name="phone" placeholder="+79990000888" required />
            <input name="password" type="password" placeholder="Пароль от 8 символов" required minLength={8} />
            <Button type="submit" disabled={isPending}>Создать аккаунт</Button>
          </form>
        </details>
      ) : null}

      <details className="panel workspace-form-disclosure">
        <summary>
          <Plus size={18} />
          <div>
            <p className="eyebrow">Врачи</p>
            <strong>Добавить врача</strong>
          </div>
        </summary>
        <form className="admin-form" onSubmit={createDoctor}>
          <input name="full_name" placeholder="ФИО врача" required />
          <input name="specialty" placeholder="Специальность" required />
          <input name="cabinet" placeholder="Кабинет" />
          <input name="experience_years" type="number" min="0" placeholder="Стаж" defaultValue="0" />
          <select name="status" defaultValue="ACTIVE">
            <option value="ACTIVE">{doctorStatusLabel('ACTIVE')}</option>
            <option value="INACTIVE">{doctorStatusLabel('INACTIVE')}</option>
            <option value="VACATION">{doctorStatusLabel('VACATION')}</option>
          </select>
          <Button type="submit" disabled={isPending}>Создать врача</Button>
        </form>
      </details>

      {canManageStaff ? (
      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Аккаунт</p>
            <h2>Привязать аккаунт к врачу</h2>
          </div>
        </div>
        <form className="admin-form" onSubmit={linkDoctorUser}>
          <select name="doctor_id" value={linkDoctorTargetId} onChange={(event) => setSelectedDoctorId(event.target.value)} required>
            <option value="">Выбери врача без аккаунта</option>
            {unlinkedDoctors.map((doctor) => (
              <option key={doctor.id} value={doctor.id}>
                {doctor.full_name} · {doctor.specialty}
              </option>
            ))}
          </select>
          <select name="user_id" required>
            <option value="">Выбери аккаунт сотрудника без врача</option>
            {linkableUsers.map((user) => (
              <option key={user.id} value={user.id}>
                {user.email} · {userRoleLabel(user.role)}
              </option>
            ))}
          </select>
          <Button type="submit" disabled={isPending || !linkDoctorTargetId}>Привязать и выдать роль врача</Button>
        </form>
        <p className="section-note">
          Для врача сначала создай аккаунт сотрудника, затем профиль врача и привяжи их здесь.
        </p>
      </article>
      ) : null}

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Врач</p>
            <h2>Редактировать врача</h2>
          </div>
        </div>
        <form className="admin-form" onSubmit={updateDoctor} key={selectedDoctor?.id ?? 'doctor-edit-empty'}>
          <select name="doctor_id" value={selectedDoctorId} onChange={(event) => setSelectedDoctorId(event.target.value)} required>
            <option value="">Выбери врача</option>
            {doctors.map((doctor) => (
              <option key={doctor.id} value={doctor.id}>
                {doctor.full_name} · {doctor.specialty}
              </option>
            ))}
          </select>
          <input name="full_name" defaultValue={selectedDoctor?.full_name ?? ''} placeholder="ФИО" />
          <input name="specialty" defaultValue={selectedDoctor?.specialty ?? ''} placeholder="Специальность" />
          <input name="cabinet" defaultValue={selectedDoctor?.cabinet ?? ''} placeholder="Кабинет" />
          <input name="experience_years" type="number" min="0" defaultValue={selectedDoctor?.experience_years ?? 0} placeholder="Стаж" />
          <select name="status" defaultValue={selectedDoctor?.status ?? 'ACTIVE'}>
            <option value="ACTIVE">{doctorStatusLabel('ACTIVE')}</option>
            <option value="INACTIVE">{doctorStatusLabel('INACTIVE')}</option>
            <option value="VACATION">{doctorStatusLabel('VACATION')}</option>
          </select>
          <Button type="submit" disabled={isPending || !selectedDoctorId}>Сохранить врача</Button>
        </form>
      </article>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Список</p>
            <h2>Врачи</h2>
          </div>
          <input
            className="table-filter"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Поиск по ФИО, специальности, статусу"
          />
        </div>
        <AdminTable
          items={filteredDoctors}
          empty="Врачей пока нет."
          columns={[
            { key: 'name', title: 'ФИО', render: (doctor) => doctor.full_name },
            { key: 'specialty', title: 'Специальность', render: (doctor) => doctor.specialty },
            { key: 'cabinet', title: 'Кабинет', render: (doctor) => doctor.cabinet },
            { key: 'experience', title: 'Стаж', render: (doctor) => `${doctor.experience_years} лет` },
            { key: 'status', title: 'Статус', render: (doctor) => <DoctorStatusBadge status={doctor.status} /> },
            { key: 'user', title: 'Аккаунт', render: (doctor) => doctor.user_id ? <span className="workspace-badge success">Привязан</span> : <span className="workspace-badge neutral">Не привязан</span> },
            { key: 'actions', title: 'Действия', render: (doctor) => (
              <Button type="button" variant="secondary" onClick={() => setSelectedDoctorId(doctor.id)}>
                Открыть
              </Button>
            ) },
          ]}
        />
      </article>
    </section>
    </>
  )
}

function DoctorWorkbench({
  doctors,
  doctor,
  user,
  schedulesCount,
  slots,
  selectedDoctorId,
  isLoading,
  onSelectDoctor,
  onOpenSchedule,
}: {
  doctors: Doctor[]
  doctor?: Doctor
  user?: User
  schedulesCount: number
  slots: TimeSlot[]
  selectedDoctorId: string
  isLoading: boolean
  onSelectDoctor: (doctorId: string) => void
  onOpenSchedule: (doctorId: string) => void
}) {
  const [openedAt] = useState(() => Date.now())
  const upcomingSlots = slots
    .filter((slot) => new Date(slot.start_at).getTime() >= openedAt)
    .sort((left, right) => new Date(left.start_at).getTime() - new Date(right.start_at).getTime())
    .slice(0, 5)
  const freeCount = slots.filter((slot) => slot.status === 'FREE' && !slot.is_blocked).length
  const bookedCount = slots.filter((slot) => slot.status === 'BOOKED').length

  return (
    <section className="doctor-workbench">
      <div className="doctor-workbench-header">
        <div>
          <p className="eyebrow">Карточка врача</p>
          <h2>Профиль, аккаунт и расписание</h2>
        </div>
        {doctor ? <DoctorStatusBadge status={doctor.status} /> : null}
      </div>

      <select value={selectedDoctorId} onChange={(event) => onSelectDoctor(event.target.value)}>
        <option value="">Выбери врача</option>
        {doctors.map((item) => (
          <option key={item.id} value={item.id}>
            {item.full_name} · {item.specialty}
          </option>
        ))}
      </select>

      {!doctor ? (
        isLoading ? (
          <ContentState variant="loading" title="Загружаю профиль врача" compact />
        ) : (
          <ContentState
            variant="empty"
            title="Врач не выбран"
            description="Открой врача из списка ниже, чтобы увидеть профиль и расписание."
            compact
          />
        )
      ) : (
        <>
          <div className="doctor-workbench-grid">
            <DoctorLine label="ID врача" value={doctor.id} />
            <DoctorLine label="ФИО" value={doctor.full_name} />
            <DoctorLine label="Специальность" value={doctor.specialty} />
            <DoctorLine label="Кабинет" value={doctor.cabinet || 'Не указан'} />
            <DoctorLine label="Стаж" value={`${doctor.experience_years} лет`} />
            <DoctorLine label="Аккаунт" value={user?.email || (doctor.user_id ? 'Привязан' : 'Не привязан')} />
            <DoctorLine label="Телефон аккаунта" value={user?.phone || 'Не указан'} />
            <DoctorLine label="Расписаний" value={`${schedulesCount}`} />
            <DoctorLine label="Слоты" value={`${freeCount} свободно · ${bookedCount} занято`} />
          </div>

          <div className="doctor-workbench-row">
            <div className="doctor-slots-preview">
              <div className="panel-header compact-header">
                <div>
                  <p className="eyebrow">Ближайшее время</p>
                  <strong>Слоты врача</strong>
                </div>
                <CalendarDays size={18} />
              </div>
              {isLoading ? <span>Загружаю расписание...</span> : null}
              {!isLoading && !upcomingSlots.length ? <span>Будущих слотов пока нет.</span> : null}
              {upcomingSlots.map((slot) => (
                <div className="doctor-slot-preview" key={slot.id}>
                  <strong>{formatDoctorSlot(slot)}</strong>
                  <span className={`workspace-badge ${slotTone(slot.status)}`}>
                    {slotStatusLabel(slot.status)}
                  </span>
                </div>
              ))}
            </div>
            <Button type="button" variant="secondary" onClick={() => onOpenSchedule(doctor.id)}>
              Открыть расписание
            </Button>
          </div>
        </>
      )}
    </section>
  )
}

function DoctorLine({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value || '-'}</strong>
    </div>
  )
}

function formatDoctorSlot(slot: TimeSlot) {
  return new Date(slot.start_at).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function DoctorStatusBadge({ status }: { status: string }) {
  return (
    <span className={`workspace-badge ${doctorStatusTone(status)}`}>
      {doctorStatusLabel(status)}
    </span>
  )
}

function slotTone(status: string) {
  switch (status) {
    case 'FREE':
      return 'success'
    case 'BOOKED':
      return 'info'
    case 'BLOCKED':
      return 'danger'
    default:
      return 'neutral'
  }
}

function doctorStatusTone(status: string) {
  switch (status) {
    case 'ACTIVE':
      return 'success'
    case 'VACATION':
      return 'warning'
    case 'INACTIVE':
      return 'neutral'
    default:
      return 'info'
  }
}
