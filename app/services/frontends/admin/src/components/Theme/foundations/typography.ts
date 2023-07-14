import RobotoFont from '@/app/fonts/fonts'
import { Palette } from '@mui/material'
import { TypographyOptions } from '@mui/material/styles/createTypography'

export const typography:
  | TypographyOptions
  | ((palette: Palette) => TypographyOptions) = {
  fontFamily: RobotoFont.style.fontFamily,
  body1: { fontFamily: RobotoFont.style.fontFamily },
  body2: { fontFamily: RobotoFont.style.fontFamily },
}
