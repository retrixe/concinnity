import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Login - concinnity',
    image: '/favicon.png',
    description:
      'Log into your account on concinnity, a FOSS, lightweight and easy to use website to watch videos together with your friends.',
  }
}
