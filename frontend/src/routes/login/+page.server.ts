import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Sign In - concinnity',
    image: '/favicon.png',
    description:
      'Log into or register on concinnity, a FOSS, lightweight and easy to use website to watch videos together with your friends.',
  }
}
