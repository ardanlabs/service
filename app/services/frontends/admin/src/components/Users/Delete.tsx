import * as React from 'react'
import { Modal } from '../Modal/Modal'
import ApiError from '../ApiError/ApiError'
import Box from '@mui/system/Box'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'
import IconButton from '@mui/material/IconButton'
import DeleteIcon from '@mui/icons-material/Delete'

interface DeleteUserProps {
  rowId: string
  setNeedsUpdate?: React.Dispatch<React.SetStateAction<boolean>>
}

export default function DeleteUser(props: DeleteUserProps) {
  const { rowId, setNeedsUpdate } = props
  const [fetchError, setFetchError] = React.useState('')
  const [open, setOpenDelete] = React.useState(false)
  const handleOpenDelete = () => setOpenDelete(true)
  const handleCloseDelete = () => setOpenDelete(false)

  async function deleteUser() {
    if (!rowId) return

    if (setNeedsUpdate) {
      setNeedsUpdate(true)
    }

    const userDelete = await fetch(
      `${process.env.NEXT_PUBLIC_BASE_API_URL}/users/${rowId}`,
      {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${process.env.NEXT_PUBLIC_TOKEN}`,
        },
      },
    )

    if (userDelete.ok) {
      if (setNeedsUpdate) {
        setNeedsUpdate(true)
      }
      return
    }

    const error: { error: string } = await userDelete.json()

    setFetchError(error.error)
    handleCloseDelete()
  }

  function handleDelete(event: React.MouseEvent<unknown>): void {
    event.stopPropagation()

    setOpenDelete(true)
  }

  return (
    <Modal
      buttonText="Delete User"
      handleOpen={handleOpenDelete}
      handleClose={handleCloseDelete}
      open={open}
      actionButton={
        <IconButton onClick={handleDelete}>
          <DeleteIcon />
        </IconButton>
      }
    >
      {fetchError ? (
        <ApiError message={fetchError} clearError={() => setFetchError('')} />
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
            <Button sx={{ m: 2 }} size="large" onClick={handleCloseDelete}>
              Cancel
            </Button>
            <Button sx={{ m: 2 }} size="large" onClick={deleteUser}>
              Confirm
            </Button>
          </Box>
        </Box>
      )}
    </Modal>
  )
}
