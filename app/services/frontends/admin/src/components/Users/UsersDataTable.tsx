'use client'

import * as React from 'react'
import DataTable from '@/components/DataTable/DataTable'
import TableRow from '@mui/material/TableRow'
import TableCell from '@mui/material/TableCell'
import { DateCell } from '@/components/DataTable/Items/DateCell'
import { DefaultAPIResponse } from '@/utils/types'
import { User, headCells } from './constants'
import { GenericProps } from '../DataTable/types'
import IconButton from '@mui/material/IconButton'
import DeleteIcon from '@mui/icons-material/Delete'
import EditIcon from '@mui/icons-material/Edit'
import ActionsCell from '@/components/DataTable/Items/ActionsCell'
import { Modal } from '../Modal/Modal'
import ApiError from '../ApiError/ApiError'
import Box from '@mui/system/Box'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'

interface UsersDataTableProps {
  needsUpdate?: boolean
  setNeedsUpdate?: React.Dispatch<React.SetStateAction<boolean>>
}

export default function UsersDataTable(props: UsersDataTableProps) {
  const { needsUpdate, setNeedsUpdate } = props
  const [serverItemsLength, setServerItemsLength] = React.useState(0)
  const [rows, setRows] = React.useState<User[]>([])
  const [open, setOpen] = React.useState(false)
  const handleOpen = () => setOpen(true)
  const handleClose = () => setOpen(false)
  const [rowDelete, setRowDelete] = React.useState('')
  const [fetchError, setFetchError] = React.useState('')

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

  function handleEdit(event: React.MouseEvent<unknown>, id: string) {
    event.stopPropagation()
  }

  async function deleteUser() {
    if (!rowDelete) return

    handleClose()
    if (setNeedsUpdate) {
      handleClose()
      setNeedsUpdate(true)
    }

    const userDelete = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_API_URL}/users/${rowDelete}`,
      {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${process.env.NEXT_PUBLIC_TOKEN}`,
        },
      },
    )

    if (userDelete.ok) {
      setRowDelete('')
      if (setNeedsUpdate) {
        handleClose()
        setNeedsUpdate(true)
      }
      return
    }

    const error: { error: string } = await userDelete.json()

    setFetchError(error.error)
    setRowDelete('')
  }

  function handleDelete(event: React.MouseEvent<unknown>, id: string): void {
    event.stopPropagation()
    setRowDelete(id)

    setOpen(true)
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
                <IconButton onClick={(event) => handleEdit(event, row.id)}>
                  <EditIcon />
                </IconButton>
                <Modal
                  buttonText="Delete User"
                  handleOpen={handleOpen}
                  handleClose={handleClose}
                  open={open}
                  actionButton={
                    <IconButton
                      onClick={(event) => handleDelete(event, row.id)}
                    >
                      <DeleteIcon />
                    </IconButton>
                  }
                >
                  {fetchError ? (
                    <ApiError
                      message={fetchError}
                      clearError={() => setFetchError('')}
                    />
                  ) : (
                    <Box
                      sx={{
                        display: 'flex',
                        alignContent: 'center',
                        justifyContent: 'center',
                        flexDirection: 'column',
                        textAlign: 'center',
                      }}
                    >
                      <Typography
                        id="modal-modal-title"
                        sx={{ fontWeight: 500, my: 4 }}
                        variant="h3"
                        component="h3"
                      >
                        DELETE USER
                      </Typography>
                      <Typography
                        id="modal-modal-title"
                        sx={{ my: 4 }}
                        variant="h6"
                        component="h6"
                      >
                        Are you sure you want to delete this user?
                      </Typography>
                      <Box
                        sx={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                        }}
                      >
                        <Button
                          sx={{ m: 2 }}
                          size="large"
                          onClick={handleClose}
                        >
                          Cancel
                        </Button>
                        <Button sx={{ m: 2 }} size="large" onClick={deleteUser}>
                          Confirm
                        </Button>
                      </Box>
                    </Box>
                  )}
                </Modal>
              </ActionsCell>
            </TableRow>
          )
        })}
      </DataTable>
    </>
  )
}
