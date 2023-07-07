import { PaletteColorOptions, PaletteOptions } from '@mui/material'

declare module '@mui/material/styles' {
  export interface PaletteOptions {
    black?: PaletteColorOptions
  }
}

export const palette: PaletteOptions = {
  primary: {
    main: '#e3430e',
  },
  black: {
    main: '#151420',
  },
}
