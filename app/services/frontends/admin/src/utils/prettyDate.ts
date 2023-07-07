import intlFormat from 'date-fns/intlFormat'
import isValid from 'date-fns/isValid'

export default function prettyDate(value: string): string {
  const date = new Date(value)

  return isValid(date) ? intlFormat(date, {}, { locale: 'en-GB' }) : '-'
  return new Date(date).toString().slice(0, 15)
}
