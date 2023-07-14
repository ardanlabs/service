import * as React from 'react'
import { Modal } from '../Modal/Modal'
import ApiError from '../ApiError/ApiError'
import Box from '@mui/system/Box'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'

interface ConfirmationModalProps {
  handleOpenModal: () => void
  handleCloseModal: () => void
  open: boolean
  actionButton?: React.ReactNode
  buttonText: string
  error: string
  clearError: () => void
  confirmAction: React.MouseEventHandler<HTMLButtonElement> | undefined
  title: string
  subtitle: string
  cancelButtonText: string
  confirmButtonText: string
}

export default function ConfirmationModal(props: ConfirmationModalProps) {
  const {
    handleOpenModal,
    handleCloseModal,
    open,
    actionButton,
    buttonText,
    error,
    clearError,
    confirmAction,
    title,
    subtitle,
    cancelButtonText,
    confirmButtonText,
  } = props

  return (
    <Modal
      buttonText={buttonText}
      handleOpen={handleOpenModal}
      handleClose={handleCloseModal}
      open={open}
      actionButton={actionButton}
    >
      {error ? (
        <ApiError message={error} clearError={() => clearError} />
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
            {title}
          </Typography>
          <Typography
            id="modal-modal-title"
            sx={{ my: 4 }}
            variant="h6"
            component="h6"
          >
            {subtitle}
          </Typography>
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <Button sx={{ m: 2 }} size="large" onClick={handleCloseModal}>
              {cancelButtonText}
            </Button>
            <Button sx={{ m: 2 }} size="large" onClick={confirmAction}>
              {confirmButtonText}
            </Button>
          </Box>
        </Box>
      )}
    </Modal>
  )
}
