import {createContext} from 'react'
const UserContext = createContext({} as {needsUpdate: boolean, setNeedsUpdate: React.Dispatch<React.SetStateAction<boolean>> })
export default UserContext
