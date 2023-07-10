'use client'

import * as React from 'react'
import DataTable from '@/components/DataTable/DataTable'
import { GenericProps } from '@/components/DataTable/types'
import Checkbox from '@mui/material/Checkbox'
import TableRow from '@mui/material/TableRow'
import TableCell from '@mui/material/TableCell'
import { DateCell } from '@/components/DataTable/Items/DateCell'
import { DefaultAPIResponse } from '@/utils/types'
import { User, headCells } from './constants'

export default function UsersDataTable() {
  const [selected, setSelected] = React.useState<readonly string[]>([])
  const [serverItemsLength, setServerItemsLength] = React.useState(0)
  const [rows, setRows] = React.useState<User[]>([])

  async function getData(props: GenericProps) {
    const fetchCall = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_API_URL}/users?page=1&rows=2`,
      {
        headers: {
          Authorization: `Bearer ${process.env.NEXT_PUBLIC_TOKEN}`,
        },
      },
    )
    if (fetchCall.ok) {
      const fetchData: DefaultAPIResponse<User> = await fetchCall.json()
      setServerItemsLength(fetchData.total)
      setRows(fetchData.items)
      return
    }
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
