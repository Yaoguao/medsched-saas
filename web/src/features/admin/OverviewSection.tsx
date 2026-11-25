import { BarChart3, Brain, CalendarDays, Stethoscope, TriangleAlert, UsersRound } from 'lucide-react'
import type { Appointment, Doctor, Patient, UserRole } from '../../shared/api/types'
import { appointmentStatusLabel, riskLevelLabel, visitTypeLabel } from '../../shared/lib/labels'
import { MetricCard } from '../../shared/ui/MetricCard'

type OverviewSectionProps = {
  role?: UserRole
  appointments: Appointment[]
  patients: Patient[]
  doctors: Doctor[]
}

export function OverviewSection({
  role,
  appointments,
  patients,
  doctors,
}: OverviewSectionProps) {
  const predictedCount = appointments.filter((appointment) => appointment.prediction_result).length
  const todayCount = appointments.filter((appointment) => isToday(appointment.appointment_at)).length
  const highRisk = appointments.filter((appointment) => appointment.prediction_result?.risk_level === 'HIGH')
  const statusCounts = appointments.reduce<Record<string, number>>((acc, appointment) => {
    acc[appointment.status] = (acc[appointment.status] ?? 0) + 1
    return acc
  }, {})
  const isManager = role === 'MANAGER'
  const isRegistrator = role === 'REGISTRATOR'

  return (
    <>
      <section className="metrics-grid">
        {!isManager ? (
          <MetricCard
            label="Записи"
            value={todayCount || appointments.length}
            hint={todayCount ? 'записи на сегодня' : 'доступные текущей роли'}
            tone="blue"
            icon={<CalendarDays size={22} />}
          />
        ) : null}
        {!isManager ? (
          <MetricCard
            label="Пациенты"
            value={patients.length}
            hint={isRegistrator ? 'карты и привязка аккаунтов' : 'загружены из core'}
            tone="green"
            icon={<UsersRound size={22} />}
          />
        ) : null}
        <MetricCard
          label="Врачи"
          value={doctors.length}
          hint={isManager ? 'для составления расписаний' : 'активный справочник'}
          tone="violet"
          icon={<Stethoscope size={22} />}
        />
        {isManager ? (
          <MetricCard
            label="Отчеты"
            value="доступны"
            hint="загрузка, no-show, статусы"
            tone="orange"
            icon={<TriangleAlert size={22} />}
          />
        ) : (
          <MetricCard
            label="ML прогнозы"
            value={predictedCount}
            hint="по загруженным записям"
            tone="orange"
            icon={<Brain size={22} />}
          />
        )}
      </section>

      <section className={isManager ? 'workspace-grid single' : 'workspace-grid'}>
        {!isManager ? (
          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Неявки</p>
                <h2>ML-прогнозы</h2>
              </div>
              <TriangleAlert size={20} />
            </div>
            <div className="risk-list">
              {highRisk.slice(0, 5).map((appointment) => (
                <div className="risk-item" key={appointment.id}>
                  <div>
                    <strong>{appointment.patient?.full_name ?? appointment.patient_id}</strong>
                    <span>{new Date(appointment.appointment_at).toLocaleString('ru-RU')} · {appointment.doctor?.full_name ?? appointment.doctor_id}</span>
                  </div>
                  <b>{riskLevelLabel(appointment.prediction_result?.risk_level ?? 'HIGH')} · {Math.round((appointment.prediction_result?.probability ?? 0) * 100)}%</b>
                </div>
              ))}
              {!highRisk.length ? (
                <p className="empty-state">Высоких рисков по загруженным записям нет.</p>
              ) : null}
            </div>
          </article>
        ) : null}

        <article className="panel">
          <div className="panel-header">
            <div>
              <p className="eyebrow">{isManager ? 'Расписание' : 'Записи'}</p>
              <h2>{isManager ? 'Планирование работы врачей' : 'Ближайшие записи'}</h2>
            </div>
            <CalendarDays size={20} />
          </div>
          {isManager ? (
            <p className="empty-state">
              Менеджер составляет расписания и управляет слотами. Запись пациентов выполняет регистратор.
            </p>
          ) : (
            <div className="appointment-list">
              {appointments.slice(0, 5).map((item) => (
                <div className="appointment-row" key={item.id}>
                  <div>
                    <span>{new Date(item.appointment_at).toLocaleString('ru-RU')}</span>
                    <small>{item.patient?.full_name ?? item.patient_id} · {visitTypeLabel(item.visit_type)}</small>
                  </div>
                  <strong>{appointmentStatusLabel(item.status)}</strong>
                </div>
              ))}
              {!appointments.length ? (
                <p className="empty-state">Записей пока нет.</p>
              ) : null}
            </div>
          )}
        </article>
      </section>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Статусы</p>
            <h2>Распределение записей</h2>
          </div>
          <BarChart3 size={20} />
        </div>
        <div className="status-summary-grid">
          {['CREATED', 'CONFIRMED', 'COMPLETED', 'NO_SHOW', 'CANCELLED'].map((status) => (
            <div key={status}>
              <span>{appointmentStatusLabel(status)}</span>
              <strong>{statusCounts[status] ?? 0}</strong>
            </div>
          ))}
        </div>
      </article>
    </>
  )
}

function isToday(value: string) {
  const date = new Date(value)
  const today = new Date()
  return date.getFullYear() === today.getFullYear()
    && date.getMonth() === today.getMonth()
    && date.getDate() === today.getDate()
}
