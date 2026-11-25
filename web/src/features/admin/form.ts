export function formString(form: FormData, key: string, fallback = '') {
  const value = form.get(key)
  return typeof value === 'string' && value.trim() ? value.trim() : fallback
}

export function formNumber(form: FormData, key: string, fallback = 0) {
  const value = Number(form.get(key))
  return Number.isFinite(value) ? value : fallback
}

export function formBoolean(form: FormData, key: string) {
  return form.get(key) === 'on'
}

export function formDateTime(form: FormData, key: string) {
  const value = formString(form, key)
  return value ? new Date(value).toISOString() : new Date().toISOString()
}

export function formDate(form: FormData, key: string) {
  const value = formString(form, key)
  return value ? `${value}T00:00:00Z` : new Date().toISOString()
}

export function formOptionalString(form: FormData, key: string) {
  const value = formString(form, key)
  return value || undefined
}

export function formOptionalNumber(form: FormData, key: string) {
  const value = form.get(key)
  if (value === null || value === '') {
    return undefined
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

export function formOptionalDate(form: FormData, key: string) {
  const value = formString(form, key)
  return value ? `${value}T00:00:00Z` : undefined
}

export function formOptionalDateTime(form: FormData, key: string) {
  const value = formString(form, key)
  return value ? new Date(value).toISOString() : undefined
}
