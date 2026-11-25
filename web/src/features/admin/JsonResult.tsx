import type { ActionResult } from './types'

type JsonResultProps = {
  result: ActionResult | null
}

export function JsonResult({ result }: JsonResultProps) {
  return (
    <pre className="api-result">
      {result ? JSON.stringify(result.payload, null, 2) : 'Результат запроса появится здесь'}
    </pre>
  )
}
