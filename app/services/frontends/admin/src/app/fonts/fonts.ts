// Fonts Example
import { Roboto } from 'next/font/google'

const RobotoFont = Roboto({
  weight: ['100', '300', '400', '500', '700', '900'],
  subsets: ['latin'],
})

export default RobotoFont

// Local Fonts example
// more details here: https://nextjs.org/docs/app/building-your-application/optimizing/fonts#local-fonts
// import localFont from 'next/font/local';
// const LocalFont = localFont({src: [{path: './path-of-font-file-regular.woff', weight: '400', style: 'normal'}], fallback: ['Arial', 'sans-serif']})
// export default LocalFont;
