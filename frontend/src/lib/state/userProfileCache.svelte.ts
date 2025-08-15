import { SvelteMap } from 'svelte/reactivity'

export interface UserProfile {
  username: string
  avatar: string | null
}

const userProfileCache = new SvelteMap<string, UserProfile | null>()

export default userProfileCache
