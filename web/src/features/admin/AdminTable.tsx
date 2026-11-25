import type { ReactNode } from 'react'
import { ContentState } from '../../shared/ui/ContentState'

type AdminTableProps<T> = {
  items: T[]
  columns: Array<{
    key: string
    title: string
    render: (item: T) => ReactNode
  }>
  empty: string
}

export function AdminTable<T>({ items, columns, empty }: AdminTableProps<T>) {
  if (!items.length) {
    return <ContentState variant="empty" title={empty} compact />
  }

  return (
    <div className="admin-table-wrap">
      <table className="admin-table">
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column.key}>{column.title}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {items.map((item, index) => (
            <tr key={index}>
              {columns.map((column) => (
                <td key={column.key}>{column.render(item)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
