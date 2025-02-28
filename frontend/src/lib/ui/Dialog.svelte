<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { Action } from 'svelte/action'
  import { blur } from 'svelte/transition'
  import Box from './Box.svelte'

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

  const handleClick = (event: MouseEvent) => {
    const outsideDialogBounds =
      event.target instanceof Element && !event.target.closest('.dialog-content')
    if (open && outsideDialogBounds) handleClose()
  }
</script>

<svelte:body use:scrolllock />
{#if open}
  <div transition:blur class="dialog" role="dialog" onclickcapture={handleClick}>
    <Box class="dialog-content">
      {@render children()}
    </Box>
  </div>
{/if}

<style lang="scss">
  .dialog {
    pointer-events: auto;
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 150;
    display: flex;
    justify-content: center;
    align-items: center;
    backdrop-filter: brightness(0.5);

    :global(.dialog-content) {
      display: flex;
      flex-direction: column;
      padding: 1.5rem;
      margin: 1.5rem;
      width: 100%;
      max-width: 400px;
    }
  }

  :global(.scrolllock) {
    overflow: hidden;
    pointer-events: none;
  }
</style>
