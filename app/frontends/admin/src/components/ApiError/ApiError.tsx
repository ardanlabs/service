import * as React from 'react'
import ErrorOutline from '@mui/icons-material/ErrorOutline'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Typography from '@mui/material/Typography'

interface ApiErrorProps {
  message: string
  clearError: () => void
}

function ApiError(props: ApiErrorProps) {
  const { message, clearError } = props

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexDirection: 'column',
        textAlign: 'center',
      }}
    >
      <ErrorOutline sx={{ color: 'red', my: 4, fontSize: '15em' }} />
      <Typography
        id="modal-modal-title"
        sx={{ color: 'red', fontWeight: 500, mb: 4 }}
        variant="h2"
        component="h2"
      >
        ERROR
      </Typography>
      <Typography
        id="modal-modal-title"
        sx={{ color: 'darkRed' }}
        variant="h6"
        component="h6"
      >
        {message}
      </Typography>
      <Button sx={{ m: 2 }} size="large" onClick={clearError}>
        Try Again
      </Button>
    </Box>
  )
}
export default ApiError
