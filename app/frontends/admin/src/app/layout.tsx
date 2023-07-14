import * as React from 'react'
import ThemeRegistry from '@/components/Theme/ThemeRegistry/ThemeRegistry'

export const metadata = {
  title: 'Service Admin',
  description: 'Service Admin',
}

// This is the entry layout of the application.
// ThemeRegistry make MUI available through the whole app.
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
