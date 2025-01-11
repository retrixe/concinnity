import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'concinnity',
    image: 'https://f002.backblazeb2.com/file/mythic-storage-public/demo-dark.webp',
    description:
      'Watch videos together with your friends using concinnity, a FOSS, lightweight and easy to use website built by a developer looking for something better.',
  }
}
