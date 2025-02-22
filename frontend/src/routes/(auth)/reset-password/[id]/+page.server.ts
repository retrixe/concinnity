import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Reset Password - concinnity',
    image: '/favicon.png',
    description: 'Reset your concinnity account password.',
  }
}
