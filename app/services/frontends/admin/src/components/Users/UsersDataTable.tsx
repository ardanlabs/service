'use client'

import * as React from 'react'
import DataTable from '@/components/DataTable/DataTable'
import TableRow from '@mui/material/TableRow'
import TableCell from '@mui/material/TableCell'
import { DateCell } from '@/components/DataTable/Items/DateCell'
import { DefaultAPIResponse } from '@/utils/types'
import { User, headCells } from './constants'
import { GenericProps } from '../DataTable/types'
import ActionsCell from '@/components/DataTable/Items/ActionsCell'
import AddUser from '@/components/Users/Add'
import DeleteUser from '@/components/Users/Delete'
import IconButton from '@mui/material/IconButton'
import EditIcon from '@mui/icons-material/Edit'

interface UsersDataTableProps {
  needsUpdate?: boolean
  setNeedsUpdate?: React.Dispatch<React.SetStateAction<boolean>>
}

export default function UsersDataTable(props: UsersDataTableProps) {
  const { needsUpdate, setNeedsUpdate } = props
  const [serverItemsLength, setServerItemsLength] = React.useState(0)
  const [rows, setRows] = React.useState<User[]>([])

  async function getData(props: GenericProps) {
    const { page, rows, order, direction } = props

    const orderString =
      order && direction ? `&orderBy=${order},${direction?.toUpperCase()}&` : ''

    const fetchCall = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_API_URL}/users?page=${page}&rows=${rows}${orderString}`,
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
  async function editClient() {}

  function handleEdit(event: React.MouseEvent<unknown>, id: string) {
    event.stopPropagation()
  }

  return (
    <>
      <DataTable
        serverItemsLength={serverItemsLength}
        rowCount={rows.length}
        headCells={headCells}
        needsUpdate={needsUpdate}
        setNeedsUpdate={setNeedsUpdate}
        getData={getData}
      >
        {rows.map((row, index) => {
          const labelId = `enhanced-table-checkbox-${index}`

          return (
            <TableRow
              hover
              role="checkbox"
              tabIndex={-1}
              key={row.id}
              id={row.id}
              sx={{ cursor: 'pointer' }}
            >
              <TableCell id={labelId}>{row.id}</TableCell>
              <TableCell> {row.name} </TableCell>
              <TableCell> {row.email} </TableCell>
              <TableCell> {row.roles.join(', ')} </TableCell>
              <TableCell> {row.department} </TableCell>
              <TableCell> {`${row.enabled}`} </TableCell>
              <DateCell value={row.dateCreated} />
              <DateCell value={row.dateUpdated} />
              <ActionsCell>
                <AddUser
                  setNeedsUpdate={setNeedsUpdate}
                  isEdit
                  user={row}
                  actionButton={
                    <IconButton>
                      <EditIcon />
                    </IconButton>
                  }
                />
                <DeleteUser rowId={row.id} setNeedsUpdate={setNeedsUpdate} />
              </ActionsCell>
            </TableRow>
          )
        })}
      </DataTable>
    </>
  )
}
