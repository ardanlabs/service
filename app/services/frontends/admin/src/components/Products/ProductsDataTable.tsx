'use client'

import * as React from 'react'
import DataTable from '@/components/DataTable/DataTable'
import { GenericProps, HeadCell } from '@/components/DataTable/types'
import Checkbox from '@mui/material/Checkbox'
import TableRow from '@mui/material/TableRow'
import TableCell from '@mui/material/TableCell'
import { DateCell } from '@/components/DataTable/Items/DateCell'

const headCells: readonly HeadCell[] = [
  { id: 'id', numeric: false, disablePadding: false, label: 'ID' },
  { id: 'name', numeric: false, disablePadding: false, label: 'Name' },
  { id: 'email', numeric: false, disablePadding: false, label: 'Email' },
  { id: 'roles', numeric: false, disablePadding: false, label: 'Roles' },
  {
    id: 'department',
    numeric: false,
    disablePadding: false,
    label: 'Department',
  },
  { id: 'enabled', numeric: false, disablePadding: false, label: 'Enabled' },
  {
    id: 'dateCreated',
    numeric: false,
    disablePadding: false,
    label: 'Date Created',
  },
  {
    id: 'dateUpdated',
    numeric: false,
    disablePadding: false,
    label: 'Date Updated',
  },
]

interface Data {
  id: string
  name: string
  email: string
  roles: string[]
  department: string
  enabled: boolean
  dateCreated: string
  dateUpdated: string
}

function createData(
  id: string,
  name: string,
  email: string,
  roles: string[],
  department: string,
  enabled: boolean,
  dateCreated: string,
  dateUpdated: string,
): Data {
  return {
    id,
    name,
    email,
    roles,
    department,
    enabled,
    dateCreated,
    dateUpdated,
  }
}
const data: Data[] = [
  createData(
    'sasd',
    'asd',
    'asd',
    ['asd'],
    'asd',
    true,
    'Thu Jan 13 2022 01:03:23',
    'Thu Jan 13 2022 01:03:23',
  ),
]

export default function ProductsDataTable() {
  const [selected, setSelected] = React.useState<readonly string[]>([])
  const [serverItemsLength, setServerItemsLength] = React.useState(0)
  const [rows, setRows] = React.useState<Data[]>([])

  function getData(props: GenericProps) {
    setServerItemsLength(data.length)
    setRows(data)
  }
  // handleRowSelectAllClick selects all rows when the checkbox is clicked
  const handleRowSelectAllClick = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    if (event.target.checked) {
      const newSelected = rows.map((n) => n.id)
      setSelected(newSelected)
      return
    }
    setSelected([])
  }

  // handleClick selects and unselects a row when clicked
  const handleClick = (event: React.MouseEvent<unknown>, id: string) => {
    const selectedIndex = selected.indexOf(id)
    let newSelected: readonly string[] = []

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selected, id)
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selected.slice(1))
    } else if (selectedIndex === selected.length - 1) {
      newSelected = newSelected.concat(selected.slice(0, -1))
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selected.slice(0, selectedIndex),
        selected.slice(selectedIndex + 1),
      )
    }

    setSelected(newSelected)
  }

  // isSelected checks if a row is selected
  const isSelected = (id: string) => selected.indexOf(id) !== -1

  return (
    <DataTable
      serverItemsLength={serverItemsLength}
      rowCount={rows.length}
      handleSelectAllClick={handleRowSelectAllClick}
      selectable={true}
      selectedCount={selected.length}
      headCells={headCells}
      getData={getData}
    >
      {rows.map((row, index) => {
        const isItemSelected = isSelected(row.id)
        const labelId = `enhanced-table-checkbox-${index}`

        return (
          <TableRow
            hover
            onClick={(event) => handleClick(event, row.id)}
            role="checkbox"
            aria-checked={isItemSelected}
            tabIndex={-1}
            key={row.id}
            selected={isItemSelected}
            sx={{ cursor: 'pointer' }}
          >
            <TableCell padding="checkbox">
              <Checkbox
                color="primary"
                checked={isItemSelected}
                inputProps={{
                  'aria-labelledby': labelId,
                }}
              />
            </TableCell>
            <TableCell id={labelId}>{row.id}</TableCell>
            <TableCell> {row.name} </TableCell>
            <TableCell> {row.email} </TableCell>
            <TableCell> {row.roles} </TableCell>
            <TableCell> {row.department} </TableCell>
            <TableCell> {`${row.enabled}`} </TableCell>
            <DateCell value={row.dateCreated} />
            <DateCell value={row.dateUpdated} />
          </TableRow>
        )
      })}
    </DataTable>
  )
}
