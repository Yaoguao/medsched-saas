import { Brain, CalendarDays, HeartPulse, Server, Stethoscope, UsersRound } from 'lucide-react'
import type { ActionResult } from './types'
import { Button } from '../../shared/ui/Button'
import { JsonResult } from './JsonResult'

type ApiConsoleSectionProps = {
  result: ActionResult | null
  isPending: boolean
  onHealth: () => void
  onUsers: () => void
  onPatients: () => void
  onDoctors: () => void
  onAppointments: () => void
  onTrainingDataset: () => void
  canCheckUsers: boolean
  canCheckPatients: boolean
  canCheckAppointments: boolean
  canCheckTrainingDataset: boolean
}

export function ApiConsoleSection({
  result,
  isPending,
  onHealth,
  onUsers,
  onPatients,
  onDoctors,
  onAppointments,
  onTrainingDataset,
  canCheckUsers,
  canCheckPatients,
  canCheckAppointments,
  canCheckTrainingDataset,
}: ApiConsoleSectionProps) {
  const actions = [
    {
      title: 'Состояние Core API',
      text: 'Проверить доступность основного backend-сервиса.',
      icon: <Server size={20} />,
      enabled: true,
      action: onHealth,
    },
    {
      title: 'Пользователи',
      text: 'Получить список аккаунтов и проверить доступ к RBAC.',
      icon: <UsersRound size={20} />,
      enabled: canCheckUsers,
      action: onUsers,
    },
    {
      title: 'Пациенты',
      text: 'Проверить справочник пациентов и медицинские карты.',
      icon: <HeartPulse size={20} />,
      enabled: canCheckPatients,
      action: onPatients,
    },
    {
      title: 'Врачи',
      text: 'Получить справочник врачей и связанных профилей.',
      icon: <Stethoscope size={20} />,
      enabled: true,
      action: onDoctors,
    },
    {
      title: 'Записи',
      text: 'Проверить записи на прием, статусы и связанные сущности.',
      icon: <CalendarDays size={20} />,
      enabled: canCheckAppointments,
      action: onAppointments,
    },
    {
      title: 'Обучающая выборка',
      text: 'Получить данные для переобучения ML-модели no-show.',
      icon: <Brain size={20} />,
      enabled: canCheckTrainingDataset,
      action: onTrainingDataset,
    },
  ]

  return (
    <section className="admin-section">
      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Инструменты</p>
            <h2>Быстрая проверка API</h2>
          </div>
          <span className="workspace-badge info">Core Service</span>
        </div>

        <div className="api-action-grid">
          {actions.map((item) => (
            <article className={item.enabled ? 'api-action-card' : 'api-action-card disabled'} key={item.title}>
              <div>
                <span>{item.icon}</span>
                <strong>{item.title}</strong>
                <small>{item.text}</small>
              </div>
              <Button variant="secondary" disabled={isPending || !item.enabled} onClick={item.action}>
                Выполнить
              </Button>
            </article>
          ))}
        </div>
      </article>

      <article className="panel api-panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Ответ</p>
            <h2>Диагностика запроса</h2>
          </div>
          {result ? <span className={result.title === 'Ошибка запроса' ? 'workspace-badge danger' : 'workspace-badge success'}>{result.title}</span> : null}
        </div>
        <JsonResult result={result} />
      </article>
    </section>
  )
}
