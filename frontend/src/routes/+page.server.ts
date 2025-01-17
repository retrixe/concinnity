import type { PageServerLoad } from './$types'

interface PageData {
  title: string
  image?: string
  imageLarge?: boolean
  description: string
  noIndex?: boolean
}

export const load: PageServerLoad<PageData> = () => {
  return {
    title: 'concinnity',
    image: 'https://f002.backblazeb2.com/file/retrixe-storage-public/concinnity/demo-dark.png',
    imageLarge: true,
    description:
      'Watch videos together with your friends using concinnity, a FOSS, lightweight and easy to use website.',
  }
}
