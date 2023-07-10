import * as React from 'react'
import Button from '@mui/material/Button'
import Box from '@mui/material/Box'
import 'react-responsive-modal/styles.css'
import { Modal as ResponsiveModal } from 'react-responsive-modal'

interface ModalProps {
  open: boolean
  handleOpen: () => void
  handleClose: () => void
  buttonText: string
  children: React.ReactNode
}

const modalContainerStyles = {
  modal: {
    minWidth: '600px',
  },
}

export function Modal(props: ModalProps) {
  const { open, handleOpen, handleClose, buttonText, children } = props
  return (
    <>
      <Button onClick={handleOpen} variant="contained">
        {buttonText}
      </Button>

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
