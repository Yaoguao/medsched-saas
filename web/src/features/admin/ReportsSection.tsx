import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { CalendarDays, CheckCircle2, CircleSlash, Clock3, Download, Stethoscope, TriangleAlert } from 'lucide-react'
import { coreApi } from '../../shared/api/core'
import { downloadCsv } from '../../shared/lib/csv'
import { appointmentStatusLabel } from '../../shared/lib/labels'
import { Button } from '../../shared/ui/Button'
import { ContentState } from '../../shared/ui/ContentState'
import { MetricCard } from '../../shared/ui/MetricCard'
import { AdminTable } from './AdminTable'

const statuses = ['CREATED', 'CONFIRMED', 'CANCELLED', 'COMPLETED', 'NO_SHOW']

export function ReportsSection() {
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const dashboardQuery = useQuery({
    queryKey: ['reports-dashboard', from, to],
    queryFn: () => coreApi.dashboardReport(from, to),
  })

  const report = dashboardQuery.data
  const appointments = report?.appointments
  const load = report?.load
  const noShow = report?.no_show
  const statusCounts = appointments?.status_counts ?? {}
  const total = appointments?.total ?? 0
  const noShowRate = Math.round((appointments?.no_show_rate ?? 0) * 100)
  const loadRate = Math.round((load?.load_rate ?? 0) * 100)
  const rangeLabel = report?.range
    ? `${report.range.from || 'начало'}_${report.range.to || 'сегодня'}`
    : `${from || 'начало'}_${to || 'сегодня'}`

  function exportStatusSummary() {
    if (!appointments) return
    downloadCsv(
      `medsched-statuses_${rangeLabel}.csv`,
      ['Период с', 'Период по', 'Статус', 'Количество', 'Доля от всех записей, %'],
      statuses.map((status) => {
        const count = statusCounts[status] ?? 0
        return [
          appointments.range.from,
          appointments.range.to,
          appointmentStatusLabel(status),
          count,
          total ? Math.round((count / total) * 10000) / 100 : 0,
        ]
      }),
    )
  }

  function exportNoShow() {
    if (!noShow) return
    downloadCsv(
      `medsched-no-show_${rangeLabel}.csv`,
      ['ID врача', 'Врач', 'Специальность', 'Записи', 'Неявки', 'Отмены', 'Завершены', 'Доля неявок, %'],
      noShow.doctors.map((row) => [
        row.doctor_id,
        row.doctor_name,
        row.specialty,
        row.total,
        row.no_show,
        row.cancelled,
        row.completed,
        Math.round(row.no_show_rate * 10000) / 100,
      ]),
    )
  }

  function exportLoad() {
    if (!load) return
    downloadCsv(
      `medsched-load_${rangeLabel}.csv`,
      ['ID врача', 'Врач', 'Специальность', 'Слоты', 'Заняты', 'Свободны', 'Заблокированы', 'Загрузка, %'],
      load.doctors.map((row) => [
        row.doctor_id,
        row.doctor_name,
        row.specialty,
        row.slots_total,
        row.booked_slots,
        row.free_slots,
        row.blocked_slots,
        Math.round(row.load_rate * 10000) / 100,
      ]),
    )
  }

  return (
    <section className="admin-section">
      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Отчеты</p>
            <h2>Период аналитики</h2>
          </div>
          <Clock3 size={20} />
        </div>
        <div className="report-filters">
          <input type="date" value={from} onChange={(event) => setFrom(event.target.value)} />
          <input type="date" value={to} onChange={(event) => setTo(event.target.value)} />
        </div>
        <div className="report-export-actions">
          <Button
            variant="secondary"
            icon={<Download size={16} />}
            disabled={!appointments || dashboardQuery.isFetching}
            onClick={exportStatusSummary}
          >
            Статусы CSV
          </Button>
          <Button
            variant="secondary"
            icon={<Download size={16} />}
            disabled={!noShow?.doctors.length || dashboardQuery.isFetching}
            onClick={exportNoShow}
          >
            Неявки CSV
          </Button>
          <Button
            variant="secondary"
            icon={<Download size={16} />}
            disabled={!load?.doctors.length || dashboardQuery.isFetching}
            onClick={exportLoad}
          >
            Загрузка CSV
          </Button>
        </div>
        {dashboardQuery.isLoading ? <ContentState variant="loading" title="Загружаю аналитику" compact /> : null}
        {dashboardQuery.isError ? (
          <ContentState variant="error" title="Не удалось загрузить отчёты" description={dashboardQuery.error.message} compact />
        ) : null}
      </article>

      <section className="metrics-grid">
        <MetricCard label="Всего записей" value={total} hint="по выбранному периоду" tone="blue" icon={<CalendarDays size={22} />} />
        <MetricCard label="Неявки" value={appointments?.no_show ?? 0} hint={`${noShowRate}% от всех записей`} tone="orange" icon={<TriangleAlert size={22} />} />
        <MetricCard label="Загрузка" value={`${loadRate}%`} hint={`${load?.booked_slots ?? 0}/${load?.slots_total ?? 0} слотов`} tone="green" icon={<Stethoscope size={22} />} />
        <MetricCard label="Отменены" value={appointments?.cancelled ?? 0} hint="отмененные записи" tone="violet" icon={<CircleSlash size={22} />} />
      </section>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Статусы</p>
            <h2>Записи по статусам</h2>
          </div>
          <CheckCircle2 size={20} />
        </div>
        <div className="report-bars">
          {statuses.map((status) => {
            const count = statusCounts[status] ?? 0
            const width = total ? `${Math.max(4, Math.round((count / total) * 100))}%` : '4%'
            return (
              <div className="report-bar" key={status}>
                <span>{appointmentStatusLabel(status)}</span>
                <div><b style={{ width }} /></div>
                <strong>{count}</strong>
              </div>
            )
          })}
        </div>
      </article>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Неявки</p>
            <h2>Неявки по врачам</h2>
          </div>
        </div>
        <AdminTable
          items={noShow?.doctors ?? []}
          empty="Нет данных по неявкам за период."
          columns={[
            { key: 'doctor', title: 'Врач', render: (row) => `${row.doctor_name} · ${row.specialty}` },
            { key: 'total', title: 'Записи', render: (row) => row.total },
            { key: 'no_show', title: 'Неявки', render: (row) => row.no_show },
            { key: 'cancelled', title: 'Отмены', render: (row) => row.cancelled },
            { key: 'completed', title: 'Завершены', render: (row) => row.completed },
            { key: 'rate', title: 'Доля', render: (row) => `${Math.round(row.no_show_rate * 100)}%` },
          ]}
        />
      </article>

      <article className="panel">
        <div className="panel-header">
          <div>
            <p className="eyebrow">Загрузка</p>
            <h2>Использование слотов врачей</h2>
          </div>
        </div>
        <AdminTable
          items={load?.doctors ?? []}
          empty="Нет слотов за выбранный период."
          columns={[
            { key: 'doctor', title: 'Врач', render: (row) => `${row.doctor_name} · ${row.specialty}` },
            { key: 'total', title: 'Слоты', render: (row) => row.slots_total },
            { key: 'booked', title: 'Заняты', render: (row) => row.booked_slots },
            { key: 'free', title: 'Свободны', render: (row) => row.free_slots },
            { key: 'blocked', title: 'Блоки', render: (row) => row.blocked_slots },
            { key: 'load', title: 'Загрузка', render: (row) => `${Math.round(row.load_rate * 100)}%` },
          ]}
        />
      </article>
    </section>
  )
}
