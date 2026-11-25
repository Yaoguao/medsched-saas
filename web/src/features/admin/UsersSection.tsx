import type { FormEvent } from 'react'
import { useMemo, useState } from 'react'
import { Shield, Stethoscope, UserCheck, UsersRound } from 'lucide-react'
import type { UpdateUserPayload, User, UserRole } from '../../shared/api/types'
import { userRoleLabel, userStatusLabel } from '../../shared/lib/labels'
import { Button } from '../../shared/ui/Button'
import { MetricCard } from '../../shared/ui/MetricCard'
import { AdminTable } from './AdminTable'
import { formOptionalString, formString } from './form'

type UsersSectionProps = {
  users: User[]
  isPending: boolean
  onUpdateUser: (userId: string, payload: UpdateUserPayload) => void
}

const roles: UserRole[] = ['ADMIN', 'REGISTRATOR', 'DOCTOR', 'PATIENT', 'MANAGER', 'IT_SUPPORT']

export function UsersSection({ users, isPending, onUpdateUser }: UsersSectionProps) {
  const [query, setQuery] = useState('')

  const filteredUsers = useMemo(() => {
    const search = query.trim().toLowerCase()
    if (!search) {
      return users
    }
    return users.filter((user) =>
      [user.email, user.phone, user.role, user.status, user.id, user.patient_id, user.doctor_id]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search)),
    )
  }, [query, users])

  function updateUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const role = formOptionalString(form, 'role') as UserRole | undefined
    onUpdateUser(formString(form, 'user_id'), {
      role,
      status: formOptionalString(form, 'status'),
      email: formOptionalString(form, 'email'),
      phone: formOptionalString(form, 'phone'),
    })
  }

  const activeCount = users.filter((user) => user.status === 'ACTIVE').length
  const staffCount = users.filter((user) => user.role !== 'PATIENT').length
  const linkedCount = users.filter((user) => user.patient_id || user.doctor_id).length

  return (
    <>
    <section className="metrics-grid">
      <MetricCard label="Аккаунты" value={users.length} hint="в текущей выборке" tone="blue" icon={<UsersRound size={22} />} />
      <MetricCard label="Активны" value={activeCount} hint="доступ разрешен" tone="green" icon={<UserCheck size={22} />} />
      <MetricCard label="Сотрудники" value={staffCount} hint="не пациентские роли" tone="violet" icon={<Shield size={22} />} />
      <MetricCard label="Привязки" value={linkedCount} hint="к пациенту или врачу" tone="orange" icon={<Stethoscope size={22} />} />
    </section>

    <section className="admin-two-column">
      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Доступ</p>
            <h2>Изменить пользователя</h2>
          </div>
        </div>
        <form className="admin-form" onSubmit={updateUser}>
          <select name="user_id" required>
            <option value="">Выбери аккаунт</option>
            {users.map((user) => (
              <option key={user.id} value={user.id}>
                {user.email} · {userRoleLabel(user.role)} · {userStatusLabel(user.status)}
              </option>
            ))}
          </select>
          <select name="role" defaultValue="">
            <option value="">Роль без изменений</option>
            {roles.map((role) => (
              <option key={role} value={role}>{userRoleLabel(role)}</option>
            ))}
          </select>
          <select name="status" defaultValue="">
            <option value="">Статус без изменений</option>
            <option value="ACTIVE">{userStatusLabel('ACTIVE')}</option>
            <option value="BLOCKED">{userStatusLabel('BLOCKED')}</option>
            <option value="INACTIVE">{userStatusLabel('INACTIVE')}</option>
          </select>
          <input name="email" type="email" placeholder="Новый email, если нужно" />
          <input name="phone" placeholder="Новый телефон, если нужно" />
          <Button type="submit" disabled={isPending}>Сохранить аккаунт</Button>
        </form>
        <p className="section-note">
          Роль врача также выдается через привязку аккаунта к врачу в разделе врачей.
        </p>
      </article>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Список</p>
            <h2>Пользователи</h2>
          </div>
          <input
            className="table-filter"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Поиск по email, роли, статусу"
          />
        </div>
        <AdminTable
          items={filteredUsers}
          empty="Пользователей пока нет."
          columns={[
            { key: 'email', title: 'Email', render: (user) => user.email },
            { key: 'phone', title: 'Телефон', render: (user) => user.phone || '-' },
            { key: 'role', title: 'Роль', render: (user) => <span className="workspace-badge info">{userRoleLabel(user.role)}</span> },
            { key: 'status', title: 'Статус', render: (user) => <UserStatusBadge status={user.status} /> },
            { key: 'patient', title: 'Пациент', render: (user) => user.patient_id ? <span className="workspace-badge success">Привязан</span> : '-' },
            { key: 'doctor', title: 'Врач', render: (user) => user.doctor_id ? <span className="workspace-badge success">Привязан</span> : '-' },
          ]}
        />
      </article>
    </section>
    </>
  )
}

function UserStatusBadge({ status }: { status: string }) {
  return (
    <span className={`workspace-badge ${userStatusTone(status)}`}>
      {userStatusLabel(status)}
    </span>
  )
}

function userStatusTone(status: string) {
  switch (status) {
    case 'ACTIVE':
      return 'success'
    case 'BLOCKED':
      return 'danger'
    case 'INACTIVE':
      return 'neutral'
    default:
      return 'info'
  }
}
