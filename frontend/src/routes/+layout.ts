import ky from '$lib/api/ky'
import userProfileCache from '$lib/state/userProfileCache.svelte'
import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async event => {
  const { fetch } = event
  event.depends('app:auth')

  if (typeof localStorage !== 'object') return {}

  // Ignore errors trying to check for authentication state
  try {
    const req = await ky('', { fetch }).json<{
      userId?: string
      username: string
      email: string
      avatar: string | null
    }>()
    if (req.userId) userProfileCache.set(req.userId, { username: req.username, avatar: req.avatar })
    return { username: req.username, userId: req.userId, email: req.email, avatar: req.avatar }
  } catch (e) {
    console.error(e)
  }
}
