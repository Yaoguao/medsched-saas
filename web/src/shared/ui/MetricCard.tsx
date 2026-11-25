import type { ReactNode } from 'react'

type MetricCardProps = {
  label: string
  value: string | number
  hint: string
  icon: ReactNode
  tone?: 'blue' | 'green' | 'orange' | 'violet'
}

export function MetricCard({ label, value, hint, icon, tone = 'blue' }: MetricCardProps) {
  return (
    <article className="metric-card">
      <div>
        <span className="metric-label">{label}</span>
        <strong>{value}</strong>
        <small>{hint}</small>
      </div>
      <div className={`metric-icon metric-${tone}`}>{icon}</div>
    </article>
  )
}
