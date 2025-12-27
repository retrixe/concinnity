import { asset } from '$app/paths'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Account Settings - concinnity',
    image: asset('/favicon.png'),
    description: 'Change your account settings on concinnity.',
    noIndex: true,
  }
}
