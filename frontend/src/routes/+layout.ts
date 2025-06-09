import ky from '$lib/api/ky'
import usernameCache from '$lib/state/usernameCache.svelte'
import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async event => {
  const { fetch } = event
  event.depends('app:auth')

  if (typeof localStorage !== 'object') return {}

  // Ignore errors trying to check for authentication state
  try {
    const req = await ky('', { fetch }).json<{ userId?: string; username: string; email: string }>()
    if (req.userId) usernameCache.set(req.userId, req.username)
    return { username: req.username, userId: req.userId, email: req.email }
  } catch (e) {
    console.error(e)
  }
}
