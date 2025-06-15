<script lang="ts">
  import type { Snippet } from 'svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { Box } from 'heliodor'

  const { children }: { children: Snippet } = $props()
  const { username } = $derived(page.data)
  $effect(() => {
    if (username) goto('/').catch(console.error)
  })
</script>

<div class="container">
  <Box>
    {@render children()}
  </Box>
</div>

<style lang="scss">
  .container > :global(div) {
    display: flex;
    flex-direction: column;
    padding: 1.5rem;
    margin: 1.5rem;
    width: 100%;
    max-width: 25rem;

    :global(.error) {
      color: var(--error-color);
    }

    :global(label) {
      padding: 0.5rem 0rem;
      font-weight: bold;
    }

    :global(.center) {
      align-self: center;
      text-align: center;
    }

    :global(.result) {
      align-self: center;
    }

    :global(.spacer) {
      margin-top: 1rem;
    }
  }

  .container {
    flex-grow: 1;
    display: flex;
    justify-content: center;
    align-items: center;
  }
</style>
