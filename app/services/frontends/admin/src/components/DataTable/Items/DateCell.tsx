import * as React from 'react'
import prettyDate from '@/utils/prettyDate'
import TableCell from '@mui/material/TableCell'

export default function DateCell(props: { value?: string }) {
  const formatedDate = props.value ? prettyDate(props.value) : '-'

  return <TableCell>{formatedDate}</TableCell>
}
