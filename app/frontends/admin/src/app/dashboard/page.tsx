'use client'

import * as React from 'react'
import UsersDataTable from '@/components/Users/UsersDataTable'
import BaseLayout from '@/layouts/BaseLayout/BaseLayout'
import Typography from '@mui/material/Typography'
import Box from '@mui/system/Box'
import AddUser from '@/components/Users/Add'
import UserContext from '@/context/UserContext'

export default function UsersPage() {
  const [needsUpdate, setNeedsUpdate] = React.useState(false)

  return (
    <BaseLayout>
      <UserContext.Provider value={{ needsUpdate, setNeedsUpdate }}>
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'space-between',
            width: '100%',
          }}
        >
          <Typography variant="h4">{'Users'}</Typography>
          <AddUser />
        </Box>
        <UsersDataTable />
      </UserContext.Provider>
    </BaseLayout>
  )
}
