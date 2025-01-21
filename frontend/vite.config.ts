import { sveltekit } from '@sveltejs/kit/vite'
import { defineConfig } from 'vite'
import { sveltePhosphorOptimize } from 'phosphor-svelte/vite'
import svelteSVG from '@hazycora/vite-plugin-svelte-svg'

export default defineConfig({
  plugins: [sveltePhosphorOptimize(), sveltekit(), svelteSVG()],
})
