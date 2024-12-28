import { PUBLIC_CONCINNITY_URL } from '$env/static/public'
import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async ({ fetch }) => {
  if (typeof localStorage !== 'object') return {}
  const token = localStorage.getItem('token')

  // Ignore errors trying to check for authentication state
  try {
    const req = await fetch(PUBLIC_CONCINNITY_URL, {
      headers: token ? { authorization: token } : {},
    })
    const data = (await req.json()) as { username?: string }
    if (req.ok) {
      return { username: data.username }
    }
    console.error(data)
  } catch (e) {
    console.error(e)
  }
}
