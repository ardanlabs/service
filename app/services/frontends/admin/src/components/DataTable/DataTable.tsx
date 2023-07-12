import * as React from 'react'
import Table from '@mui/material/Table'
import Box from '@mui/material/Box'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableContainer from '@mui/material/TableContainer'
import TablePagination from '@mui/material/TablePagination'
import TableRow from '@mui/material/TableRow'
import Paper from '@mui/material/Paper'
import { DataTableProps, Order } from './types'
import EnhancedTableHead from './EnhancedTableHead'

type PropsWithChildren<P> = P & { children?: React.ReactNode }

export default function DataTable(props: PropsWithChildren<DataTableProps>) {
  // Extract props.
  const {
    headCells,
    dense,
    defaultOrder,
    defaultOrderDirection,
    rowsPerPageProp,
    pageProp,
    getData,
    rowsPerPageOptions,
    selectable,
    selectedCount,
    rowCount,
    serverItemsLength,
    handleSelectAllClick,
    children,
    needsUpdate,
    setNeedsUpdate,
  } = props
  // Set init states.
  const [orderDirection, setOrderDirection] = React.useState<Order>(
    defaultOrderDirection || 'asc',
  )

  const [orderBy, setOrderBy] = React.useState(defaultOrder || '')
  const [page, setPage] = React.useState(pageProp || 0)
  const [rowsPerPage, setRowsPerPage] = React.useState(rowsPerPageProp || 10)

  // handleSelectAllClick handles the sorting changes
  const handleRequestSort = (
    event: React.MouseEvent<unknown>,
    property: string,
  ) => {
    const isAsc = orderBy === property && orderDirection === 'asc'
    const isDesc = orderBy === property && orderDirection === 'desc'
    setOrderBy(property)
    if (isAsc) {
      setOrderDirection('desc')
      return
    }
    if (isDesc) {
      setOrderDirection(undefined)
      setOrderBy('')
      return
    }
    setOrderDirection('asc')
  }

  // handleChangePage handles page selection.
  const handleChangePage = (event: unknown, newPage: number) => {
    if (newPage) setPage(newPage)
  }

  // handleChangeRowsPerPage handles rowsPerPage selection.
  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    setRowsPerPage(parseInt(event.target.value, 10))
    setPage(0)
  }

  // This effect gets the data from the API everytime any of the order or pages changes
  React.useEffect(() => {
    getData({
      page: page + 1,
      rows: rowsPerPage,
      order: orderBy,
      direction: orderDirection,
    })

    if (setNeedsUpdate) {
      setNeedsUpdate(false)
    }
  }, [page, rowsPerPage, orderBy, orderDirection, needsUpdate])

  return (
    <Box sx={{ width: '100%', my: 2 }}>
      <Paper sx={{ width: '100%', mb: 2 }}>
        <TableContainer>
          <Table
            sx={{ minWidth: 750 }}
            aria-labelledby="tableTitle"
            size={dense ? 'small' : 'medium'}
            className="data-table"
          >
            <EnhancedTableHead
              numSelected={selectedCount || 0}
              order={orderDirection}
              orderBy={orderBy}
              onSelectAllClick={handleSelectAllClick}
              onRequestSort={handleRequestSort}
              rowCount={rowCount}
              headCells={headCells}
              selectEnabled={Boolean(selectable)}
            />
            <TableBody>
              {children ? (
                children
              ) : (
                <TableRow
                  style={{
                    height: dense ? 33 : 53,
                  }}
                >
                  <TableCell colSpan={6} />
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          rowsPerPageOptions={rowsPerPageOptions || [10, 20, 50]}
          component="div"
          count={serverItemsLength}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Paper>
    </Box>
  )
}
