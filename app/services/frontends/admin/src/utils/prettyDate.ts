import intlFormat from 'date-fns/intlFormat'
import isValid from 'date-fns/isValid'

// prettyDate returns a curated with the following format: 3/23/2019
export default function prettyDate(value: string): string {
  const date = new Date(value)

  return isValid(date) ? intlFormat(date, {}, { locale: 'en-US' }) : '-'
}
