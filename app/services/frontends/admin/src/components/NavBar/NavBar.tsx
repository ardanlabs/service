'use client'

import * as React from 'react'
import Box from '@mui/material/Box'
import MuiDrawer from '@mui/material/Drawer'
import Toolbar from '@mui/material/Toolbar'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemText from '@mui/material/ListItemText'
import { NavBarProps } from '@/types/props'
import Button from '@mui/material/Button'
import { AvailableMenus } from '@/utils/access'
import { CSSObject, Theme, styled, useTheme } from '@mui/material'

const drawerWidth = 200

const openedMixin = (theme: Theme): CSSObject => ({
  width: drawerWidth,
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.enteringScreen,
  }),
  overflowX: 'hidden',
})

const closedMixin = (theme: Theme): CSSObject => ({
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  overflowX: 'hidden',
  width: `calc(${theme.spacing(7)} + 1px)`,
  [theme.breakpoints.up('sm')]: {
    width: `calc(${theme.spacing(8)} + 1px)`,
  },
})

const Drawer = styled(MuiDrawer, {
  shouldForwardProp: (prop) => prop !== 'open',
})(({ theme, open }) => ({
  width: drawerWidth,
  flexShrink: 0,
  whiteSpace: 'nowrap',
  boxSizing: 'border-box',
  ...(open && {
    ...openedMixin(theme),
    '& .MuiDrawer-paper': openedMixin(theme),
  }),
  ...(!open && {
    ...closedMixin(theme),
    '& .MuiDrawer-paper': closedMixin(theme),
  }),
}))

export default function NavBar(props: NavBarProps) {
  const theme = useTheme()
  const [drawerOpened, setDrawerOpened] = React.useState(false)

  const toggleDrawer =
    (open: boolean) => (event: React.KeyboardEvent | React.MouseEvent) => {
      setDrawerOpened(open)
    }

  return (
    <Drawer
      variant="permanent"
      onMouseOver={toggleDrawer(true)}
      onMouseOut={toggleDrawer(false)}
      open={drawerOpened}
      hideBackdrop
    >
      <Toolbar
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignContent: 'center',
          marginTop: '16px',
        }}
      >
        <Button component="a" href="#">
          <img
            src="https://www.ardanlabs.com/images/ardanlabs-logo.svg"
            alt="Ardan Labs"
          />
        </Button>
      </Toolbar>
      <Box sx={{ overflow: 'auto' }}>
        <List>
          {AvailableMenus.map((menu, index) => (
            <ListItem key={menu.text} disablePadding>
              <ListItemButton
                href={menu.href}
                sx={{
                  padding: 0,
                  overflowX: 'hidden',
                }}
              >
                <Box
                  sx={{
                    width: '24px',
                    height: '32px',
                    marginLeft: `calc(${theme.spacing(2)} + 4px)`,
                    marginRight: `calc(${theme.spacing(2)} + 4px)`,
                    alignItems: 'center',
                    display: 'flex',
                  }}
                >
                  <ListItemIcon>
                    {React.createElement(menu.icon, {
                      key: index,
                    })}
                  </ListItemIcon>
                </Box>
                <ListItemText primary={menu.text} />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      </Box>
    </Drawer>
  )
}
