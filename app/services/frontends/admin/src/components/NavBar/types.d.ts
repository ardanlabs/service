import { SvgIconTypeMap } from '@mui/material'
import { OverridableComponent } from '@mui/material/OverridableComponent'

export interface NavBarProps {}

export interface NavBarMenu {
  href: string
  text: string
  icon: OverridableComponent<SvgIconTypeMap> & { muiName: string }
}
