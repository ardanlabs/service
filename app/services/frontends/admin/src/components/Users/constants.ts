import { HeadCell } from '@/components/DataTable/types'

export const headCells: readonly HeadCell[] = [
  {
    id: 'id',
    numeric: false,
    disablePadding: false,
    label: 'ID',
    sortable: false,
  },
  {
    id: 'name',
    numeric: false,
    disablePadding: false,
    label: 'Name',
    sortable: true,
  },
  {
    id: 'email',
    numeric: false,
    disablePadding: false,
    label: 'Email',
    sortable: true,
  },
  {
    id: 'roles',
    numeric: false,
    disablePadding: false,
    label: 'Roles',
    sortable: true,
  },
  {
    id: 'department',
    numeric: false,
    disablePadding: false,
    label: 'Department',
    sortable: false,
  },
  {
    id: 'enabled',
    numeric: false,
    disablePadding: false,
    label: 'Enabled',
    sortable: true,
  },
  {
    id: 'dateCreated',
    numeric: false,
    disablePadding: false,
    label: 'Date Created',
    sortable: false,
  },
  {
    id: 'dateUpdated',
    numeric: false,
    disablePadding: false,
    label: 'Date Updated',
    sortable: false,
  },
  {
    id: 'actions',
    numeric: false,
    disablePadding: false,
    label: 'Actions',
    sortable: false,
  },
]

export interface User {
  id: string
  name: string
  email: string
  roles: string[]
  department: string
  enabled: boolean
  dateCreated: string
  dateUpdated: string
}
