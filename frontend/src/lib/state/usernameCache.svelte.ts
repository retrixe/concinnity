import { SvelteMap } from 'svelte/reactivity'

const usernameCache = new SvelteMap<string, string | null>()

export default usernameCache
