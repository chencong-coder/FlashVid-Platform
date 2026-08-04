const BACKEND_DATE_TIME_PATTERN = /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/

/** Parse the backend's local datetime format consistently across H5 browsers. */
export const parseApiDate = (value: string): Date | null => {
  const normalized = BACKEND_DATE_TIME_PATTERN.test(value) ? value.replace(' ', 'T') : value
  const date = new Date(normalized)

  return Number.isNaN(date.getTime()) ? null : date
}
