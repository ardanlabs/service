'use client'

import * as React from 'react'
import Typography from '@mui/material/Typography'
import Link from '@mui/material/Link'

export default function Copyright() {
  return (
    <Typography
      variant="body2"
      color="text.secondary"
      align="center"
      sx={{ alignSelf: 'center', justifySelf: 'flex-end' }}
    >
      <Link color="inherit">Ardan Labs</Link>
      {' Copyrights © '}
      {new Date().getFullYear()}.
    </Typography>
  )
}
