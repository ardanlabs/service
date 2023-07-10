export interface HeadCell {
  disablePadding: boolean
  id: string
  label: string
  numeric: boolean
  sortable: boolean
}

type Order = 'asc' | 'desc' | undefined

export interface GenericProps {
  page: number
  rows: number
  order: string
  direction: Order
}

export interface DataTableProps {
  headCells: readonly HeadCell[]
  getData: (GenericProps) => any
  dense?: boolean
  defaultOrder?: string
  defaultOrderDirection?: Order
  pageProp?: number
  rowsPerPageProp?: number
  rowsPerPageOptions?: number[]
  selectable?: boolean
  rowCount: number
  serverItemsLength: number
  selectedCount: number
  handleSelectAllClick: (event: React.ChangeEvent<HTMLInputElement>) => void
}
