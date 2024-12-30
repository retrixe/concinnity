import { PUBLIC_BACKEND_URL } from '$env/static/public'
import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async event => {
  const { fetch } = event
  event.depends('app:auth')

  if (typeof localStorage !== 'object') return {}
  const token = localStorage.getItem('concinnity:token') ?? ''

  // Ignore errors trying to check for authentication state
  try {
    const req = await fetch(PUBLIC_BACKEND_URL, { headers: { authorization: token } })
    const data = (await req.json()) as { username?: string; error?: string }
    if (req.ok) {
      return { username: data.username }
    }
    console.error('Failed to check for auth!', data.error ?? req.statusText)
  } catch (e) {
    console.error(e)
  }
}
