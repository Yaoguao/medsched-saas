import type { FormEvent } from 'react'
import type {
  Appointment,
  CreateMLTrainingJobPayload,
  MLModelInfo,
  MLModelVersion,
  MLTrainingJob,
} from '../../shared/api/types'
import { appointmentStatusLabel } from '../../shared/lib/labels'
import { Button } from '../../shared/ui/Button'
import { formNumber, formString } from './form'

type MlSectionProps = {
  appointments: Appointment[]
  isPending: boolean
  modelInfo?: MLModelInfo
  isModelInfoLoading: boolean
  trainingJobs: MLTrainingJob[]
  models: MLModelVersion[]
  isRegistryLoading: boolean
  canManage: boolean
  onPredict: (appointmentId: string) => void
  onStartTraining: (payload: CreateMLTrainingJobPayload) => void
  onActivateModel: (version: string) => void
}

export function MlSection({
  appointments,
  isPending,
  modelInfo,
  isModelInfoLoading,
  trainingJobs,
  models,
  isRegistryLoading,
  canManage,
  onPredict,
  onStartTraining,
  onActivateModel,
}: MlSectionProps) {
  function predict(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onPredict(formString(form, 'appointment_id'))
  }

  function startTraining(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    onStartTraining({
      source: formString(form, 'source', 'csv') as CreateMLTrainingJobPayload['source'],
      algorithm: formString(form, 'algorithm', 'xgboost') as CreateMLTrainingJobPayload['algorithm'],
      core_limit: formNumber(form, 'core_limit', 5000),
    })
  }

  const hasActiveJob = trainingJobs.some((job) => job.status === 'QUEUED' || job.status === 'RUNNING')

  return (
    <section className="admin-two-column">
      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Активная модель</p>
            <h2>Состояние прогнозирования</h2>
          </div>
          <span className={modelInfo?.loaded ? 'status-pill success' : 'status-pill warning'}>
            {isModelInfoLoading ? 'Загрузка' : modelInfo?.loaded ? 'Модель загружена' : 'Резервная эвристика'}
          </span>
        </div>
        <div className="metrics-grid ml-model-grid">
          <Metric label="Версия" value={modelInfo?.model_version ?? '-'} />
          <Metric label="Алгоритм" value={modelInfo?.algorithm ?? '-'} />
          <Metric label="ROC AUC" value={formatMetric(modelInfo?.metrics.roc_auc)} />
          <Metric label="Recall" value={formatMetric(modelInfo?.metrics.recall)} />
        </div>
        <dl className="ml-model-details">
          <div>
            <dt>Дата обучения</dt>
            <dd>{formatDateMetric(modelInfo?.metrics.trained_at)}</dd>
          </div>
          <div>
            <dt>Строк датасета</dt>
            <dd>{formatMetric(modelInfo?.metrics.rows)}</dd>
          </div>
          <div>
            <dt>Average precision</dt>
            <dd>{formatMetric(modelInfo?.metrics.average_precision)}</dd>
          </div>
          <div>
            <dt>Пороги риска</dt>
            <dd>{formatThresholds(modelInfo)}</dd>
          </div>
        </dl>
      </article>

      {canManage ? (
        <article className="panel">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Переобучение</p>
              <h2>Новое задание</h2>
            </div>
          </div>
          <form className="admin-form" onSubmit={startTraining}>
            <label>
              Источник данных
              <select name="source" defaultValue="csv">
                <option value="csv">Базовый healthcare_noshows.csv</option>
                <option value="core">Накопленные данные Core</option>
              </select>
            </label>
            <label>
              Алгоритм
              <select name="algorithm" defaultValue="xgboost">
                <option value="xgboost">XGBoost</option>
                <option value="logistic_regression">Logistic Regression</option>
              </select>
            </label>
            <label>
              Максимум завершённых записей
              <input name="core_limit" type="number" min="30" max="100000" defaultValue="5000" />
            </label>
            <Button type="submit" disabled={isPending || hasActiveJob}>
              {hasActiveJob ? 'Обучение уже выполняется' : 'Запустить обучение'}
            </Button>
          </form>
        </article>
      ) : null}

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Прогноз</p>
            <h2>Риск по записи</h2>
          </div>
        </div>
        <form className="admin-form" onSubmit={predict}>
          <select name="appointment_id" required>
            <option value="">Выберите запись</option>
            {appointments.map((appointment) => (
              <option key={appointment.id} value={appointment.id}>
                {new Date(appointment.appointment_at).toLocaleString('ru-RU')} · {appointmentStatusLabel(appointment.status)}
              </option>
            ))}
          </select>
          <Button type="submit" disabled={isPending}>Рассчитать риск</Button>
        </form>
      </article>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Очередь</p>
            <h2>Задания обучения</h2>
          </div>
          <span>{isRegistryLoading ? 'Обновление...' : `${trainingJobs.length} заданий`}</span>
        </div>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>Версия</th>
                <th>Алгоритм</th>
                <th>Источник</th>
                <th>Статус</th>
                <th>Данные</th>
                <th>Создано</th>
              </tr>
            </thead>
            <tbody>
              {trainingJobs.map((job) => (
                <tr key={job.id}>
                  <td>{job.model_version}</td>
                  <td>{algorithmLabel(job.algorithm)}</td>
                  <td>{sourceLabel(job.source)}</td>
                  <td><span className={statusClass(job.status)}>{jobStatusLabel(job.status)}</span></td>
                  <td>{job.dataset_rows || '-'}</td>
                  <td>{formatDate(job.created_at)}</td>
                </tr>
              ))}
              {!trainingJobs.length ? (
                <tr><td colSpan={6}>Заданий обучения пока нет.</td></tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </article>

      <article className="panel admin-wide">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Реестр</p>
            <h2>Версии моделей</h2>
          </div>
        </div>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>Версия</th>
                <th>Статус</th>
                <th>Источник</th>
                <th>ROC AUC</th>
                <th>Recall</th>
                <th>Строк</th>
                <th aria-label="Действия" />
              </tr>
            </thead>
            <tbody>
              {models.map((model) => (
                <tr key={model.id}>
                  <td>{model.version}</td>
                  <td><span className={statusClass(model.status)}>{modelStatusLabel(model.status)}</span></td>
                  <td>{sourceLabel(model.dataset_source)}</td>
                  <td>{formatMetric(numberMetric(model.metrics.roc_auc))}</td>
                  <td>{formatMetric(numberMetric(model.metrics.recall))}</td>
                  <td>{model.dataset_rows}</td>
                  <td>
                    {canManage && (model.status === 'CANDIDATE' || model.status === 'ARCHIVED') ? (
                      <Button
                        type="button"
                        variant="secondary"
                        disabled={isPending}
                        onClick={() => onActivateModel(model.version)}
                      >
                        Активировать
                      </Button>
                    ) : null}
                  </td>
                </tr>
              ))}
              {!models.length ? (
                <tr><td colSpan={7}>Обученные версии пока не зарегистрированы.</td></tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </article>
    </section>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric-card compact-metric">
      <div>
        <small>{label}</small>
        <strong>{value}</strong>
      </div>
    </div>
  )
}

function formatMetric(value: number | string | undefined) {
  if (value === undefined || value === '') return '-'
  if (typeof value === 'number') return Number.isInteger(value) ? String(value) : value.toFixed(4)
  return value
}

function numberMetric(value: unknown) {
  return typeof value === 'number' ? value : undefined
}

function formatDateMetric(value: unknown) {
  return typeof value === 'string' ? formatDate(value) : '-'
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('ru-RU')
}

function formatThresholds(modelInfo?: MLModelInfo) {
  const thresholds = modelInfo?.metrics.risk_thresholds
  if (!thresholds) return '-'
  return `Средний ${thresholds.medium.toFixed(2)}, высокий ${thresholds.high.toFixed(2)}`
}

function algorithmLabel(value: string) {
  return value === 'logistic_regression' ? 'Logistic Regression' : 'XGBoost'
}

function sourceLabel(value: string) {
  return value === 'csv' ? 'Базовый CSV' : 'Данные Core'
}

function jobStatusLabel(value: MLTrainingJob['status']) {
  return {
    QUEUED: 'В очереди',
    RUNNING: 'Обучается',
    SUCCEEDED: 'Завершено',
    FAILED: 'Ошибка',
    CANCELLED: 'Отменено',
  }[value]
}

function modelStatusLabel(value: MLModelVersion['status']) {
  return {
    CANDIDATE: 'Кандидат',
    ACTIVE: 'Активна',
    REJECTED: 'Отклонена',
    ARCHIVED: 'Архив',
  }[value]
}

function statusClass(value: string) {
  if (value === 'ACTIVE' || value === 'SUCCEEDED') return 'status-pill success'
  if (value === 'FAILED' || value === 'REJECTED') return 'status-pill warning'
  return 'status-pill'
}
