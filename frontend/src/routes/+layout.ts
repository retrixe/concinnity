import { PUBLIC_BACKEND_URL } from '$env/static/public'
import usernameCache from '$lib/state/usernameCache.svelte'
import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async event => {
  const { fetch } = event
  event.depends('app:auth')

  if (typeof localStorage !== 'object') return {}
  const token = localStorage.getItem('concinnity:token') ?? ''

  // Ignore errors trying to check for authentication state
  try {
    const req = await fetch(PUBLIC_BACKEND_URL, { headers: { authorization: token } })
    const data = (await req.json()) as { username?: string; userId?: string; error?: string }
    if (req.ok) {
      if (data.userId && data.username) usernameCache.set(data.userId, data.username)
      return { username: data.username, userId: data.userId }
    }
    console.error('Failed to check for auth!', data.error ?? req.statusText)
  } catch (e) {
    console.error(e)
  }
}
