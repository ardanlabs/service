import * as React from 'react'
import Button from '@mui/material/Button'
import 'react-responsive-modal/styles.css'
import { Modal as ResponsiveModal } from 'react-responsive-modal'

interface ModalProps {
  open: boolean
  handleOpen: () => void
  handleClose: () => void
  buttonText: string
  children: React.ReactNode
  actionButton?: React.ReactNode
}

const modalContainerStyles = {
  modal: {
    minWidth: '600px',
    borderRadius: '4px',
  },
}

export function Modal(props: ModalProps) {
  const { open, handleOpen, handleClose, buttonText, children, actionButton } =
    props
  return (
    <>
      {actionButton ? (
        actionButton
      ) : (
        <Button onClick={handleOpen} variant="contained">
          {buttonText}
        </Button>
      )}

      <ResponsiveModal
        styles={modalContainerStyles}
        open={open}
        onClose={handleClose}
        center
      >
        {children}
      </ResponsiveModal>
    </>
  )
}
