'use client'

import * as React from 'react'
import Container from '@mui/material/Container'
import Box from '@mui/material/Box'
import Copyright from '@/components/CopyRight/Copyright'
import NavBar from '@/components/NavBar/NavBar'

interface BaseLayout {
  children?: React.ReactNode
}

export default function BaseLayout(props: BaseLayout) {
  const { children } = props
  return (
    <Container maxWidth="xl" disableGutters sx={{ height: '100%' }}>
      <NavBar />
      <Box
        sx={{
          my: 4,
          mx: 10,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'start',
          alignItems: 'start',
        }}
      >
        {children}
        <Copyright />
      </Box>
    </Container>
  )
}
