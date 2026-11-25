export type AdminSection =
  | 'overview'
  | 'users'
  | 'patients'
  | 'doctors'
  | 'schedules'
  | 'appointments'
  | 'notifications'
  | 'integrations'
  | 'reports'
  | 'ml'
  | 'api'

export type ActionResult = {
  title: string
  payload: unknown
}
