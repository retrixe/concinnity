import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Verify Account - concinnity',
    image: '/favicon.png',
    description: 'Verify your newly created concinnity account!',
  }
}
