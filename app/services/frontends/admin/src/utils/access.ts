import { NavBarMenu } from '@/components/NavBar/types'
import InventoryIcon from '@mui/icons-material/Inventory'
import PeopleIcon from '@mui/icons-material/People'

export const AvailableMenus: NavBarMenu[] = [
  { href: '/', text: 'Users', icon: PeopleIcon },
  { href: '/products', text: 'Products', icon: InventoryIcon },
]
