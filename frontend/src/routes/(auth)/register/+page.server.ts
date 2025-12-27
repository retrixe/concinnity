import { asset } from '$app/paths'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Register - concinnity',
    image: asset('/favicon.png'),
    description:
      'Register an account on concinnity, a FOSS, lightweight and easy to use website to watch videos together with your friends.',
  }
}
