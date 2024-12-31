import type { PageServerLoad } from './$types'

export const load: PageServerLoad = () => {
  return {
    title: 'Watch Together Room - concinnity',
    image: '/favicon.png',
    description: "You're invited to watch a video together with your friends on concinnity!",
  }
}
