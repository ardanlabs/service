import * as React from 'react'
import IconButton from '@mui/material/IconButton'
import DeleteIcon from '@mui/icons-material/Delete'
import ConfirmationModal from '../ConfirmationModal/ConfirmationModal'

interface DeleteUserProps {
  rowId: string
  setNeedsUpdate?: React.Dispatch<React.SetStateAction<boolean>>
}

// Delete user displays a confirmation modal that lets you delete a user.
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
    <ConfirmationModal
      buttonText="Delete User"
      handleOpenModal={handleOpenDelete}
      handleCloseModal={handleCloseDelete}
      open={open}
      actionButton={
        <IconButton onClick={handleDelete}>
          <DeleteIcon />
        </IconButton>
      }
      error={fetchError}
      clearError={() => setFetchError('')}
      confirmAction={deleteUser}
      title="DELETE USER"
      subtitle="Are you sure you want to delete this user?"
      cancelButtonText="Cancel"
      confirmButtonText="Confirm"
    />
  )
}
