import { assets } from '$app/paths'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Reset Password - concinnity',
    image: assets + '/favicon.png',
    description: 'Reset your concinnity account password.',
  }
}
