/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  modularizeImports: {
    '@mui/icons-material': {
      transform: '@mui/icons-material/{{member}}',
    },
  },
  env: {
    baseAPIUrl: 'http://localhost:3000/v1',
  },
}

module.exports = nextConfig
