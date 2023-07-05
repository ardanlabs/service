import * as React from 'react'
import ThemeRegistry from '@/components/Theme/ThemeRegistry/ThemeRegistry'

export const metadata = {
  title: 'Service Admin',
  description: 'Service Admin',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <ThemeRegistry>{children}</ThemeRegistry>
      </body>
    </html>
  )
}
