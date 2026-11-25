import { useQuery } from '@tanstack/react-query'
import { Bell, CheckCircle2, CircleAlert, Clock3, Mail, MessageSquareText } from 'lucide-react'
import type { FormEvent, ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { coreApi } from '../../shared/api/core'
import type { Appointment, CreateNotificationPayload } from '../../shared/api/types'
import { appointmentStatusLabel } from '../../shared/lib/labels'
import { Button } from '../../shared/ui/Button'
import { ContentState } from '../../shared/ui/ContentState'
import { formDateTime, formString } from './form'

type NotificationsSectionProps = {
  appointments: Appointment[]
  appointmentId: string
  isPending: boolean
  onSelectAppointment: (appointmentId: string) => void
  onCreateNotification: (payload: CreateNotificationPayload) => void
}

export function NotificationsSection({
  appointments,
  appointmentId,
  isPending,
  onSelectAppointment,
  onCreateNotification,
}: NotificationsSectionProps) {
  const notificationsQuery = useQuery({
    queryKey: ['appointment-notifications', appointmentId],
    queryFn: () => coreApi.appointmentNotifications(appointmentId, { limit: 50 }),
    enabled: Boolean(appointmentId),
  })

  function createNotification(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onCreateNotification({
      appointment_id: formString(form, 'appointment_id', appointmentId),
      patient_id: formString(form, 'patient_id'),
      channel: formString(form, 'channel', 'EMAIL') as CreateNotificationPayload['channel'],
      type: formString(form, 'type', 'REMINDER') as CreateNotificationPayload['type'],
      status: formString(form, 'status', 'PENDING') as CreateNotificationPayload['status'],
      scheduled_at: formDateTime(form, 'scheduled_at'),
      provider: formString(form, 'provider'),
      external_id: formString(form, 'external_id'),
      error_message: formString(form, 'error_message'),
    })
  }

  const selectedAppointment = appointments.find((item) => item.id === appointmentId)
  const notifications = notificationsQuery.data?.items ?? []
  const deliveredCount = notifications.filter((item) => item.status === 'DELIVERED' || item.status === 'SENT').length
  const pendingCount = notifications.filter((item) => item.status === 'PENDING').length
  const failedCount = notifications.filter((item) => item.status === 'FAILED').length

  return (
    <section className="notifications-workspace">
      <article className="notifications-overview">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Центр уведомлений</p>
            <h2>История доставки по записи</h2>
          </div>
          <Bell size={20} />
        </div>
        <select value={appointmentId} onChange={(event) => onSelectAppointment(event.target.value)}>
          <option value="">Выбери запись</option>
          {appointments.map((appointment) => (
            <option key={appointment.id} value={appointment.id}>
              {new Date(appointment.appointment_at).toLocaleString('ru-RU')} · {appointment.patient?.full_name ?? 'Пациент'} · {appointmentStatusLabel(appointment.status)}
            </option>
          ))}
        </select>

        {selectedAppointment ? (
          <div className="notification-appointment-summary">
            <div>
              <span>Пациент</span>
              <strong>{selectedAppointment.patient?.full_name ?? selectedAppointment.patient_id}</strong>
            </div>
            <div>
              <span>Прием</span>
              <strong>{new Date(selectedAppointment.appointment_at).toLocaleString('ru-RU')}</strong>
            </div>
            <div>
              <span>Статус записи</span>
              <strong>{appointmentStatusLabel(selectedAppointment.status)}</strong>
            </div>
            <Link className="button button-secondary" to={`/workspace/appointments/${selectedAppointment.id}`}>
              Открыть запись
            </Link>
          </div>
        ) : null}
      </article>

      <section className="notification-metrics">
        <NotificationMetric icon={<MessageSquareText size={19} />} label="Всего" value={notifications.length} />
        <NotificationMetric icon={<CheckCircle2 size={19} />} label="Отправлено" value={deliveredCount} tone="success" />
        <NotificationMetric icon={<Clock3 size={19} />} label="Ожидает" value={pendingCount} tone="warning" />
        <NotificationMetric icon={<CircleAlert size={19} />} label="Ошибки" value={failedCount} tone="danger" />
      </section>

      <article className="panel notification-timeline-panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Лента доставки</p>
            <h2>События уведомлений</h2>
          </div>
        </div>
        {notificationsQuery.isFetching ? <ContentState variant="loading" title="Загружаю уведомления" compact /> : null}
        {notificationsQuery.isError ? (
          <ContentState variant="error" title="Не удалось загрузить уведомления" description={notificationsQuery.error.message} compact />
        ) : null}
        {!appointmentId ? (
          <ContentState variant="empty" title="Запись не выбрана" description="Выбери запись, чтобы увидеть историю уведомлений." compact />
        ) : null}
        {!notificationsQuery.isFetching && appointmentId && !notifications.length ? (
          <ContentState variant="empty" title="Уведомлений пока нет" description="Они создадутся автоматически при событиях записи." compact />
        ) : null}
        <div className="notification-delivery-list">
          {notifications.map((item) => (
            <article key={item.id}>
              <div className={`notification-channel-icon ${item.channel.toLowerCase()}`}>
                {item.channel === 'EMAIL' ? <Mail size={17} /> : <MessageSquareText size={17} />}
              </div>
              <div>
                <strong>{notificationTypeLabel(item.type)}</strong>
                <span>{notificationChannelLabel(item.channel)} · {item.provider || 'внутренний сервис'}</span>
                <small>
                  Запланировано: {new Date(item.scheduled_at).toLocaleString('ru-RU')}
                  {item.sent_at ? ` · отправлено: ${new Date(item.sent_at).toLocaleString('ru-RU')}` : ''}
                </small>
                {item.error_message ? <p className="notification-error">{item.error_message}</p> : null}
                <Link className="notification-appointment-link" to={`/workspace/appointments/${item.appointment_id}`}>
                  Открыть запись
                </Link>
              </div>
              <NotificationStatusBadge status={item.status} />
            </article>
          ))}
        </div>
      </article>

      <details className="panel notification-manual">
        <summary>Дополнительная ручная отправка</summary>
        <p className="section-note">
          Основные подтверждения и напоминания создаются автоматически. Эта форма нужна для диагностики и повторной отправки.
        </p>
        <form className="admin-form" onSubmit={createNotification}>
          <input name="appointment_id" type="hidden" value={appointmentId} />
          <input name="patient_id" type="hidden" value={selectedAppointment?.patient_id ?? ''} />
          <select name="channel" defaultValue="EMAIL">
            <option value="EMAIL">{notificationChannelLabel('EMAIL')}</option>
            <option value="SMS">{notificationChannelLabel('SMS')}</option>
            <option value="PUSH">{notificationChannelLabel('PUSH')}</option>
          </select>
          <select name="type" defaultValue="REMINDER">
            <option value="CONFIRMATION">{notificationTypeLabel('CONFIRMATION')}</option>
            <option value="REMINDER">{notificationTypeLabel('REMINDER')}</option>
            <option value="CANCELLATION">{notificationTypeLabel('CANCELLATION')}</option>
          </select>
          <input name="scheduled_at" type="datetime-local" required />
          <input name="provider" placeholder="Провайдер, если требуется" />
          <input name="external_id" placeholder="Внешний ID, если известен" />
          <input name="error_message" type="hidden" value="" />
          <input name="status" type="hidden" value="PENDING" />
          <Button type="submit" disabled={isPending || !selectedAppointment}>Отправить вручную</Button>
        </form>
      </details>
    </section>
  )
}

function NotificationMetric({
  icon,
  label,
  value,
  tone = 'neutral',
}: {
  icon: ReactNode
  label: string
  value: number
  tone?: 'neutral' | 'success' | 'warning' | 'danger'
}) {
  return (
    <article className={tone}>
      {icon}
      <div><span>{label}</span><strong>{value}</strong></div>
    </article>
  )
}

function NotificationStatusBadge({ status }: { status: string }) {
  return (
    <span className={`workspace-badge ${notificationStatusTone(status)}`}>
      {notificationStatusLabel(status)}
    </span>
  )
}

function notificationStatusLabel(status: string) {
  const labels: Record<string, string> = {
    PENDING: 'Запланировано',
    SENT: 'Отправлено',
    FAILED: 'Ошибка',
    DELIVERED: 'Доставлено',
  }
  return labels[status] ?? status
}

function notificationStatusTone(status: string) {
  switch (status) {
    case 'SENT':
    case 'DELIVERED':
      return 'success'
    case 'FAILED':
      return 'danger'
    case 'PENDING':
      return 'warning'
    default:
      return 'neutral'
  }
}

function notificationTypeLabel(type: string) {
  const labels: Record<string, string> = {
    CONFIRMATION: 'Подтверждение',
    REMINDER: 'Напоминание',
    CANCELLATION: 'Отмена',
  }
  return labels[type] ?? type
}

function notificationChannelLabel(channel: string) {
  const labels: Record<string, string> = {
    EMAIL: 'Email',
    SMS: 'SMS',
    PUSH: 'Push',
  }
  return labels[channel] ?? channel
}
