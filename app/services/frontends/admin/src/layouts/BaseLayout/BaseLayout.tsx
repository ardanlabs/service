'use client'

import * as React from 'react'
import Container from '@mui/material/Container'
import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'
import Copyright from '@/components/CopyRight/Copyright'
import NavBar from '@/components/NavBar/NavBar'
import UsersDataTable from '@/components/Users/UsersDataTable'
import ProductsDataTable from '@/components/Products/ProductsDataTable'

interface BaseLayout {
  title: string
  children?: React.ReactNode
}

export default function BaseLayout(props: BaseLayout) {
  const { title, children } = props
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
        <Typography variant="h4">{title}</Typography>
        {children}
        <Copyright />
      </Box>
    </Container>
  )
}
