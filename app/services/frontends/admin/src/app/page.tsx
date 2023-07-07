'use client'

import ProductsDataTable from '@/components/Products/ProductsDataTable'
import UsersDataTable from '@/components/Users/UsersDataTable'
import BaseLayout from '@/layouts/BaseLayout/BaseLayout'
import { usePathname } from 'next/navigation'
import * as React from 'react'

interface PagesComponentInterface {
  [key: string]: React.ReactNode
}

const PagesComponent: PagesComponentInterface = {
  '/': <UsersDataTable />,
  '/products': <ProductsDataTable />,
}

export default function RootPage() {
  const pathName = usePathname()
  const [slug, setSlug] = React.useState<keyof PagesComponentInterface>('/')

  React.useEffect(() => {
    setSlug(pathName)
  }, [pathName])
  return <BaseLayout title="Users">{PagesComponent[slug]}</BaseLayout>
}
