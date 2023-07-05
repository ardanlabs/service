import { SvgIconTypeMap } from '@mui/material'
import { OverridableComponent } from '@mui/material/OverridableComponent'

export type NavBarMenu = {
  href: string
  text: string
  icon: OverridableComponent<SvgIconTypeMap> & { muiName: string }
}
