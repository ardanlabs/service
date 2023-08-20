'use client'

import * as React from 'react'
import TableRow from '@mui/material/TableRow'
import TableCell from '@mui/material/TableCell'
import IconButton from '@mui/material/IconButton'
import EditIcon from '@mui/icons-material/Edit'

import DataTable from '@/components/DataTable/DataTable'
import DateCell from '@/components/DataTable/Items/DateCell'
import ActionsCell from '@/components/DataTable/Items/ActionsCell'
import parseOrderBy from '@/components/DataTable/parseOrderBy'
import { GenericProps } from '@/components/DataTable/types'

import AddUser from '@/components/Users/Add'
import DeleteUser from '@/components/Users/Delete'
import { DefaultAPIResponse } from '@/utils/types'
import { User, headCells } from '@/components/Users/constants'

export default function UsersDataTable() {
  const [serverItemsLength, setServerItemsLength] = React.useState(0)
  const [rows, setRows] = React.useState<User[]>([])

  async function getData(props: GenericProps) {
    const { page, rows, order, direction } = props

    const orderString = parseOrderBy(order, direction)

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

  return (
    <>
      <DataTable
        serverItemsLength={serverItemsLength}
        rowCount={rows.length}
        headCells={headCells}
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
                  isEdit
                  user={row}
                  actionButton={
                    <IconButton>
                      <EditIcon />
                    </IconButton>
                  }
                />
                <DeleteUser rowId={row.id} />
              </ActionsCell>
            </TableRow>
          )
        })}
      </DataTable>
    </>
  )
}
