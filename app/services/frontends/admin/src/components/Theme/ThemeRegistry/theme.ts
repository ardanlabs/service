import { createTheme, ThemeOptions } from '@mui/material/styles'
import { foundations } from '@/components/Theme/foundations'

// ThemeOptions are set for the MUI framework.
const themeOptions: ThemeOptions = {
  ...foundations,
}

// When needed::: first argument is needed if you have common enterprise theme, and second argument is to override your enterprise theme.
// apply fonts to all other typography options like headings, subtitles, etc...
const defaultTheme = createTheme({
  ...themeOptions,
})

export default defaultTheme
