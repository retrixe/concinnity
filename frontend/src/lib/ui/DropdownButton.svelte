<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'
  import Button from './Button.svelte'

  const {
    primaryChild,
    secondaryChild,
    children,
    primary = {},
    secondary = {},
  }: {
    primaryChild: Snippet
    secondaryChild: Snippet
    children?: Snippet
    primary?: HTMLButtonAttributes
    secondary?: HTMLButtonAttributes
  } = $props()
</script>

<div class="dropdown-button">
  <Button {...primary}>{@render primaryChild()}</Button>
  <Button {...secondary}>{@render secondaryChild()}</Button>
  {@render children?.()}
</div>

<style lang="scss">
  .dropdown-button {
    position: relative;
    display: flex;
    gap: 2px;
    > :global(button:first-child) {
      border-radius: 0.6rem 0 0 0.6rem;
    }
    > :global(button:nth-child(2)) {
      border-radius: 0 0.6rem 0.6rem 0;
      padding: 0.6rem;
    }
  }
</style>
