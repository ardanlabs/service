'use client'

import { createTheme } from '@mui/material/styles'
import RobotoFont from '@/app/fonts/fonts'

// When needed::: first argument is needed if you have common enterprise theme, and second argument is to override your enterprise theme.
// apply fonts to all other typography options like headings, subtitles, etc...
const defaultTheme = createTheme({
  typography: {
    fontFamily: RobotoFont.style.fontFamily,
    body1: { fontFamily: RobotoFont.style.fontFamily },
    body2: { fontFamily: RobotoFont.style.fontFamily },
  },
})

export default defaultTheme
