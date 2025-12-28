import { sveltekit } from '@sveltejs/kit/vite'
import { defineConfig } from 'vite'
import devtoolsJson from 'vite-plugin-devtools-json'
import { sveltePhosphorOptimize } from 'phosphor-svelte/vite'

export default defineConfig({
  plugins: [sveltekit(), sveltePhosphorOptimize(), devtoolsJson()],
})
