import { Order } from '@/components/DataTable/types'

export default function parseOrderBy(order: string, direction: Order): string {
  return order && direction
    ? `&orderBy=${order},${direction?.toUpperCase()}&`
    : ''
}
