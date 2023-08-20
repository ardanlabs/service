/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  modularizeImports: {
    '@mui/icons-material': {
      transform: '@mui/icons-material/{{member}}',
    },
  },
  // redirects sets all redirects for the app router.
   async redirects() {
    return [
      // This points the base url "/" to the dashboard.
      {
        source: '/',
        destination: '/dashboard',
        permanent: true,
      },
    ]
  },
}

module.exports = nextConfig
