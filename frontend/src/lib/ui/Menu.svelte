<script lang="ts">
  import type { Snippet } from 'svelte'
  import { blur } from 'svelte/transition'
  import Box from './Box.svelte'
  import type { Action } from 'svelte/action'

  const {
    open,
    onClose: handleClose,
    children,
  }: { open: boolean; onClose: () => void; children: Snippet } = $props()

  const scrolllock: Action<HTMLBodyElement> = node => {
    $effect(() => {
      node.classList.toggle('scrolllock', open)
      return () => node.classList.remove('scrolllock')
    })
  }

  const handleWindowClick = (event: MouseEvent) => {
    const outsideMenuBounds = event.target instanceof Element && !event.target.closest('.menu')
    if (open && outsideMenuBounds) handleClose()
  }
</script>

<svelte:body use:scrolllock />
<svelte:window onclickcapture={handleWindowClick} />
{#if open}
  <div transition:blur class="menu">
    <Box>
      {@render children()}
    </Box>
  </div>
{/if}

<style lang="scss">
  .menu {
    position: absolute;
    top: calc(100% - 4px);
    right: 0;
    min-width: 12rem;
    z-index: 50;
    pointer-events: auto;
    > :global(div) {
      display: flex;
      flex-direction: column;
      padding: 4px;
    }
  }

  :global(.scrolllock) {
    overflow: hidden;
    pointer-events: none;
  }
</style>
