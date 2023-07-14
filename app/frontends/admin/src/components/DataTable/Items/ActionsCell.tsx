import * as React from 'react'
import TableCell from '@mui/material/TableCell'
import Box from '@mui/material/Box'

export default function ActionsCell(props: { children: React.ReactNode }) {
  const { children } = props
  return (
    <TableCell>
      <Box sx={{ display: 'flex' }}>{children}</Box>
    </TableCell>
  )
}
