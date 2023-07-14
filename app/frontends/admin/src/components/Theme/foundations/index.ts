import { palette } from './palette'
import { typography } from './typography'

// Foundations is the group of configuration present inside the ThemeOptions object.
// To add a new foundation, create a new file inside the foundation folder with
// the name of one of the ThemeOptions keys. After that, export a constant of the
// specific type of that key. Inside that constant you can populate with the available
// settings.
// Your last step is to import the file here and add it to the object below.
export const foundations = { palette, typography }
