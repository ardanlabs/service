import { NavBarMenu } from '@/components/NavBar/types'
import PeopleIcon from '@mui/icons-material/People'

// Here you can add sidebar menus, be sure to set the page inside the app folder.
export const AvailableMenus: NavBarMenu[] = [
  { href: '/', text: 'Users', icon: PeopleIcon },
]
