import { SvelteMap } from 'svelte/reactivity'
import type { UserProfile } from '$lib/api/room'

const userProfileCache = new SvelteMap<string, UserProfile | null>()

export default userProfileCache
